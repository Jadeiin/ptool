package qbittorrent

// qb web API: https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	qbt "github.com/autobrr/go-qbittorrent"
	log "github.com/sirupsen/logrus"

	"github.com/sagan/ptool/client"
	"github.com/sagan/ptool/config"
	"github.com/sagan/ptool/constants"
	"github.com/sagan/ptool/util"
	"github.com/sagan/ptool/util/torrentutil"
)

type Client struct {
	Name                      string
	ClientConfig              *config.ClientConfigStruct
	Config                    *config.ConfigStruct
	webAPI                    *qbt.Client
	apiBaseURL                *url.URL
	data                      *qbt.MainData
	preferences               map[string]any
	datatime                  int64
	unfinishedSize            int64
	unfinishedDownloadingSize int64
	contentPathTorrents       map[string][]string
}

func (qbclient *Client) GetTorrentsByContentPath(contentPath string) ([]*client.Torrent, error) {
	err := qbclient.sync()
	if err != nil {
		return nil, err
	}
	var torrents []*client.Torrent
	for _, hash := range qbclient.contentPathTorrents[contentPath] {
		torrent := torrentInfo(qbclient.data.Torrents[hash])
		torrents = append(torrents, torrent.toTorrent())
	}
	return torrents, nil
}

func (qbclient *Client) SetAllTorrentsShareLimits(ratioLimit float64, seedingTimeLimit int64) error {
	return qbclient.SetTorrentsShareLimits([]string{"all"}, ratioLimit, seedingTimeLimit)
}

func (qbclient *Client) SetTorrentsShareLimits(infoHashes []string, ratioLimit float64, seedingTimeLimit int64) error {
	if len(infoHashes) == 0 {
		return nil
	}
	if ratioLimit == 0 {
		ratioLimit = -2 // use global limit
	} else if ratio := strconv.FormatFloat(ratioLimit, 'f', 2, 64); ratio == "0.00" || ratio == "-0.00" {
		return fmt.Errorf("share ratio limit %g rounds to zero at two decimal places", ratioLimit)
	}
	if seedingTimeLimit == 0 {
		seedingTimeLimit = -2 // use global limit
	} else if seedingTimeLimit > 0 {
		seedingTimeLimit = seedingTimeLimit/60 + 1
	}
	return qbclient.webAPI.SetTorrentShareLimit(infoHashes, qbt.ShareLimitOptions{
		RatioLimit:               ratioLimit,
		SeedingTimeLimit:         seedingTimeLimit,
		InactiveSeedingTimeLimit: -2,
	})
}

func (qbclient *Client) GetName() string {
	return qbclient.Name
}

func (qbclient *Client) GetClientConfig() *config.ClientConfigStruct {
	return qbclient.ClientConfig
}

func (qbclient *Client) AddTorrent(torrentContent []byte, option *client.TorrentOption, meta map[string]int64) error {
	options := map[string]string{
		"root_folder":   "true",
		"contentLayout": "Original",
	}
	name := ""
	if option != nil {
		name = option.Name
		if option.Category != constants.NONE {
			options["category"] = option.Category
		}
		options["tags"] = strings.Join(option.Tags, ",")
		options["paused"] = strconv.FormatBool(option.Pause)
		options["stopped"] = options["paused"]
		if option.SkipChecking {
			if option.Pause {
				options["stopCondition"] = "FilesChecked"
			}
			options["skip_checking"] = "true"
			options["seedMode"] = "true" // qBittorrent 5.3+ renamed skip_checking.
		}
		options["upLimit"] = strconv.FormatInt(option.UploadSpeedLimit, 10)
		options["dlLimit"] = strconv.FormatInt(option.DownloadSpeedLimit, 10)
		if option.SavePath != "" {
			options["savepath"] = option.SavePath
			options["autoTMM"] = "false"
		}
		if option.SequentialDownload {
			options["sequentialDownload"] = "true"
		}
		if option.RatioLimit != 0 {
			options["ratioLimit"] = fmt.Sprint(option.RatioLimit)
		}
		if option.SeedingTimeLimit != 0 {
			value := option.SeedingTimeLimit
			if value > 0 {
				value = value/60 + 1
			}
			options["seedingTimeLimit"] = strconv.FormatInt(value, 10)
		}
	}
	options["rename"] = client.GenerateNameWithMeta(name, meta)
	var result *qbt.TorrentAddResponse
	var err error
	torrentURL := string(torrentContent)
	if util.IsTorrentUrl(torrentURL) {
		result, err = qbclient.webAPI.AddTorrentFromUrl(torrentURL, options)
	} else {
		result, err = qbclient.webAPI.AddTorrentFromMemory(torrentContent, options)
	}
	if err != nil {
		return fmt.Errorf("add torrent error: %w", err)
	}
	if result.FailureCount > 0 {
		return qbt.ErrTorrentAddFailed
	}
	return nil
}

