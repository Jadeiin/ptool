package reseed

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/sagan/ptool/config"
	"github.com/sagan/ptool/constants"
	"github.com/sagan/ptool/site"
	"github.com/sagan/ptool/util"
	"github.com/shibumi/go-pathspec"
	log "github.com/sirupsen/logrus"
)

// Reseed exposes Socket.IO and HTTP APIs. The backend source is at
// https://github.com/tongyifan/Reseed-backend.
// All APIs (except "login" API) requires authorization token in header: "Authorization: Bearar <token>".
// To acquire token: POST https://reseed-api.tongyifan.me/login with username & password,
// receive json {msg, success: true, token}. token is ephemeral, expires in 1 day.
// The Socket.IO server returns HTTP 500 when the token is missing.
// With an invalid or expired token the server can leave the request waiting for results.
// The Socket.IO API uses these events.
// "file" event: client -> server.
// "reseed result" event: server -> client.
// The backend expects ASCII frames, so JSON must escape Unicode as \uXXXX.
const RESEED_API = "https://reseed-api.tongyifan.me/"

type Site struct {
	BaseUrl string `json:"base_url,omitempty"` // e.g. "https://hdtime.org/"
	Name    string `json:"name,omitempty"`     // e.g. "HDTIME"
	// we do not need below fields:
	// _enable
	// _passkey
}

// the "file" event payload of Reseed sock.io API
type File map[string]any

type ReseedResultSite struct {
	// reseed torrent id
	Id int64 `json:"id,omitempty"`
	// comma-separated site torrent ids, e.g. NexusHD-123456,HDU-23456,TJUPT-34567
	Sites string `json:"sites,omitempty"`
}

type Torrent struct {
	Id       string // local site torrent id, e.g. hdupt.23456
	ReseedId string // Reseed torrent id, e.g. HDU-23456
	SavePath string // Local path of torrent contents, e.g. D:\Downloads
	Filename string // Filename of torrent content file or root folder
	Success  bool   // true: a full match (success). false: partial-match (warning)
}

// the 'reseed result' event returned by Reseed backend
// see https://github.com/tongyifan/Reseed-backend/blob/890dfcb20b98684bf315c8c9f5352c062ae93166/views/reseed.py#L57
type ReseedResult struct {
	Name       string             `json:"name,omitempty"`
	CmpSuccess []ReseedResultSite `json:"cmp_success,omitempty"`
	CmpWarning []ReseedResultSite `json:"cmp_warning,omitempty"`
}

func (t *Torrent) String() string {
	return t.Id
}

// It's performance is terrible, but who cares?
func (f File) MarshalJSON() ([]byte, error) {
	if bytes, err := json.Marshal(map[string]any(f)); err != nil {
		return nil, err
	} else {
		return []byte(Escape2Anscii(string(bytes))), nil
	}
}

func (site *Site) MatchFilter(filter string) bool {
	return filter == "" || util.ContainsI(site.Name, filter) || util.ContainsI(site.BaseUrl, filter)
}

