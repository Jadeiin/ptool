// Package config provides configuration management for ptool.
//
// Configuration is loaded from TOML or YAML files, with TOML being the recommended format.
// The default configuration file location is:
//   - Linux: ~/.config/ptool/ptool.toml
//   - Windows: %USERPROFILE%\.config\ptool\ptool.toml
//
// The configuration file consists of several sections:
//   - Global settings (iyuuToken, reseedUsername, etc.)
//   - [[clients]]: BitTorrent client configurations
//   - [[sites]]: PT site configurations
//   - [[groups]]: Site groups for batch operations
//   - [[aliases]]: Command aliases
//   - [[cookieclouds]]: CookieCloud server configurations
//
// # Example Configuration
//
//	iyuuToken = "IYUU001122..."
//
//	[[clients]]
//	name = "local"
//	type = "qbittorrent"
//	url = "http://localhost:8080/"
//	username = "admin"
//	password = "adminadmin"
//
//	[[sites]]
//	type = "keepfrds"
//	cookie = "cookie_here"
//
//	[[groups]]
//	name = "acg"
//	sites = ["u2", "kamept"]
//
// # Configuration Types
//
// The main configuration types are:
//   - ConfigStruct: Root configuration structure
//   - ClientConfigStruct: BitTorrent client settings
//   - SiteConfigStruct: PT site settings
//   - GroupConfigStruct: Site group definition
//   - AliasConfigStruct: Command alias definition
//   - CookiecloudConfigStruct: CookieCloud server settings
//
// # Constants
//
// The package defines various constants for default values and internal use:
//   - Category names: BRUSH_CAT, SEEDING_CAT, DYNAMIC_SEEDING_CAT_PREFIX
//   - Tag names: XSEED_TAG, NOADD_TAG, NODEL_TAG, HR_TAG
//   - Default values for timeouts, speed limits, etc.
//
// # Usage
//
// Load and access configuration:
//
//	config.Init()
//	cfg := config.Get()
//	for _, client := range cfg.ClientsEnabled {
//	    fmt.Println(client.Name)
//	}
//
// Access individual client or site config:
//
//	clientConfig := config.GetClientConfig("local")
//	siteConfig := config.GetSiteConfig("keepfrds")
//
// For more details, see the documentation for individual types and functions.
package config