func (qbclient *Client) PauseTorrents(infoHashes []string) error {
	if len(infoHashes) == 0 {
		return nil
	}
	return qbclient.webAPI.Pause(infoHashes)
}

func (qbclient *Client) ResumeTorrents(infoHashes []string) error {
	if len(infoHashes) == 0 {
		return nil
	}
	return qbclient.webAPI.Resume(infoHashes)
}

func (qbclient *Client) RecheckTorrents(infoHashes []string) error {
	if len(infoHashes) == 0 {
		return nil
	}
	return qbclient.webAPI.Recheck(infoHashes)
}

func (qbclient *Client) ReannounceTorrents(infoHashes []string) error {
	if len(infoHashes) == 0 {
		return nil
	}
	return qbclient.webAPI.ReAnnounceTorrents(infoHashes)
}

func (qbclient *Client) AddTagsToTorrents(infoHashes []string, tags []string) error {
	if len(infoHashes) == 0 || len(tags) == 0 {
		return nil
	}
	return qbclient.webAPI.AddTags(infoHashes, strings.Join(tags, ","))
}

func (qbclient *Client) RemoveTagsFromTorrents(infoHashes []string, tags []string) error {
	if len(infoHashes) == 0 || len(tags) == 0 {
		return nil
	}
	return qbclient.webAPI.RemoveTags(infoHashes, strings.Join(tags, ","))
}

func (qbclient *Client) SetTorrentsSavePath(infoHashes []string, savePath string) error {
	if len(infoHashes) == 0 {
		return nil
	}
	savePath = strings.TrimSpace(savePath)
	if savePath == "" {
		return fmt.Errorf("savePath is empty")
	}
	return qbclient.webAPI.SetLocation(infoHashes, savePath)
}

func (qbclient *Client) PauseAllTorrents() error {
	return qbclient.PauseTorrents([]string{"all"})
}

func (qbclient *Client) ResumeAllTorrents() error {
	return qbclient.ResumeTorrents([]string{"all"})
}

func (qbclient *Client) RecheckAllTorrents() error {
	return qbclient.RecheckTorrents([]string{"all"})
}

func (qbclient *Client) ReannounceAllTorrents() error {
	return qbclient.ReannounceTorrents([]string{"all"})
}

func (qbclient *Client) AddTagsToAllTorrents(tags []string) error {
	return qbclient.AddTagsToTorrents([]string{"all"}, tags)
}

func (qbclient *Client) RemoveTagsFromAllTorrents(tags []string) error {
	return qbclient.RemoveTagsFromTorrents([]string{"all"}, tags)
}

func (qbclient *Client) SetAllTorrentsSavePath(savePath string) error {
	return qbclient.SetTorrentsSavePath([]string{"all"}, savePath)
}

func (qbclient *Client) GetTags() ([]string, error) {
	return qbclient.webAPI.GetTags()
}

func (qbclient *Client) CreateTags(tags ...string) error {
	return qbclient.webAPI.CreateTags(tags)
}

func (qbclient *Client) DeleteTags(tags ...string) error {
	return qbclient.webAPI.DeleteTags(tags)
}

func (qbclient *Client) MakeCategory(category string, savePath string) error {
	path := savePath
	if path == constants.NONE {
		path = ""
	}
	err := qbclient.webAPI.CreateCategory(category, path)
	if errors.Is(err, qbt.ErrInvalidCategoryName) {
		if savePath != constants.NONE {
			return qbclient.webAPI.EditCategory(category, path)
		}
		return nil
	}
	return err
}

func (qbclient *Client) DeleteCategories(categories []string) error {
	return qbclient.webAPI.RemoveCategories(categories)
}

