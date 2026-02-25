package unit3d

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/sagan/ptool/site"
	"github.com/sagan/ptool/util"
)

const (
	SELECTOR_TORRENTS_LIST       = "table.data-table, div.table-responsive table"
	SELECTOR_TORRENT_BLOCK       = "tbody > tr"
	SELECTOR_DOWNLOAD_LINK       = "a[href*='/download/']"
	SELECTOR_DOWNLOAD_CHECK_LINK = "a[href*='/download_check/']"
	SELECTOR_DETAILS_LINK        = "a.view-torrent, a.torrent-search--list__name"
	SELECTOR_TIME                = "time, td[class*='created_at'], td[class$='age-header'], td span[title]"
	SELECTOR_SIZE                = "td[class*='torrent-listings-size'], td[class$='size-header'], td:has(i.fa-database), td.torrent-search--list__size"
	SELECTOR_SEEDERS             = "td:has(i.fa-arrow-alt-circle-up), td.torrent-search--list__seeders"
	SELECTOR_LEECHERS            = "td:has(i.fa-arrow-alt-circle-down), td.torrent-search--list__leechers"
	SELECTOR_COMPLETED           = "td:has(i.fa-check-circle), td.torrent-search--list__completed"
	SELECTOR_AUTHOR              = "span:has(> i.fa-upload), span.torrent-search--list__uploader"
	SELECTOR_CATEGORY            = "a[href*='/categories/'] > div > img, a[href*='/categories/'] > div > i, div.torrent-search--list__category img"
	SELECTOR_COMMENTS            = "a[href*='#comments'], i.torrent-icons__comments"

	// 优惠信息选择器
	SELECTOR_FREE       = "i.fa-star.text-gold, i.fa-globe.text-blue, i.torrent-icons__freeleech, i[title*='100% Free'], span[title*='100% Free']"
	SELECTOR_2XUP       = "i.fa-gem.text-green, i.fa-chevron-double-up, i.torrent-icons__double-upload, i[title*='Double Upload']"
	SELECTOR_75PERCENT  = "i[title*='75%'], span[title*='75%']"
	SELECTOR_50PERCENT  = "i[title*='50%'], span[title*='50%']"
	SELECTOR_25PERCENT  = "i[title*='25%'], span[title*='25%']"
	SELECTOR_FEATURED   = "i.torrent-icons__featured, i[title*='Featured']"
	SELECTOR_INTERNAL   = "i.torrent-icons__internal"
	SELECTOR_PERSONAL   = "i.torrent-icons__personal-release"
	SELECTOR_HIGHSPEED  = "i.torrent-icons__highspeed"
	SELECTOR_REFUNDABLE = "i.fa-percentage, i[title*='Refundable']"
	SELECTOR_PINNED     = "i.fa-thumbtack"

	// H&R 选择器
	SELECTOR_HNR = "i.fa-exclamation-circle.text-red, i[title*='H&R'], i[title*='Hit and Run']"
)

type TorrentsParserOption struct {
	location                    *time.Location
	siteurl                     string
	idRegexp                    *regexp.Regexp
	selectorTorrentsList        string
	selectorTorrentBlock        string
	selectorTorrentDownloadLink string
	selectorTorrentDetailsLink  string
	selectorTorrentTime         string
	selectorTorrentSize         string
	selectorTorrentSeeders      string
	selectorTorrentLeechers     string
	selectorTorrentSnatched     string
	selectorTorrentFree         string
	selectorTorrent2xup         string
}

