# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**cliforge** is a shared Go library (module: `github.com/rios0rios0/cliforge`) providing reusable components for building CLI tools: cross-platform file operations and self-update from GitHub releases. It is not a standalone binary.

## Build & Development Commands

```bash
go build ./...          # Compile the library
go mod download         # Install dependencies
go test -tags unit ./...           # Run all tests
go test -tags unit -run TestName ./pkg/selfupdate  # Run a single test
make lint               # Lint (requires pipelines setup: make setup)
make test               # Run tests via Makefile
make sast               # Run security analysis
```

## Architecture

Two packages, both consumed as library imports by downstream CLI tools:

### `pkg/platform/` -- Cross-platform OS abstraction
- `OS` interface (`os.go`) defines 5 operations: `Download`, `Extract`, `Move`, `Remove`, `MakeExecutable`
- `OSUnix` (`os_unix.go`) implements via shell commands (`unzip`, `mv`, `rm`) and Go's `os.Chmod`
- `OSWindows` (`os_windows.go`) implements via PowerShell
- `Info` (`platform.go`) normalizes `runtime.GOOS`/`runtime.GOARCH` (handles Android-to-Linux mapping)
- `//go:build !windows` on `os_unix.go` and Go's `_windows.go` filename convention select the implementation at compile time

### `pkg/selfupdate/` -- GitHub release self-update
- `Command` (`selfupdate.go`) is the main public API. Created via `NewCommand(owner, repo, binaryName, currentVersion)`, executed via `Execute(dryRun, force)` or checked passively via `CheckForUpdates()`
- Update flow: fetch latest GitHub release -> compare versions -> download matching asset -> extract -> backup current binary -> replace -> cleanup
- `CheckForUpdates` (`check_for_updates.go`) passively checks for newer versions on CLI startup. Skips if the current version is `"dev"`, the binary was modified today, or a marker file under the user's cache directory (`os.UserCacheDir()`) was already touched today. The network call runs in a goroutine to avoid blocking startup. Errors are silently logged at debug level
- `ShouldCheckForUpdates` (`check_for_updates.go`) is a pure function that returns false when two timestamps fall on the same calendar day; used for both the binary modification time check and the daily marker file check
- `fetchLatestRelease` (`github.go`) calls `api.github.com` with 30s timeout, matches assets by pattern `{binary}-{version}-{os}-{arch}.{tar.gz|zip}`
- `CompareVersions` (`version.go`) implements semver comparison; treats `"dev"` as always older; pads unequal-length versions with zeros
- `extractArchive` (`archive.go`) delegates to platform-specific extraction

### Dependency flow
```
Consumer CLI tool
  -> selfupdate.Command
       -> platform.OS (interface, injected per OS via build tags)
       -> selfupdate.CompareVersions (pure function)
       -> selfupdate.fetchLatestRelease (HTTP + JSON)
       -> logrus (structured logging)
```

## Asset Naming Convention

The self-update system expects GoReleaser-standard asset names: `{binary}-{version}-{os}-{arch}.tar.gz` (Unix) or `.zip` (Windows).