func (qbclient *Client) GetCategories() ([]*client.TorrentCategory, error) {
	categories, err := qbclient.webAPI.GetCategories()
	if err != nil {
		return nil, err
	}
	cats := make([]*client.TorrentCategory, 0, len(categories))
	for _, category := range categories {
		cats = append(cats, &client.TorrentCategory{Name: category.Name, SavePath: category.SavePath})
	}
	return cats, nil
}

func (qbclient *Client) SetTorrentsCatetory(infoHashes []string, category string) error {
	if len(infoHashes) == 0 {
		return nil
	}
	if category == constants.NONE {
		category = ""
	}
	return qbclient.webAPI.SetCategory(infoHashes, category)
}

func (qbclient *Client) SetAllTorrentsCatetory(category string) error {
	return qbclient.SetTorrentsCatetory([]string{"all"}, category)
}

func (qbclient *Client) DeleteTorrents(infoHashes []string, deleteFiles bool) error {
	if len(infoHashes) == 0 {
		return nil
	}
	if err := qbclient.webAPI.DeleteTorrents(infoHashes, deleteFiles); err != nil {
		return err
	}
	if qbclient.Cached() {
		for _, infoHash := range infoHashes {
			delete(qbclient.data.Torrents, infoHash)
		}
		qbclient.buildDerivative()
	}
	return nil
}

func (qbclient *Client) ModifyTorrent(infoHash string,
	option *client.TorrentOption, meta map[string]int64) error {
	if option == nil {
		option = &client.TorrentOption{}
	}
	err := qbclient.sync()
	if err != nil {
		return err
	}

	torrent, ok := qbclient.data.Torrents[infoHash]
	if !ok {
		return fmt.Errorf("torrent not exists")
	}
	qbtorrent := torrentInfo(torrent)
	infoHashes := []string{infoHash}

	if option.Name != "" || len(meta) > 0 {
		name := option.Name
		if name == "" {
			name, _ = client.ParseMetaFromName(qbtorrent.Name)
		}
		name = client.GenerateNameWithMeta(name, meta)
		if name != qbtorrent.Name {
			if err := qbclient.webAPI.SetTorrentName(infoHash, name); err != nil {
				return err
			}
		}
	}

	// @todo: apply option.SequentialDownload using qb toggleSequentialDownload.
	// However, we must know current sequentialDownload status in ahead.

	if option.Category != "" {
		category := option.Category
		if category == constants.NONE {
			category = ""
		}
		if category != qbtorrent.Category {
			if err := qbclient.webAPI.SetCategory(infoHashes, category); err != nil {
				return err
			}
		}
	}

	if len(option.Tags) > 0 || len(option.RemoveTags) > 0 {
		qbTags := util.SplitCsv(qbtorrent.Tags)
		addTags := []string{}
		removeTags := []string{}
		for _, addTag := range option.Tags {
			if !slices.Contains(qbTags, addTag) {
				addTags = append(addTags, addTag)
			}
		}
		for _, removeTag := range option.RemoveTags {
			if slices.Contains(qbTags, removeTag) {
				removeTags = append(removeTags, removeTag)
			}
		}
		if len(removeTags) > 0 {
			if err := qbclient.RemoveTagsFromTorrents(infoHashes, removeTags); err != nil {
				return err
			}
		}
		if len(addTags) > 0 {
			if err := qbclient.AddTagsToTorrents(infoHashes, addTags); err != nil {
				return err
			}
		}
	}

	if option.DownloadSpeedLimit != 0 && option.DownloadSpeedLimit != qbtorrent.DlLimit {
		if err := qbclient.webAPI.SetTorrentDownloadLimit(infoHashes, option.DownloadSpeedLimit); err != nil {
			return err
		}
	}

	if option.UploadSpeedLimit != 0 && option.UploadSpeedLimit != qbtorrent.UpLimit {
		if err := qbclient.webAPI.SetTorrentUploadLimit(infoHashes, option.UploadSpeedLimit); err != nil {
			return err
		}
	}

	if option.RatioLimit != 0 || option.SeedingTimeLimit != 0 {
		if err := qbclient.SetTorrentsShareLimits(infoHashes, option.RatioLimit, option.SeedingTimeLimit); err != nil {
			return err
		}
	}

	if option.Pause {
		if !qbtorrent.canResume() {
			return qbclient.PauseTorrents(infoHashes)
		}
	} else if option.Resume {
		if qbtorrent.canResume() {
			return qbclient.ResumeTorrents(infoHashes)
		}
	}

	return nil
}

