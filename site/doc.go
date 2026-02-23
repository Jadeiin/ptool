// Package site provides interfaces and implementations for interacting with PT (Private Tracker) sites.
//
// It supports various site architectures:
//   - NexusPHP (most common)
//   - GazellePW
//   - Unit3D
//   - TNode
//   - Discuz
//   - MTorrent (M-Team)
//
// # Core Types
//
// The main types for site operations are:
//   - Site: Interface for PT site operations
//   - Torrent: Represents a torrent on the site
//   - Status: User status information (uploaded, downloaded, etc.)
//   - RegInfo: Site registration information
//
// # Site Interface
//
// The Site interface provides methods for:
//   - Downloading torrents by ID or URL
//   - Getting latest torrents list
//   - Searching torrents
//   - Publishing new torrents
//   - Getting user status
//
// # Usage
//
// Get a site instance:
//
//	siteInstance, err := site.GetSite("keepfrds")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Get latest torrents:
//
//	torrents, err := siteInstance.GetLatestTorrents(false)
//	for _, t := range torrents {
//	    fmt.Printf("%s: %s (%s)\n", t.Id, t.Name, util.BytesSize(t.Size))
//	}
//
// Search torrents:
//
//	results, err := siteInstance.SearchTorrents("keyword", "")
//
// Download torrent:
//
//	content, filename, id, err := siteInstance.DownloadTorrent("mteam.12345")
//
// The package maintains a registry of supported sites that can be accessed
// using site names, aliases, or by iterating through all registered sites.
package site