// parseTorrents 解析种子列表页面
func parseTorrents(doc *goquery.Document, option *TorrentsParserOption,
	doctime int64, sitename string) (torrents []*site.Torrent, err error) {
	torrents = []*site.Torrent{}

	// 使用默认选择器（如果未指定）
	selectorTorrentsList := option.selectorTorrentsList
	if selectorTorrentsList == "" {
		selectorTorrentsList = SELECTOR_TORRENTS_LIST
	}
	selectorTorrentBlock := option.selectorTorrentBlock
	if selectorTorrentBlock == "" {
		selectorTorrentBlock = SELECTOR_TORRENT_BLOCK
	}
	selectorTorrentDetailsLink := option.selectorTorrentDetailsLink
	if selectorTorrentDetailsLink == "" {
		selectorTorrentDetailsLink = SELECTOR_DETAILS_LINK
	}
	selectorTorrentDownloadLink := option.selectorTorrentDownloadLink
	if selectorTorrentDownloadLink == "" {
		selectorTorrentDownloadLink = SELECTOR_DOWNLOAD_LINK
	}
	selectorTorrentTime := option.selectorTorrentTime
	if selectorTorrentTime == "" {
		selectorTorrentTime = SELECTOR_TIME
	}
	selectorTorrentSize := option.selectorTorrentSize
	if selectorTorrentSize == "" {
		selectorTorrentSize = SELECTOR_SIZE
	}
	selectorTorrentSeeders := option.selectorTorrentSeeders
	if selectorTorrentSeeders == "" {
		selectorTorrentSeeders = SELECTOR_SEEDERS
	}
	selectorTorrentLeechers := option.selectorTorrentLeechers
	if selectorTorrentLeechers == "" {
		selectorTorrentLeechers = SELECTOR_LEECHERS
	}
	selectorTorrentSnatched := option.selectorTorrentSnatched
	if selectorTorrentSnatched == "" {
		selectorTorrentSnatched = SELECTOR_COMPLETED
	}

	torrentsList := doc.Find(selectorTorrentsList)
	if torrentsList.Length() == 0 {
		return nil, fmt.Errorf("no torrents list found with selector: %s", selectorTorrentsList)
	}

	log.Tracef("Unit3D parser found table with selector: %s", selectorTorrentsList)

	// 尝试获取表头信息以确定列位置
	fieldIndex := map[string]int{
		"time":     -1,
		"size":     -1,
		"seeders":  -1,
		"leechers": -1,
		"snatched": -1,
	}

	// 查找表头
	headerRow := torrentsList.Find("thead > tr")
	if headerRow.Length() > 0 {
		headerRow.Find("th").Each(func(i int, th *goquery.Selection) {
			html, _ := th.Html()
			text := util.DomSanitizedText(th)
			class, _ := th.Attr("class")

			// 检测时间列
			if strings.Contains(html, "created_at") ||
				strings.HasSuffix(class, "age-header") ||
				strings.Contains(text, "时间") ||
				th.Find("i.fa-clock").Length() > 0 {
				fieldIndex["time"] = i
				return
			}

			// 检测大小列
			if strings.Contains(class, "torrent-listings-size") ||
				strings.HasSuffix(class, "size-header") ||
				th.Find("i.fa-database").Length() > 0 ||
				strings.Contains(text, "大小") ||
				strings.Contains(text, "Size") {
				fieldIndex["size"] = i
				return
			}

			// 检测种子数列
			if th.Find("i.fa-arrow-alt-circle-up").Length() > 0 ||
				strings.Contains(text, "种子") ||
				strings.Contains(text, "Seeders") {
				fieldIndex["seeders"] = i
				return
			}

			// 检测下载数列
			if th.Find("i.fa-arrow-alt-circle-down").Length() > 0 ||
				strings.Contains(text, "下载") ||
				strings.Contains(text, "Leechers") {
				fieldIndex["leechers"] = i
				return
			}

			// 检测完成数列
			if th.Find("i.fa-check-circle").Length() > 0 ||
				strings.Contains(text, "完成") ||
				strings.Contains(text, "Completed") {
				fieldIndex["snatched"] = i
				return
			}
		})
	}

	log.Tracef("Unit3D parser field index: %v", fieldIndex)

	// 遍历种子行
	torrentRows := torrentsList.Find(selectorTorrentBlock)
	torrentRows.Each(func(i int, row *goquery.Selection) {
		// 跳过表头行
		if row.Find("th").Length() > 0 {
			return
		}

		cells := row.Find("td")
		if cells.Length() == 0 {
			return
		}

		var name, id string
		var description string
		var downloadUrl string
		var size int64
		var seeders, leechers, snatched int64
		var timestamp int64
		var author string
		var tags []string
		downloadMultiplier := 1.0
		uploadMultiplier := 1.0
		hasHnR := false
		isActive := false
		isCurrentActive := false

		// 获取标题和详情链接
		titleEl := row.Find(selectorTorrentDetailsLink).First()
		if titleEl.Length() > 0 {
			name = strings.TrimSpace(titleEl.Text())
			detailsUrl := titleEl.AttrOr("href", "")

			// 解析ID
			if option.idRegexp != nil {
				if m := option.idRegexp.FindStringSubmatch(detailsUrl); m != nil {
					id = m[option.idRegexp.SubexpIndex("id")]
				}
			}
			// 默认ID解析
			if id == "" {
				if matches := regexp.MustCompile(`/torrents/(\d+)`).FindStringSubmatch(detailsUrl); matches != nil {
					id = matches[1]
				}
			}

			// 尝试获取副标题
			subtitleEl := row.Find(".torrent-search--list__subtitle, .torrent__subtitle, small.text-muted")
			if subtitleEl.Length() > 0 {
				description = strings.TrimSpace(subtitleEl.Text())
			}
		}

		// 获取下载链接
		downloadEl := row.Find(selectorTorrentDownloadLink).First()
		if downloadEl.Length() == 0 {
			// 尝试 download_check 链接
			downloadEl = row.Find(SELECTOR_DOWNLOAD_CHECK_LINK).First()
			if downloadEl.Length() > 0 {
				href := downloadEl.AttrOr("href", "")
				downloadUrl = strings.Replace(href, "/download_check/", "/download/", 1)
			}
		} else {
			downloadUrl = downloadEl.AttrOr("href", "")
		}

		if downloadUrl != "" && !util.IsUrl(downloadUrl) {
			downloadUrl = option.siteurl + strings.TrimPrefix(downloadUrl, "/")
		}

		// 获取分类
		categoryEl := row.Find(SELECTOR_CATEGORY).First()
		if categoryEl.Length() > 0 {
			// 尝试获取 data-original-title 或 alt 属性
			catTitle := categoryEl.AttrOr("data-original-title", "")
			if catTitle == "" {
				catTitle = categoryEl.AttrOr("alt", "")
			}
			// 移除 " torrent" 后缀
			catTitle = strings.TrimSuffix(catTitle, " torrent")
			catTitle = strings.TrimSuffix(catTitle, " Torrent")
			if catTitle != "" {
				tags = append(tags, catTitle)
			}
		}

		// 使用时间列索引或选择器获取时间
		if fieldIndex["time"] >= 0 {
			timeCell := cells.Eq(fieldIndex["time"])
			timeStr := timeCell.Find("span[title]").AttrOr("title", "")
			if timeStr == "" {
				timeStr = timeCell.AttrOr("title", "")
			}
			if timeStr == "" {
				timeStr = util.DomSanitizedText(timeCell)
			}
			timestamp, _ = util.ParseTime(timeStr, option.location)
		} else {
			timeEl := row.Find(selectorTorrentTime).First()
			if timeEl.Length() > 0 {
				timeStr := timeEl.AttrOr("title", "")
				if timeStr == "" {
					timeStr = util.DomSanitizedText(timeEl)
				}
				timestamp, _ = util.ParseTime(timeStr, option.location)
			}
		}

		// 使用大小列索引或选择器获取大小
		if fieldIndex["size"] >= 0 {
			sizeCell := cells.Eq(fieldIndex["size"])
			size, _ = util.RAMInBytes(util.DomSanitizedText(sizeCell))
		} else {
			sizeEl := row.Find(selectorTorrentSize).First()
			if sizeEl.Length() > 0 {
				size, _ = util.RAMInBytes(util.DomSanitizedText(sizeEl))
			}
		}

		// 使用种子数列索引或选择器
		if fieldIndex["seeders"] >= 0 {
			seeders = util.ParseInt(util.DomSanitizedText(cells.Eq(fieldIndex["seeders"])))
		} else {
			seeders = util.ParseInt(util.DomSanitizedText(row.Find(selectorTorrentSeeders).First()))
		}

		// 使用下载数列索引或选择器
		if fieldIndex["leechers"] >= 0 {
			leechers = util.ParseInt(util.DomSanitizedText(cells.Eq(fieldIndex["leechers"])))
		} else {
			leechers = util.ParseInt(util.DomSanitizedText(row.Find(selectorTorrentLeechers).First()))
		}

		// 使用完成数列索引或选择器
		if fieldIndex["snatched"] >= 0 {
			snatched = util.ParseInt(util.DomSanitizedText(cells.Eq(fieldIndex["snatched"])))
		} else {
			snatched = util.ParseInt(util.DomSanitizedText(row.Find(selectorTorrentSnatched).First()))
		}

		// 获取发布者
		authorEl := row.Find(SELECTOR_AUTHOR).First()
		if authorEl.Length() > 0 {
			author = strings.TrimSpace(authorEl.Text())
			author = regexp.MustCompile(`\s*\(`).ReplaceAllString(author, " (")
		}

		// 检测优惠信息（优先级：Free > 75% > 50% > 25%）
		if row.Find(SELECTOR_FREE).Length() > 0 ||
			row.Find("i.torrent-icons__featured").Length() > 0 {
			downloadMultiplier = 0
			tags = append(tags, "Free")
		} else if row.Find(SELECTOR_75PERCENT).Length() > 0 {
			downloadMultiplier = 0.25
			tags = append(tags, "75%")
		} else if row.Find(SELECTOR_50PERCENT).Length() > 0 {
			downloadMultiplier = 0.5
			tags = append(tags, "50%")
		} else if row.Find(SELECTOR_25PERCENT).Length() > 0 {
			downloadMultiplier = 0.75
			tags = append(tags, "25%")
		}

		// 检测2x上传
		if row.Find(SELECTOR_2XUP).Length() > 0 {
			uploadMultiplier = 2
			tags = append(tags, "2xUp")
		}

		// 检测其他标签
		if row.Find(SELECTOR_FEATURED).Length() > 0 {
			tags = append(tags, "Featured")
		}
		if row.Find(SELECTOR_INTERNAL).Length() > 0 {
			tags = append(tags, "Internal")
		}
		if row.Find(SELECTOR_PERSONAL).Length() > 0 {
			tags = append(tags, "Personal")
		}
		if row.Find(SELECTOR_HIGHSPEED).Length() > 0 {
			tags = append(tags, "Highspeed")
		}
		if row.Find(SELECTOR_REFUNDABLE).Length() > 0 {
			tags = append(tags, "Refundable")
		}
		if row.Find(SELECTOR_PINNED).Length() > 0 {
			tags = append(tags, "置顶")
		}

		// 检测H&R
		if row.Find(SELECTOR_HNR).Length() > 0 {
			hasHnR = true
			tags = append(tags, "H&R")
		}

		// 检测状态（做种/下载中）
		if row.Find("i.fa-arrow-circle-up, i.torrent-icons__seeding").Length() > 0 {
			isActive = true
			isCurrentActive = true
		} else if row.Find("i.fa-arrow-circle-down, i.torrent-icons__downloading").Length() > 0 {
			isActive = true
			isCurrentActive = true
		} else if row.Find("i.fa-do-not-enter, i.torrent-icons__inactive").Length() > 0 {
			isActive = true
		}

		// 去重标签
		tags = util.UniqueSlice(tags)

		// 添加种子到列表
		if name != "" && (downloadUrl != "" || id != "") {
			if id != "" && sitename != "" {
				id = sitename + "." + id
			}
			torrents = append(torrents, &site.Torrent{
				Name:               name,
				Description:        description,
				Id:                 id,
				Size:               size,
				DownloadUrl:        downloadUrl,
				Seeders:            seeders,
				Leechers:           leechers,
				Snatched:           snatched,
				Time:               timestamp,
				HasHnR:             hasHnR,
				DownloadMultiplier: downloadMultiplier,
				UploadMultiplier:   uploadMultiplier,
				IsActive:           isActive,
				IsCurrentActive:    isCurrentActive,
				Tags:               tags,
			})
		}
	})

	log.Tracef("Unit3D parser found %d torrents", len(torrents))
	return torrents, nil
}