// Escape all unicode (non-ASCII) characters to '\uXXXX' format.
func Escape2Anscii(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r > 0xFFFF {
			r1, r2 := utf16.EncodeRune(r)
			sb.WriteString(fmt.Sprintf("\\u%04X", r1))
			sb.WriteString(fmt.Sprintf("\\u%04X", r2))
		} else if r > 0x7F {
			sb.WriteString(fmt.Sprintf("\\u%04X", r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// Request Reseed API and return xseed torrents (full match (success) & partial match (warning) results)
// found by Reseed backend.
func GetReseedTorrents(username string, password string, sites []*config.SiteConfigStruct, timeout int64,
	savePath ...string) (results []*Torrent, results2 []*Torrent, err error) {
	file, savePathMap, err := scan(savePath...)
	if err != nil {
		err = fmt.Errorf("failed to scan savePath(s): %w", err)
		return
	}
	if len(file) == 0 {
		log.Debugf("All savePath does NOT has any contents")
		return
	}
	token, err := Login(username, password)
	if err != nil {
		err = fmt.Errorf("failed to login to reseed server: %w", err)
		return
	}
	reseedSites, err := GetSites(token)
	if err != nil {
		err = fmt.Errorf("failed to get reseed sites: %w", err)
		return
	}
	reseed2LocalMap := GenerateReseed2LocalSiteMap(reseedSites, sites)
	cntResult := 0
	err = requestReseedResults(RESEED_API, token, file, time.Duration(timeout)*time.Second, func(result ReseedResult) {
		log.Tracef("reseed result: %v", result)
		cntResult++
		results = append(results, parseReseedResult(reseed2LocalMap, savePathMap,
			result.Name, true, result.CmpSuccess)...)
		results2 = append(results2, parseReseedResult(reseed2LocalMap, savePathMap,
			result.Name, false, result.CmpWarning)...)
	})
	if cntResult == 0 && err == nil {
		log.Debugf("server did not return any response")
	}
	return
}

// Scan top-level "Download" dirs and generate Reseed "file" request payload
func scan(dirs ...string) (file File, savePathMap map[string]string, err error) {
	file = File{}
	savePathMap = map[string]string{}
	for _, dir := range dirs {
		err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				if path != dir && info.IsDir() {
					log.Errorf("Failed to access %s (dir=%t, err=%v), skip it", info.Name(), info.IsDir(), err)
					return filepath.SkipDir
				}
				return err
			}
			if !info.Mode().IsRegular() && !info.IsDir() {
				return nil
			}
			relpath, err := filepath.Rel(dir, path)
			if err != nil || relpath == "" || relpath == "." {
				return nil // just ignore it
			}
			relpath = strings.ReplaceAll(relpath, `/`, `\`)
			inTopLevelDir := !strings.Contains(relpath, `\`)
			isTorrentFile := !info.IsDir() && strings.HasSuffix(util.TrimAnySuffix(path,
				constants.ProcessedFilenameSuffixes...), ".torrent")
			ignore := false
			if inTopLevelDir {
				ignore = isTorrentFile || util.First(pathspec.GitIgnore(constants.DefaultIgnorePatterns, relpath))
			} else {
				ignore = isTorrentFile
			}
			if ignore {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
			if index := strings.Index(relpath, `\`); index == -1 { // top-level item
				savePathMap[relpath] = dir
				if info.IsDir() {
					file[relpath] = map[string]any{}
				} else {
					file[relpath] = info.Size()
				}
			} else {
				dirFile := file[relpath[:index]].(map[string]any)
				dirFile[relpath[index+1:]] = info.Size()
			}
			return nil
		})
		if err != nil {
			return
		}
	}
	return
}

// https://reseed-api.tongyifan.me/sites_info
func GetSites(token string) ([]Site, error) {
	var sites []Site
	err := util.FetchJson(RESEED_API+"sites_info", &sites, nil, http.Header{
		"Authorization": []string{"Bearar " + token},
	})
	if err != nil {
		return nil, err
	}
	return sites, nil
}

// return ReseedSite => localSiteName map
func GenerateReseed2LocalSiteMap(reseedSites []Site,
	localSites []*config.SiteConfigStruct) map[string]string {
	iyuu2LocalSiteMap := map[string]string{} // iyuu sid => local site name
	for _, reseedSite := range reseedSites {
		localSite := util.FindInSlice(localSites, func(siteConfig *config.SiteConfigStruct) bool {
			reseedSiteName := strings.ToLower(reseedSite.Name)
			if siteConfig.Url != "" && siteConfig.Url == reseedSite.BaseUrl {
				return true
			}
			regInfo := site.GetConfigSiteReginfo(siteConfig.GetName())
			return regInfo != nil && (regInfo.Name == reseedSiteName || slices.Contains(regInfo.Aliases, reseedSiteName))
		})
		if localSite != nil {
			iyuu2LocalSiteMap[reseedSite.Name] = (*localSite).GetName()
		}
	}
	return iyuu2LocalSiteMap
}

type ReseedLoginResult struct {
	Msg     string
	Success bool
	Token   string
}

func Login(username string, password string) (token string, err error) {
	data := url.Values{
		"username": []string{username},
		"password": []string{password},
	}
	var result *ReseedLoginResult
	if err = util.PostUrlForJson(RESEED_API+"login", data, &result, nil, nil); err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	if !result.Success {
		return "", fmt.Errorf("failed login: %s", result.Msg)
	}
	return result.Token, nil
}

func parseReseedResult(reseed2LocalMap map[string]string, savePathMap map[string]string, name string, success bool,
	sites []ReseedResultSite) (results []*Torrent) {
	for _, successResult := range sites {
		torrents := util.SplitCsv(successResult.Sites)
		for _, torrent := range torrents {
			id := ""
			info := strings.Split(torrent, "-")
			if len(info) == 2 && reseed2LocalMap[info[0]] != "" {
				id = reseed2LocalMap[info[0]] + "." + info[1]
			}
			results = append(results, &Torrent{
				Id:       id,
				ReseedId: torrent,
				SavePath: savePathMap[name],
				Filename: name,
				Success:  success,
			})
		}
	}
	return
}
