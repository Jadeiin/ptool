package unit3d

// UNIT3D ( https://github.com/HDInnovations/UNIT3D-Community-Edition )
// JptvClub、莫妮卡、普斯特等站使用架构
// 种子下载链接格式：https://jptv.club/torrents/download/39683

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Noooste/azuretls-client"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/sagan/ptool/config"
	"github.com/sagan/ptool/site"
	"github.com/sagan/ptool/util"
)

type Site struct {
	Name                 string
	Location             *time.Location
	SiteConfig           *config.SiteConfigStruct
	Config               *config.ConfigStruct
	HttpClient           *azuretls.Session
	HttpHeaders          [][]string
	latestTorrents       []*site.Torrent
	datatime             int64
	torrentsParserOption *TorrentsParserOption
}

// PublishTorrent implements site.Site.
func (usite *Site) PublishTorrent(contents []byte, metadata url.Values) (id string, err error) {
	return "", site.ErrUnimplemented
}

const (
	DEFAULT_TORRENTS_URL       = "torrents"
	SELECTOR_USERNAME          = ".top-nav__username"
	SELECTOR_USER_UPLOADED     = ".ratio-bar__uploaded"
	SELECTOR_USER_DOWNLOADED   = ".ratio-bar__downloaded"
	SELECTOR_USER_RATIO        = ".ratio-bar__ratio"
	SELECTOR_USER_BONUS        = ".ratio-bar__points"
	SELECTOR_USER_SEEDING      = ".ratio-bar__seeding"
	SELECTOR_USER_LEECHING     = ".ratio-bar__leeching"
	SELECTOR_USER_SEEDING_SIZE = ".ratio-bar__seeding-size"
)

func (usite *Site) GetDefaultHttpHeaders() [][]string {
	return usite.HttpHeaders
}

func (usite *Site) PurgeCache() {
}

func (usite *Site) GetName() string {
	return usite.Name
}

func (usite *Site) GetSiteConfig() *config.SiteConfigStruct {
	return usite.SiteConfig
}

func (usite *Site) GetStatus() (*site.Status, error) {
	doc, res, err := util.GetUrlDocWithAzuretls(usite.SiteConfig.Url+"torrents", usite.HttpClient,
		usite.GetSiteConfig().Cookie, site.GetUa(usite), usite.GetDefaultHttpHeaders())
	if err != nil {
		return nil, err
	}
	if strings.Contains(res.Request.Url, "/login") {
		return nil, fmt.Errorf("not logined (cookie may has expired)")
	}

	// 使用配置的选择器或默认选择器
	userNameSelector := SELECTOR_USERNAME
	userUploadedSelector := SELECTOR_USER_UPLOADED
	userDownloadedSelector := SELECTOR_USER_DOWNLOADED
	if usite.SiteConfig.SelectorUserInfoUserName != "" {
		userNameSelector = usite.SiteConfig.SelectorUserInfoUserName
	}
	if usite.SiteConfig.SelectorUserInfoUploaded != "" {
		userUploadedSelector = usite.SiteConfig.SelectorUserInfoUploaded
	}
	if usite.SiteConfig.SelectorUserInfoDownloaded != "" {
		userDownloadedSelector = usite.SiteConfig.SelectorUserInfoDownloaded
	}

	// 获取基本用户信息
	usernameEl := doc.Find(userNameSelector)
	uploadedEl := doc.Find(userUploadedSelector)
	downloadedEl := doc.Find(userDownloadedSelector)

	userName := util.DomSanitizedText(usernameEl)
	userUploaded, _ := util.ExtractSizeStr(util.DomSanitizedText(uploadedEl))
	userDownloaded, _ := util.ExtractSizeStr(util.DomSanitizedText(downloadedEl))

	// 尝试获取做种/下载数量
	var seedingCnt, leechingCnt int64
	seedingEl := doc.Find(SELECTOR_USER_SEEDING)
	if seedingEl.Length() > 0 {
		seedingCnt = util.ParseInt(util.DomSanitizedText(seedingEl))
	}
	leechingEl := doc.Find(SELECTOR_USER_LEECHING)
	if leechingEl.Length() > 0 {
		leechingCnt = util.ParseInt(util.DomSanitizedText(leechingEl))
	}

	// 记录额外的用户信息到日志
	ratioEl := doc.Find(SELECTOR_USER_RATIO)
	if ratioEl.Length() > 0 {
		ratioText := util.DomSanitizedText(ratioEl)
		log.Tracef("Site %s user ratio: %s", usite.Name, ratioText)
	}

	bonusEl := doc.Find(SELECTOR_USER_BONUS)
	if bonusEl.Length() > 0 {
		bonusText := util.DomSanitizedText(bonusEl)
		log.Tracef("Site %s user bonus: %s", usite.Name, bonusText)
	}

	return &site.Status{
		UserName:            userName,
		UserUploaded:        userUploaded,
		UserDownloaded:      userDownloaded,
		TorrentsSeedingCnt:  seedingCnt,
		TorrentsLeechingCnt: leechingCnt,
	}, nil
}