func (qbclient *Client) PurgeCache() {
	qbclient.data = nil
	qbclient.preferences = nil
	qbclient.unfinishedSize = 0
	qbclient.unfinishedDownloadingSize = 0
	qbclient.datatime = 0
	qbclient.contentPathTorrents = nil
}

func (qbclient *Client) Cached() bool {
	return qbclient.datatime > 0
}

func (qbclient *Client) sync() error {
	if qbclient.datatime > 0 {
		return nil
	}
	data, err := qbclient.webAPI.SyncMainDataCtx(context.Background(), 0)
	if err != nil {
		return err
	}
	qbclient.data = data
	qbclient.datatime = util.Now()
	qbclient.buildDerivative()
	return nil
}

func (qbclient *Client) buildDerivative() {
	unfinishedSize := int64(0)
	unfinishedDownloadingSize := int64(0)
	contentPathTorrents := map[string][]string{}
	for hash, torrent := range qbclient.data.Torrents {
		usize := torrent.Size - torrent.Completed
		unfinishedSize += usize
		if torrent.State != "pausedDL" && torrent.State != "stoppedDL" {
			unfinishedDownloadingSize += usize
		}
		path := torrentInfo(torrent).contentPath()
		contentPathTorrents[path] = append(contentPathTorrents[path], hash)
	}
	qbclient.unfinishedSize = unfinishedSize
	qbclient.unfinishedDownloadingSize = unfinishedDownloadingSize
	qbclient.contentPathTorrents = contentPathTorrents
}

func (qbclient *Client) TorrentRootPathExists(rootFolder string) bool {
	if rootFolder == "" {
		return false
	}
	err := qbclient.sync()
	if err != nil {
		return false
	}
	for _, torrent := range qbclient.data.Torrents {
		info := torrentInfo(torrent)
		if strings.HasSuffix(info.contentPath(), info.sep()+rootFolder) {
			return true
		}
	}
	return false
}

func (qbclient *Client) GetStatus() (*client.Status, error) {
	err := qbclient.sync()
	if err != nil {
		return nil, err
	}
	var status client.Status
	server := qbclient.data.ServerState
	status.DownloadSpeed = server.DlInfoSpeed
	status.UploadSpeed = server.UpInfoSpeed
	status.DownloadSpeedLimit = server.DlRateLimit
	status.UploadSpeedLimit = server.UpRateLimit
	status.FreeSpaceOnDisk = server.FreeSpaceOnDisk
	status.UnfinishedSize = qbclient.unfinishedSize
	status.UnfinishedDownloadingSize = qbclient.unfinishedDownloadingSize
	// @workaround
	// qb 的 Web API 有 bug，有时 FreeSpaceOnDisk 返回 0，但实际硬盘剩余空间充足，原因尚不明确。
	// 目前在 Windows QB 4.5.2 上发现此现象。
	if status.FreeSpaceOnDisk == 0 {
		hasDownloadingTorrent := false
		hasErrorTorrent := false
		for _, qbtorrent := range qbclient.data.Torrents {
			if qbtorrent.State == "downloading" {
				hasDownloadingTorrent = true
				break
			}
			if qbtorrent.State == "error" && qbtorrent.Completed < qbtorrent.Size {
				hasErrorTorrent = true
			}
		}
		if hasDownloadingTorrent || !hasErrorTorrent {
			status.FreeSpaceOnDisk = -1
		}
	}
	if slices.Contains(qbclient.data.Tags, config.NOADD_TAG) {
		status.NoAdd = true
	}
	if slices.Contains(qbclient.data.Tags, config.NODEL_TAG) {
		status.NoDel = true
	}
	return &status, nil
}

