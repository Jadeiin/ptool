package transmission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	transmissionrpc "github.com/hekmon/transmissionrpc/v2"
)

const sessionIDHeader = "X-Transmission-Session-Id"

var trackerHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// Transmission 3 expects integer and string pairs in trackerReplace. The v2
// module exposes this field as []string, so send this one operation directly.
func (trclient *Client) replaceTracker(ctx context.Context, torrentID, trackerID int64, trackerURL string) error {
	body, err := json.Marshal(map[string]any{
		"method": "torrent-set",
		"arguments": map[string]any{
			"ids":            []int64{torrentID},
			"trackerReplace": []any{trackerID, trackerURL},
		},
	})
	if err != nil {
		return err
	}
	sessionID, _ := trclient.trackerSessionID.Load().(string)
	for range 2 {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, trclient.rpcURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("User-Agent", "github.com/hekmon/transmissionrpc")
		request.Header.Set(sessionIDHeader, sessionID)
		request.SetBasicAuth(trclient.ClientConfig.Username, trclient.ClientConfig.Password)
		response, err := trackerHTTPClient.Do(request)
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
