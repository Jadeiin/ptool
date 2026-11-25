package qbittorrent

import (
	"strings"

	qbt "github.com/autobrr/go-qbittorrent"

	"github.com/sagan/ptool/client"
	"github.com/sagan/ptool/util"
)

type torrentInfo qbt.Torrent

func (qt torrentInfo) canResume() bool {
	return qt.State == "pausedUP" || qt.State == "stoppedUP" || qt.State == "pausedDL" || qt.State == "stoppedDL" ||
		qt.State == "queuedUP" || qt.State == "queuedDL" || qt.State == "error"
}

func (qbtorrent torrentInfo) toTorrentState() string {
	state := ""
	switch qbtorrent.State {
	case "stalledUP", "queuedUP", "forcedUP", "uploading":
		state = "seeding"
	case "metaDL", "forcedMetaDL", "allocating", "stalledDL", "queuedDL", "forcedDL", "downloading":
		state = "downloading"
	case "pausedUP", "stoppedUP":
		state = "completed"
	case "pausedDL", "stoppedDL":
		state = "paused"
	case "checkingUP", "checkingDL", "checkingResumeData":
		state = "checking"
	case "error", "missingFiles", "unknown":
		state = "error"
	default:
		state = "unknown"
	}
	return state
}

func (qbtorrent torrentInfo) toTorrent() *client.Torrent {
	torrent := &client.Torrent{
		InfoHash:           qbtorrent.Hash,
		Name:               qbtorrent.Name,
		TrackerDomain:      util.ParseUrlHostname(qbtorrent.Tracker),
		TrackerBaseDomain:  util.GetUrlDomain(qbtorrent.Tracker),
		Tracker:            qbtorrent.Tracker,
		State:              qbtorrent.toTorrentState(),
		LowLevelState:      string(qbtorrent.State),
		Atime:              qbtorrent.AddedOn,
		Ctime:              qbtorrent.CompletionOn,
		ActivityTime:       qbtorrent.LastActivity,
		Downloaded:         qbtorrent.Downloaded,
		DownloadSpeed:      qbtorrent.DlSpeed,
		DownloadSpeedLimit: qbtorrent.DlLimit,
		Uploaded:           qbtorrent.Uploaded,
		UploadSpeed:        qbtorrent.UpSpeed,
		UploadedSpeedLimit: qbtorrent.UpLimit,
		Category:           qbtorrent.Category,
		SavePath:           qbtorrent.SavePath,
		ContentPath:        qbtorrent.contentPath(),
		Tags:               util.SplitCsv(qbtorrent.Tags),
		Seeders:            qbtorrent.NumComplete,
		Size:               qbtorrent.Size,
		SizeCompleted:      qbtorrent.Completed,
		SizeTotal:          qbtorrent.TotalSize,
		Leechers:           qbtorrent.NumIncomplete,
		Meta:               map[string]int64{},
	}
	torrent.Name, torrent.Meta = client.ParseMetaFromName(torrent.Name)
	return torrent
}

// Return the real content path (root folder) of torrent. It differs with qBittorrent ContentPath in a special case:
// For a single file torrent. qb put abs path of that single file as ContentPath, even if it has a wrap folder.
// E.g. if a torrent has a name "foo" and only one file "bar.txt" with a save path "/downloads",
// ContentPath will be "/downloads/foo/bar.txt".
// This function returns "/downloads/foo" in this case.
func (qbtorrent torrentInfo) contentPath() string {
	sep := qbtorrent.sep()
	cp := qbtorrent.ContentPath
	if !strings.HasPrefix(cp, qbtorrent.SavePath+sep) {
		return cp
	}
	relativepath := cp[len(qbtorrent.SavePath)+1:]
	i := strings.Index(relativepath, sep)
	if i == -1 {
		return cp
	}
	return qbtorrent.SavePath + sep + relativepath[:i]
}

// Return path sep (either '/' or '\') of this torrent.
func (qbtorrent torrentInfo) sep() string {
	if !strings.Contains(qbtorrent.SavePath, `/`) && !strings.Contains(qbtorrent.ContentPath, `/`) {
		return `\`
	}
	return `/`
}
