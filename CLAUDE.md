# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**cliforge** is a shared Go library (module: `github.com/rios0rios0/cliforge`) providing reusable components for building CLI tools: cross-platform file operations and self-update from GitHub releases. It is not a standalone binary.

## Build & Development Commands

```bash
go build ./...          # Compile the library
go mod download         # Install dependencies
go test ./...           # Run all tests
go test -run TestName ./selfupdate  # Run a single test
make lint               # Lint (requires pipelines setup: make setup)
make test               # Run tests via Makefile
make sast               # Run security analysis
```

## Architecture

Two packages, both consumed as library imports by downstream CLI tools:

### `platform/` -- Cross-platform OS abstraction
- `OS` interface (`os.go`) defines 5 operations: `Download`, `Extract`, `Move`, `Remove`, `MakeExecutable`
- `OSUnix` (`os_unix.go`) implements via shell commands (`tar`, `mv`, `rm`, `chmod`)
- `OSWindows` (`os_windows.go`) implements via PowerShell
- `Info` (`platform.go`) normalizes `runtime.GOOS`/`runtime.GOARCH` (handles Android-to-Linux mapping)
- Build tags (`//go:build !windows` / `//go:build windows`) select the implementation at compile time

### `selfupdate/` -- GitHub release self-update
- `Command` (`selfupdate.go`) is the main public API. Created via `NewCommand(owner, repo, binaryName, currentVersion)`, executed via `Execute(dryRun, force)`
- Update flow: fetch latest GitHub release -> compare versions -> download matching asset -> extract -> backup current binary -> replace -> cleanup
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
