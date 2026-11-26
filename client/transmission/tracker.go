package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	transmissionrpc "github.com/hekmon/transmissionrpc/v3"
	"github.com/sagan/ptool/util"
)

const sessionIDHeader = "X-Transmission-Session-Id"

func (trclient *Client) getTorrentWithTrackers(infoHash string) (*transmissionrpc.Torrent, error) {
	torrents, err := trclient.client.TorrentGetHashes(context.TODO(), []string{"id", "trackers"}, []string{infoHash})
	if err != nil {
		return nil, err
	}
	if len(torrents) == 0 {
		return nil, fmt.Errorf("torrent %s not found", infoHash)
	}
	return &torrents[0], nil
}

// The SDK only exposes whole-list replacement. Apply incremental changes on the
// daemon's current list so other tracker updates and fallback tiers are preserved.
func (trclient *Client) mutateTrackers(infoHash, field string, value any) error {
	body, err := json.Marshal(map[string]any{
		"method": "torrent-set",
		"arguments": map[string]any{
			"ids": []string{infoHash},
			field: value,
		},
	})
	if err != nil {
		return err
	}
	sessionID, _ := trclient.trackerSessionID.Load().(string)
	for range 2 {
		request, err := http.NewRequest(http.MethodPost, trclient.rpcURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(sessionIDHeader, sessionID)
		response, err := trclient.rpcHTTP.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode == http.StatusConflict {
			sessionID = response.Header.Get(sessionIDHeader)
			_, _ = io.Copy(io.Discard, response.Body)
			response.Body.Close()
			if sessionID == "" {
				return fmt.Errorf("Transmission response omitted %s", sessionIDHeader)
			}
			trclient.trackerSessionID.Store(sessionID)
			continue
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return transmissionrpc.HTTPStatusCode(response.StatusCode)
		}
		var reply struct {
			Result string `json:"result"`
		}
		if err := json.NewDecoder(response.Body).Decode(&reply); err != nil {
			return fmt.Errorf("decode Transmission response: %w", err)
		}
		if reply.Result != "success" {
			return fmt.Errorf("Transmission RPC result: %q", reply.Result)
		}
		return nil
	}
	return fmt.Errorf("Transmission rejected the session ID twice")
}

func (trclient *Client) EditTorrentTracker(infoHash, oldTracker, newTracker string, replaceHost bool) error {
	torrent, err := trclient.getTorrentWithTrackers(infoHash)
	if err != nil {
		return err
	}
	for _, tracker := range torrent.Trackers {
		address := tracker.Announce
		replacement := newTracker
		if replaceHost {
			trackerURL, err := url.Parse(address)
			if err != nil || trackerURL.Host != oldTracker {
				continue
			}
			if !util.IsUrl(newTracker) {
				trackerURL.Host = newTracker
				replacement = trackerURL.String()
			}
		} else if address != oldTracker {
			continue
		}
		if replacement == address {
			return nil
		}
		return trclient.mutateTrackers(infoHash, "trackerReplace", []any{tracker.ID, replacement})
	}
	return fmt.Errorf("torrent %s old tracker %s does NOT exist", infoHash, oldTracker)
}

func (trclient *Client) AddTorrentTrackers(infoHash string, trackers []string, oldTracker string, removeExisting bool) error {
	if len(trackers) == 0 {
		return nil
	}
	torrent := trclient.torrents[infoHash]
	if torrent == nil || oldTracker != "" || removeExisting {
		var err error
		torrent, err = trclient.getTorrentWithTrackers(infoHash)
		if err != nil {
			return err
		}
	}
	if oldTracker != "" && !slices.ContainsFunc(torrent.Trackers, func(tracker transmissionrpc.Tracker) bool {
		return util.MatchUrlWithHostOrUrl(tracker.Announce, oldTracker)
	}) {
		return nil
	}
	if !removeExisting {
		trackers = util.Filter(trackers, func(address string) bool {
			return !slices.ContainsFunc(torrent.Trackers, func(tracker transmissionrpc.Tracker) bool {
				return tracker.Announce == address
			})
		})
		if len(trackers) == 0 {
			return nil
		}
		return trclient.mutateTrackers(infoHash, "trackerAdd", trackers)
	}
	var replacement []string
	for _, address := range trackers {
		if strings.TrimSpace(address) == "" || strings.ContainsAny(address, "\r\n") {
			return fmt.Errorf("invalid tracker address %q", address)
		}
		if len(replacement) > 0 {
			replacement = append(replacement, "")
		}
		replacement = append(replacement, address)
	}
	return trclient.client.TorrentSet(context.TODO(), transmissionrpc.TorrentSetPayload{
		IDs:         []int64{*torrent.ID},
		TrackerList: replacement,
	})
}

func (trclient *Client) RemoveTorrentTrackers(infoHash string, trackers []string) error {
	torrent, err := trclient.getTorrentWithTrackers(infoHash)
	if err != nil {
		return err
	}
	var ids []int64
	for _, tracker := range torrent.Trackers {
		if slices.Contains(trackers, tracker.Announce) {
			ids = append(ids, tracker.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return trclient.mutateTrackers(infoHash, "trackerRemove", ids)
}