func (qbclient *Client) getPreferences() (map[string]any, error) {
	if qbclient.preferences != nil {
		return qbclient.preferences, nil
	}
	endpoint := qbclient.apiBaseURL.JoinPath("app/preferences")
	httpClient := qbclient.webAPI.GetHTTPClient()
	hasSession := len(httpClient.Jar.Cookies(endpoint)) > 0
	if !hasSession {
		if err := qbclient.webAPI.Login(); err != nil {
			return nil, err
		}
	}
	resp, err := httpClient.Get(endpoint.String())
	if err != nil {
		return nil, err
	}
	config := qbclient.ClientConfig
	if resp.StatusCode == http.StatusForbidden && hasSession && (config.Username != "" || config.Password != "") {
		resp.Body.Close()
		if err := qbclient.webAPI.Login(); err != nil {
			return nil, err
		}
		resp, err = httpClient.Get(endpoint.String())
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get preferences: status=%d", resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	var preferences map[string]any
	if err := decoder.Decode(&preferences); err != nil {
		return nil, err
	}
	qbclient.preferences = preferences
	return preferences, nil
}

func (qbclient *Client) GetConfig(variable string) (string, error) {
	if strings.HasPrefix(variable, "qb_") && len(variable) > 3 {
		preferences, err := qbclient.getPreferences()
		if err != nil {
			return "", err
		}
		key := variable[3:]
		if value, ok := preferences[key]; ok {
			return fmt.Sprint(value), nil
		}
		return "", fmt.Errorf("unknown qBittorrent preference: %s", key)
	}

	switch variable {
	case "global_download_speed_limit":
		v, err := qbclient.webAPI.GetGlobalDownloadLimit()
		return fmt.Sprint(v), err
	case "global_upload_speed_limit":
		v, err := qbclient.webAPI.GetGlobalUploadLimit()
		return fmt.Sprint(v), err
	case "free_disk_space":
		status, err := qbclient.GetStatus()
		if err != nil {
			return "", err
		}
		return fmt.Sprint(status.FreeSpaceOnDisk), nil
	case "global_download_speed":
		status, err := qbclient.GetStatus()
		if err != nil {
			return "", err
		}
		return fmt.Sprint(status.DownloadSpeed), nil
	case "global_upload_speed":
		status, err := qbclient.GetStatus()
		if err != nil {
			return "", err
		}
		return fmt.Sprint(status.UploadSpeed), nil
	case "save_path":
		return qbclient.webAPI.GetDefaultSavePath()
	default:
		return "", nil
	}
}

func (qbclient *Client) SetConfig(variable string, value string) error {
	if strings.HasPrefix(variable, "qb_") && len(variable) > 3 {
		preference, _ := util.String2Any(value)
		return qbclient.webAPI.SetPreferences(map[string]any{variable[3:]: preference})
	}
	switch variable {
	case "global_download_speed_limit", "global_upload_speed_limit":
		limit, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid speed limit: %w", err)
		}
		if variable == "global_download_speed_limit" {
			return qbclient.webAPI.SetGlobalDownloadLimit(limit)
		}
		return qbclient.webAPI.SetGlobalUploadLimit(limit)
	case "free_disk_space", "global_download_speed", "global_upload_speed":
		return fmt.Errorf("%s is read-only", variable)
	case "save_path":
		return qbclient.webAPI.SetPreferences(map[string]any{"save_path": value})
	default:
		return nil
	}
}

func (qbclient *Client) exportTorrentFileFromLocalTorrentsPath(infoHash string) ([]byte, error) {
	torrentFile := filepath.Join(qbclient.ClientConfig.LocalTorrentsPath, infoHash+".torrent")
	contents, err := os.ReadFile(torrentFile)
	if err != nil {
		return nil, err
	}
	metainfo, err := torrentutil.ParseTorrent(contents)
	if err != nil {
		return nil, err
	}
	// .torrent file in BT_backup could be trackerless: https://github.com/qbittorrent/qBittorrent/issues/18569 .
	// Extract trackers from .fastresume file instead in this case.
	if len(metainfo.Trackers) == 0 {
		fastresumeFile := filepath.Join(qbclient.ClientConfig.LocalTorrentsPath, infoHash+".fastresume")
		fastresume, err := parseQbFastresumeFile(fastresumeFile)
		if err != nil {
			return nil, err
		}
		if len(fastresume.Trackers) == 0 {
			return nil, fmt.Errorf("no trackers in fastresume file")
		}
		for tier, trackers := range fastresume.Trackers {
			for _, tracker := range trackers {
				metainfo.AddTracker(tracker, tier)
			}
		}
	}
	return metainfo.ToBytes()
}

// The export API of qb exists but currently is not documented in
// https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1) .
// See https://github.com/qbittorrent/qBittorrent/issues/18746 for more info.
func (qbclient *Client) ExportTorrentFile(infoHash string) ([]byte, error) {
	if qbclient.ClientConfig.LocalTorrentsPath != "" {
		contents, err := qbclient.exportTorrentFileFromLocalTorrentsPath(infoHash)
		if err == nil {
			return contents, nil
		}
		log.Debugf("Failed to export qb torrent from file system: %v. Fallback to QB Web API", err)
	}
	return qbclient.webAPI.ExportTorrent(infoHash)
}

