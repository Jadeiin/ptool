// Package client provides interfaces and implementations for interacting with BitTorrent clients.
//
// It supports multiple BitTorrent client types:
//   - qBittorrent v4.1+ (recommended)
//   - Transmission (<= v3.0)
//
// # Core Types
//
// The main types for interacting with clients are:
//   - Client: Interface for BitTorrent client operations
//   - Torrent: Represents a torrent in the client
//   - Status: Client status information
//   - TorrentContentFile: Individual file within a torrent
//   - TorrentTracker: Tracker information for a torrent
//
// # Tracker Validity
//
// TrackerValidity represents the validity status of a torrent's tracker:
//   - TRACKER_VALIDITY_OK: Normal status
//   - TRACKER_VALIDITY_VIOLATE_RULE: Exceeding simultaneous client limit
//   - TRACKER_VALIDITY_INVALID_AUTH: Invalid passkey/authkey
//   - TRACKER_VALIDITY_NOT_EXIST: Torrent not registered on tracker
//
// # Usage
//
// Get a client instance:
//
//	clientInstance, err := client.GetClient("local")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// List torrents:
//
//	torrents, err := clientInstance.GetTorrents("", "", true)
//	for _, t := range torrents {
//	    fmt.Printf("%s: %s\n", t.InfoHash, t.Name)
//	}
//
// Add a torrent:
//
//	err = clientInstance.AddTorrent(content, "", "", false)
//
// The package also provides utility functions for common operations like
// parsing info hashes, checking tracker status, and filtering torrents.
package client
