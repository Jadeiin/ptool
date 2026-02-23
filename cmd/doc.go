// Package cmd implements the command-line interface for ptool.
//
// It uses the Cobra library for command management and provides an interactive
// shell mode using go-prompt for enhanced user experience.
//
// # Command Structure
//
// Commands are organized into sub-packages by functionality:
//   - Brush: Automatic brushing (brush)
//   - Xseed: Cross-seeding (iyuu, reseed, xseedadd)
//   - Client: BT client control (status, add, delete, pause, resume, etc.)
//   - Torrent: Torrent file utilities (parsetorrent, verifytorrent, maketorrent, etc.)
//   - Site: PT site functions (search, batchdl, publish, sites)
//   - CookieCloud: Cookie synchronization (cookiecloud)
//   - Utility: Other tools (shell, config, version, alias)
//
// # Usage
//
// Basic command execution:
//
//	ptool <command> [args...] [flags]
//
// Examples:
//
//	// Brush (auto download and manage torrents)
//	ptool brush local keepfrds
//
//	// Cross-seed using IYUU
//	ptool iyuu xseed local
//
//	// Show client status
//	ptool status local -t
//
//	// Search site torrents
//	ptool search keepfrds keyword
//
// # Global Flags
//
// Flags available for all commands:
//   - --config string: Specify config file path
//   - -v, -vv, -vvv: Increase verbosity level
//   - --timeout int: Network timeout in seconds
//   - --proxy string: HTTP proxy
//   - --timezone string: Override timezone
//   - --insecure: Skip TLS verification
//   - --lock string: Lock file path
//   - --global-lock: Use global lock
//
// # Interactive Shell
//
// Launch interactive shell mode:
//
//	ptool shell
//
// The interactive shell provides:
//   - Tab completion for commands and arguments
//   - Dynamic suggestions for client names, site names, etc.
//   - Command history
//
// # Exit Codes
//
//   - 0: Success
//   - 1: General error
//   - Other: Command-specific error codes
//
// # Registering Commands
//
// Sub-commands are registered in their respective sub-packages using init()
// functions that call cmd.RootCmd.AddCommand(). Each sub-package should be
// imported anonymously in cmd/all package to ensure registration.
//
// For more information about individual commands, see the documentation
// for specific sub-packages or run 'ptool <command> -h'.
package cmd