func (usite *Site) GetAllTorrents(sort string, desc bool, pageMarker string, baseUrl string) (
	torrents []*site.Torrent, nextPageMarker string, err error) {

	page := 1
	if pageMarker != "" {
		if parsedPage := util.ParseInt(pageMarker); parsedPage > 0 {
			page = int(parsedPage)
		}
	}

	url := usite.SiteConfig.TorrentsUrl
	if url == "" {
		url = DEFAULT_TORRENTS_URL
	}
	if baseUrl != "" {
		url = baseUrl
	}

	// 构建分页URL
	pageUrl := url
	if strings.Contains(url, "?") {
		pageUrl = fmt.Sprintf("%s&page=%d", url, page)
	} else {
		pageUrl = fmt.Sprintf("%s?page=%d", url, page)
	}

	// 添加排序参数
	if sort != "" {
		sortParam := sort
		if desc {
			sortParam = "-" + sortParam
		}
		pageUrl = fmt.Sprintf("%s&sort=%s", pageUrl, sortParam)
	}

	pageUrl = usite.SiteConfig.ParseSiteUrl(pageUrl, false)

	doc, res, err := util.GetUrlDocWithAzuretls(pageUrl, usite.HttpClient,
		usite.SiteConfig.Cookie, site.GetUa(usite), usite.GetDefaultHttpHeaders())
	if err != nil {
		return nil, "", fmt.Errorf("failed to get torrents page: %w", err)
	}
	if strings.Contains(res.Request.Url, "/login") {
		return nil, "", fmt.Errorf("not logined (cookie may has expired)")
	}

	doctime := util.Now()
	torrents, err = parseTorrents(doc, usite.torrentsParserOption, doctime, usite.GetName())
	if err != nil {
		return nil, "", err
	}

	// 检查是否有下一页
	hasNextPage := false
	doc.Find("ul.pagination li a[href*='page=']").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if strings.Contains(href, fmt.Sprintf("page=%d", page+1)) {
			hasNextPage = true
		}
	})

	// 如果当前页有结果且可能有下一页，返回下一页标记
	if len(torrents) > 0 && hasNextPage {
		nextPageMarker = fmt.Sprintf("%d", page+1)
	}

	return torrents, nextPageMarker, nil
}

func (usite *Site) GetLatestTorrents(full bool) ([]*site.Torrent, error) {
	if usite.datatime > 0 && !full {
		return usite.latestTorrents, nil
	}

	url := usite.SiteConfig.TorrentsUrl
	if url == "" {
		url = DEFAULT_TORRENTS_URL
	}
	url = usite.SiteConfig.ParseSiteUrl(url, false)

	doc, res, err := util.GetUrlDocWithAzuretls(url, usite.HttpClient,
		usite.SiteConfig.Cookie, site.GetUa(usite), usite.GetDefaultHttpHeaders())
	if !usite.SiteConfig.AcceptAnyHttpStatus && err != nil || doc == nil {
		return nil, fmt.Errorf("failed to get site page dom: %w", err)
	}
	if strings.Contains(res.Request.Url, "/login") {
		return nil, fmt.Errorf("not logined (cookie may has expired)")
	}

	usite.datatime = util.Now()
	torrents, err := parseTorrents(doc, usite.torrentsParserOption, usite.datatime, usite.GetName())
	if err != nil {
		log.Errorf("failed to parse site page torrents: %v", err)
		return nil, err
	}

	usite.latestTorrents = torrents
	return torrents, nil
}