func (qbclient *Client) GetTorrent(infoHash string) (*client.Torrent, error) {
	err := qbclient.sync()
	if err != nil {
		return nil, err
	}
	torrent, ok := qbclient.data.Torrents[infoHash]
	if !ok {
		return nil, nil
	}
	return torrentInfo(torrent).toTorrent(), nil
}

func (qbclient *Client) GetTorrents(stateFilter string, category string, showAll bool) ([]*client.Torrent, error) {
	err := qbclient.sync()
	if err != nil {
		return nil, err
	}
	torrents := []*client.Torrent{}

	for _, qbtorrent := range qbclient.data.Torrents {
		if category != "" {
			if category == constants.NONE {
				if qbtorrent.Category != "" {
					continue
				}
			} else if category != qbtorrent.Category {
				continue
			}
		}
		if !showAll && qbtorrent.DlSpeed < 1024 && qbtorrent.UpSpeed < 1024 {
			continue
		}
		torrent := torrentInfo(qbtorrent).toTorrent()
		if !torrent.MatchStateFilter(stateFilter) {
			continue
		}
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

func (qbclient *Client) GetTorrentContents(infoHash string) ([]*client.TorrentContentFile, error) {
	qbTorrentContents, err := qbclient.webAPI.GetFilesInformation(infoHash)
	if err != nil {
		return nil, err
	}
	torrentContents := make([]*client.TorrentContentFile, 0, len(*qbTorrentContents))
	for _, qbTorrentContent := range *qbTorrentContents {
		torrentContents = append(torrentContents, &client.TorrentContentFile{
			Index:    int64(qbTorrentContent.Index),
			Path:     strings.ReplaceAll(qbTorrentContent.Name, `\`, "/"),
			Size:     qbTorrentContent.Size,
			Ignored:  qbTorrentContent.Priority == 0,
			Complete: qbTorrentContent.IsSeed,
			Progress: float64(qbTorrentContent.Progress),
		})
	}
	sort.Slice(torrentContents, func(i, j int) bool {
		return torrentContents[i].Index < torrentContents[j].Index
	})
	return torrentContents, nil
}

func (qbclient *Client) GetTorrentTrackers(infoHash string) (client.TorrentTrackers, error) {
	qbTorrentTrackers, err := qbclient.webAPI.GetTorrentTrackers(infoHash)
	if err != nil {
		return nil, err
	}
	if qbTorrentTrackers == nil {
		return nil, fmt.Errorf("torrent %s: %w", infoHash, qbt.ErrTorrentNotFound)
	}
	qbTorrentTrackers = util.Filter(qbTorrentTrackers, func(tracker qbt.TorrentTracker) bool {
		// exclude qb  "** [DHT] **", "** [PeX] **", "** [LSD] **" trackers
		return !strings.HasPrefix(tracker.Url, "**")
	})
	trackers := util.Map(qbTorrentTrackers, func(qbtracker qbt.TorrentTracker) client.TorrentTracker {
		status := ""
		switch qbtracker.Status {
		case qbt.TrackerStatusDisabled:
			status = "disabled"
		case qbt.TrackerStatusNotContacted:
			status = "notcontacted"
		case qbt.TrackerStatusOK:
			status = "working"
		case qbt.TrackerStatusUpdating:
			status = "updating"
		case qbt.TrackerStatusNotWorking:
			status = "error"
		default:
			status = "unknown"
		}
		return client.TorrentTracker{
			Url:    qbtracker.Url,
			Msg:    qbtracker.Message,
			Status: status,
		}
	})
	return trackers, nil
}

func (qbclient *Client) EditTorrentTracker(infoHash string, oldTracker string,
	newTracker string, replaceHost bool) error {
	if replaceHost {
		torrent, err := qbclient.GetTorrent(infoHash)
		if err != nil {
			return err
		}
		if torrent == nil {
			return fmt.Errorf("torrent %s not found", infoHash)
		}
		trackers, err := qbclient.GetTorrentTrackers(torrent.InfoHash)
		if err != nil {
			return fmt.Errorf("failed to get torrent %s trackers: %w", torrent.InfoHash, err)
		}
		oldTrackerUrl := ""
		newTrackerUrl := ""
		directNewUrlMode := util.IsUrl(newTracker)
		index := trackers.FindIndex(oldTracker)
		if index != -1 {
			oldTrackerUrl = trackers[index].Url
			if directNewUrlMode {
				newTrackerUrl = newTracker
			} else {
				oldTrackerUrlObj, err := url.Parse(oldTrackerUrl)
				if err == nil {
					oldTrackerUrlObj.Host = newTracker
					newTrackerUrl = oldTrackerUrlObj.String()
				}
			}
		}
		if oldTrackerUrl != "" && newTrackerUrl != "" {
			if oldTrackerUrl == newTrackerUrl {
				return nil
			}
			err := qbclient.EditTorrentTracker(torrent.InfoHash, oldTrackerUrl, newTrackerUrl, false)
			if err != nil {
				log.Errorf("Failed to replace torrent %s tracker domain: %v", torrent.InfoHash, err)
			}
			return err
		}
		return fmt.Errorf("torrent %s old tracker does NOT exist", torrent.InfoHash)
	}
	return qbclient.webAPI.EditTracker(infoHash, oldTracker, newTracker)
}

// trackers - new trackers full URLs; oldTracker - existing tracker host or URL
func (qbclient *Client) AddTorrentTrackers(infoHash string, trackers []string,
	oldTracker string, removeExisting bool) error {
	var existingTrackers []string
	if oldTracker != "" {
		torrentTrackers, err := qbclient.GetTorrentTrackers(infoHash)
		if err != nil {
			return err
		}
		index := torrentTrackers.FindIndex(oldTracker)
		if index == -1 {
			return nil
		}
		if removeExisting {
			existingTrackers = util.Map(torrentTrackers, func(t client.TorrentTracker) string { return t.Url })
		} else {
			trackers = util.Filter(trackers, func(tracker string) bool {
				return torrentTrackers.FindIndex(tracker) == -1
			})
		}
	}
	if len(trackers) == 0 {
		return nil
	}
	urls := strings.Join(trackers, "\n")
	if err := qbclient.webAPI.AddTrackers(infoHash, urls); err != nil {
		return err
	}
	if removeExisting && len(existingTrackers) > 0 {
		return qbclient.RemoveTorrentTrackers(infoHash, existingTrackers)
	}
	return nil
}

func (qbclient *Client) RemoveTorrentTrackers(infoHash string, trackers []string) error {
	urls := strings.Join(trackers, "|")
	return qbclient.webAPI.RemoveTrackers(infoHash, urls)
}

func (qbclient *Client) SetFilePriority(infoHash string, fileIndexes []int64, priority int64) error {
	if len(fileIndexes) == 0 {
		return fmt.Errorf("must provide at least fileIndex")
	}
	ids := util.Map(fileIndexes, func(index int64) string { return strconv.FormatInt(index, 10) })
	return qbclient.webAPI.SetFilePriority(infoHash, strings.Join(ids, "|"), int(priority))
}

func (qbclient *Client) Close() {
	qbclient.PurgeCache()
	if qbclient.ClientConfig.QbittorrentNoLogout {
		return
	}
	logoutURL := qbclient.apiBaseURL.JoinPath("auth/logout")
	httpClient := qbclient.webAPI.GetHTTPClient()
	if len(httpClient.Jar.Cookies(logoutURL)) == 0 {
		return
	}
	resp, err := httpClient.PostForm(logoutURL.String(), nil)
	if err == nil {
		resp.Body.Close()
	}
}

func NewClient(name string, clientConfig *config.ClientConfigStruct, config *config.ConfigStruct) (
	client.Client, error) {
	baseURL, err := url.Parse(clientConfig.Url)
	if err != nil {
		return nil, err
	}
	apiBaseURL := baseURL.JoinPath("api/v2")
	apiBaseURL.RawQuery = ""
	webAPI := qbt.NewClient(qbt.Config{
		Host:     clientConfig.Url,
		Username: clientConfig.Username,
		Password: clientConfig.Password,
	})
	client := &Client{
		Name:         name,
		ClientConfig: clientConfig,
		Config:       config,
		webAPI:       webAPI,
		apiBaseURL:   apiBaseURL,
	}
	return client, nil
}

func init() {
	client.Register(&client.RegInfo{
		Name:    "qbittorrent",
		Creator: NewClient,
	})
}

var (
	_ client.Client = (*Client)(nil)
)
