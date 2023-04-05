package nexusphp

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sagan/ptool/config"
	"github.com/sagan/ptool/site"
	"github.com/sagan/ptool/utils"
)

type Site struct {
	Name       string
	SiteConfig *config.SiteConfigStruct
	Config     *config.ConfigStruct
	HttpClient *http.Client
}

func (npclient *Site) GetName() string {
	return npclient.Name
}

func (npclient *Site) GetSiteConfig() *config.SiteConfigStruct {
	return npclient.SiteConfig
}

func (npclient *Site) DownloadTorrent(url string) ([]byte, error) {
	res, err := utils.FetchUrl(url, npclient.SiteConfig.Cookie, npclient.HttpClient)
	if err != nil {
		log.Fatal("Can not fetch torrents from site")
	}
	return io.ReadAll(res.Body)
}

func (npclient *Site) GetLatestTorrents(url string) ([]site.SiteTorrent, error) {
	if url == "" {
		url = npclient.SiteConfig.Url + "torrents.php"
	}
	res, err := utils.FetchUrl(url, npclient.SiteConfig.Cookie, npclient.HttpClient)
	if err != nil {
		return nil, fmt.Errorf("can not fetch torrents from site: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("can not fetch torrents from site: status=%d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("parse torrents page DOM error: %v", err)
	}
	headerTr := doc.Find("table.torrents > tbody > tr").First()
	if headerTr.Length() == 0 {
		return nil, fmt.Errorf("no torrents found in page, possible a parser error")
	}
	fieldColumIndex := map[string]int{
		"time":     -1,
		"size":     -1,
		"seeders":  -1,
		"leechers": -1,
		"snatched": -1,
	}
	processFieldIndex := -1 // m-team
	headerTr.Children().Each(func(i int, s *goquery.Selection) {
		for field := range fieldColumIndex {
			if s.Find("*[alt=\""+field+"\"]").Length() > 0 {
				fieldColumIndex[field] = i
				break
			}
		}
		if s.Text() == "進度" || s.Text() == "进度" {
			processFieldIndex = i
		}
	})
	torrents := []site.SiteTorrent{}
	doc.Find("table.torrents > tbody > tr").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}
		name := ""
		downloadUrl := ""
		size := int64(0)
		seeders := int64(0)
		leechers := int64(0)
		snatched := int64(0)
		time := int64(0)
		hnr := false
		downloadMultiplier := 1.0
		uploadMultiplier := 1.0
		discountEndTime := int64(-1)
		isActive := false
		var error error = nil
		processValueRegexp := regexp.MustCompile(`\d+(\.\d+)?%`)

		s.Children().Each(func(i int, s *goquery.Selection) {
			for field, index := range fieldColumIndex {
				if processFieldIndex == i {
					if m := processValueRegexp.MatchString(s.Text()); m {
						isActive = true
					}
					continue
				}
				if index != i {
					continue
				}
				switch field {
				case "size":
					size, _ = utils.RAMInBytes(s.Text())
				case "seeders":
					seeders = utils.ParseInt(s.Text())
				case "leechers":
					leechers = utils.ParseInt(s.Text())
				case "snatched":
					snatched = utils.ParseInt(s.Text())
				case "time":
					title := s.Find("*[title]").AttrOr("title", "")
					time, error = utils.ParseTime(title)
					if error == nil {
						break
					}
					time, error = utils.ParseTime((s.Text()))
				}
			}
		})
		titleEl := s.Find("a[href^=\"details.php?\"]")
		if titleEl.Length() > 0 {
			name = titleEl.Text()
			name = strings.ReplaceAll(name, "[email protected]", "") // CloudFlare email obfuscation sometimes confuses with 0day torrent names such as "***-DIY@Audies"
		}
		downloadEl := s.Find("a[href^=\"download.php?\"]")
		if downloadEl.Length() > 0 {
			downloadUrl = npclient.SiteConfig.Url + downloadEl.AttrOr("href", "")
		}
		if s.Find(`*[title="H&R"],*[alt="H&R"]`).Length() > 0 {
			hnr = true
		}
		if s.Find(`*[title="免费"],*[alt="Free"]`).Length() > 0 {
			downloadMultiplier = 0
		}
		if s.Find(`*[title^="seeding"],*[title^="leeching"],*[title^="downloading"],*[title^="uploading"]`).Length() > 0 {
			isActive = true
		}
		re := regexp.MustCompile(`剩余(时间)?\s*(：|:)\s*(?P<time>[YMDHMSymdhms年月天时時分秒\d]+)`)
		m := re.FindStringSubmatch(s.Text())
		if m != nil {
			discountEndTime, _ = utils.ParseFutureTime(m[re.SubexpIndex("time")])
		}
		if name != "" && downloadUrl != "" {
			torrents = append(torrents, site.SiteTorrent{
				Name:               name,
				Size:               size,
				DownloadUrl:        downloadUrl,
				Leechers:           leechers,
				Seeders:            seeders,
				Snatched:           snatched,
				Time:               time,
				HasHnR:             hnr,
				DownloadMultiplier: downloadMultiplier,
				UploadMultiplier:   uploadMultiplier,
				DiscountEndTime:    discountEndTime,
				IsActive:           isActive,
			})
		}
	})
	return torrents, nil
}

func NewSite(name string, siteConfig *config.SiteConfigStruct, config *config.ConfigStruct) (site.Site, error) {
	client := &Site{
		Name:       name,
		SiteConfig: siteConfig,
		Config:     config,
		HttpClient: &http.Client{},
	}
	return client, nil
}

func init() {
	site.Register(&site.RegInfo{
		Name:    "nexusphp",
		Creator: NewSite,
	})
}

var (
	_ site.Site = (*Site)(nil)
)