func (usite *Site) SearchTorrents(keyword string, baseUrl string) ([]*site.Torrent, error) {
	if keyword == "" {
		return nil, fmt.Errorf("keyword is empty")
	}

	// Unit3D 搜索URL格式: /torrents?name=keyword
	searchUrl := "torrents"
	if baseUrl != "" {
		searchUrl = baseUrl
	}

	// 构建搜索URL
	if strings.Contains(searchUrl, "?") {
		searchUrl = fmt.Sprintf("%s&name=%s", searchUrl, url.QueryEscape(keyword))
	} else {
		searchUrl = fmt.Sprintf("%s?name=%s", searchUrl, url.QueryEscape(keyword))
	}

	searchUrl = usite.SiteConfig.ParseSiteUrl(searchUrl, false)

	doc, res, err := util.GetUrlDocWithAzuretls(searchUrl, usite.HttpClient,
		usite.SiteConfig.Cookie, site.GetUa(usite), usite.GetDefaultHttpHeaders())
	if err != nil {
		return nil, fmt.Errorf("failed to search torrents: %w", err)
	}
	if strings.Contains(res.Request.Url, "/login") {
		return nil, fmt.Errorf("not logined (cookie may has expired)")
	}

	doctime := util.Now()
	torrents, err := parseTorrents(doc, usite.torrentsParserOption, doctime, usite.GetName())
	if err != nil {
		return nil, err
	}

	return torrents, nil
}

func (usite *Site) DownloadTorrent(torrentUrl string) (content []byte, filename string, id string, err error) {
	if !util.IsUrl(torrentUrl) {
		id = strings.TrimPrefix(torrentUrl, usite.GetName()+".")
		content, filename, err = usite.DownloadTorrentById(id)
		return
	}
	if !strings.Contains(torrentUrl, "/torrents/download/") {
		idRegexp := regexp.MustCompile(`torrents/(?P<id>\d+)\b`)
		if m := idRegexp.FindStringSubmatch(torrentUrl); m != nil {
			id = m[idRegexp.SubexpIndex("id")]
			content, filename, err = usite.DownloadTorrentById(id)
			return
		}
	}
	idRegexp := regexp.MustCompile(`/download/(?P<id>\d+)\b`)
	if m := idRegexp.FindStringSubmatch(torrentUrl); m != nil {
		id = m[idRegexp.SubexpIndex("id")]
	}
	content, filename, err = site.DownloadTorrentByUrl(usite, usite.HttpClient, torrentUrl, id)
	return
}

func (usite *Site) DownloadTorrentById(id string) ([]byte, string, error) {
	torrentUrl := usite.SiteConfig.Url + "torrents/download/" + id
	return site.DownloadTorrentByUrl(usite, usite.HttpClient, torrentUrl, id)
}

func NewSite(name string, siteConfig *config.SiteConfigStruct, config *config.ConfigStruct) (site.Site, error) {
	if siteConfig.Cookie == "" {
		log.Warnf("Site %s has no cookie provided", name)
	}
	location, err := time.LoadLocation(siteConfig.GetTimezone())
	if err != nil {
		return nil, fmt.Errorf("invalid site timezone %s: %w", siteConfig.GetTimezone(), err)
	}
	httpClient, httpHeaders, err := site.CreateSiteHttpClient(siteConfig, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create site http client: %w", err)
	}
	site := &Site{
		Name:        name,
		Location:    location,
		SiteConfig:  siteConfig,
		Config:      config,
		HttpClient:  httpClient,
		HttpHeaders: httpHeaders,
		torrentsParserOption: &TorrentsParserOption{
			location:                    location,
			siteurl:                     siteConfig.Url,
			selectorTorrentsList:        siteConfig.SelectorTorrentsList,
			selectorTorrentBlock:        siteConfig.SelectorTorrentBlock,
			selectorTorrentDownloadLink: siteConfig.SelectorTorrentDownloadLink,
			selectorTorrentDetailsLink:  siteConfig.SelectorTorrentDetailsLink,
			selectorTorrentTime:         siteConfig.SelectorTorrentTime,
			selectorTorrentSize:         siteConfig.SelectorTorrentSize,
			selectorTorrentSeeders:      siteConfig.SelectorTorrentSeeders,
			selectorTorrentLeechers:     siteConfig.SelectorTorrentLeechers,
			selectorTorrentSnatched:     siteConfig.SelectorTorrentSnatched,
			selectorTorrentFree:         siteConfig.SelectorTorrentFree,
		},
	}
	if siteConfig.TorrentUrlIdRegexp != "" {
		site.torrentsParserOption.idRegexp = regexp.MustCompile(siteConfig.TorrentUrlIdRegexp)
	}
	return site, nil
}

func init() {
	site.Register(&site.RegInfo{
		Name:    "unit3d",
		Creator: NewSite,
	})
}
