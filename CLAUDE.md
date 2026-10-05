# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**cliforge** is a shared Go library (module: `github.com/rios0rios0/cliforge`) providing reusable components for building CLI tools: cross-platform file operations and self-update from GitHub releases. It is not a standalone binary.

## Build & Development Commands

```bash
go build ./...          # Compile the library
go mod download         # Install dependencies
go test ./...           # Run all tests
go test -run TestName ./pkg/selfupdate  # Run a single test
make lint               # Lint (requires pipelines setup: make setup)
make test               # Run tests via Makefile
make sast               # Run security analysis
```

## Architecture

Two packages, both consumed as library imports by downstream CLI tools:

### `pkg/platform/` -- Cross-platform OS abstraction
- `OS` interface (`os.go`) defines 5 operations: `Download`, `Extract`, `Move`, `Remove`, `MakeExecutable`
- `OSUnix` (`os_unix.go`) and `OSWindows` (`os_windows.go`) share one pure-Go implementation: `Move` is `moveFile` (`move_file.go`), a rename that falls back to a copy only when the two paths sit on different volumes, finishing with a rename so the destination always becomes a new file; `Remove` is `os.Remove`; `Extract` is `extractZip` (`extract_zip.go`), which unpacks through an `os.Root` with names converted by `filepath.Localize`, regular files only, capped at 512 MiB. Only `MakeExecutable` (`os.Chmod` 0755 vs a no-op) and the cross-volume error (`syscall.EXDEV` vs `windows.ERROR_NOT_SAME_DEVICE`) differ per OS
- `Info` (`platform.go`) normalizes `runtime.GOOS`/`runtime.GOARCH` (handles Android-to-Linux mapping)
- `//go:build !windows` on `os_unix.go` and Go's `_windows.go` filename convention select the implementation at compile time

### `pkg/selfupdate/` -- GitHub release self-update
- `Command` (`selfupdate.go`) is the main public API. Created via `NewCommand(owner, repo, binaryName, currentVersion)`, executed via `Execute(dryRun, force)` or checked passively via `CheckForUpdates()`
- Update flow: fetch latest GitHub release -> compare versions -> download matching asset -> extract -> `installBinary` (`install_binary.go`): stage the new binary beside the running one as `<exe>.new` (the only step that can cross volumes) -> remove backups earlier updates left -> move the running binary to a unique `<exe>.backup-<unixnano>` -> move the staged binary into its place -> remove the backup, or keep it when Windows refuses because it is still the image of a running process (the next update removes it). Unique backup names keep a binary still running from an earlier backup, such as a daemon, from blocking the next update. The new binary always reaches the path as a new file, never written in place, so consumers can detect an install with `os.SameFile`
- `CheckForUpdates` (`check_for_updates.go`) passively checks for newer versions on CLI startup. Skips if the current version is `"dev"`, the binary was modified today, or a lookup already answered today. Its state lives under the user's cache directory (`os.UserCacheDir()`, see `update_check_state.go`): `last_update_check` is touched only after a lookup answers (after warning), so a command that exits before its lookup returns leaves the check to the next command; `update_check_attempts` counts today's started lookups and caps them at `maxUpdateCheckAttemptsPerDay` (5), so short-lived callers cannot hammer the API. When that state cannot be resolved, read or written, the lookup is skipped rather than run unthrottled. The lookup runs in the background (`Command.background`) to avoid blocking startup; errors are logged at debug level. `Command`'s `executable`, `cacheDir`, `now`, `background` and `apiBaseURL` fields exist so `export_test.go` can drive all of this deterministically
- `ShouldCheckForUpdates` (`check_for_updates.go`) is a pure function that returns false when two timestamps fall on the same calendar day; used for the binary modification time and for both state files
- `fetchLatestRelease` (`github.go`) calls `api.github.com` with 30s timeout, matches assets by pattern `{binary}-{version}-{os}-{arch}.{tar.gz|zip}`
- `CompareVersions` (`version.go`) implements semver comparison; treats `"dev"` as always older; pads unequal-length versions with zeros
- `extractArchive` (`archive.go`) runs `tar` on Unix and `platform.OS.Extract` (Go `archive/zip`) on Windows

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

<!-- chlog:start -->
## Changelog (chlog) — MANDATORY

If the repository you are working in uses chlog (a `.chlog.yaml` or `.chlog.yml`
config file, or a `.changes/` directory, exists at the project root), the
following is binding and ALWAYS applies: whenever you make ANY change, you MUST
create a changelog fragment as part of the same change — automatically, without
being asked, before committing.

- Do NOT edit CHANGELOG.md directly; it is generated from fragments.
- Create the fragment with:
  `chlog new --kind <Kind> --body '<past-tense description>'`
- Write an apostrophe inside the single-quoted body as `'\''`.
- Valid kinds: Added, Changed, Deprecated, Removed, Fixed, Security
- Choose the kind that best matches the change (e.g., new feature → Added,
  bug fix → Fixed, behavior change → Changed, removal → Removed, security fix → Security).
- If the change is backward-INCOMPATIBLE with the public API (a breaking
  change), you MUST add the `--breaking` flag:
  `chlog new --kind <Kind> --breaking --body '<past-tense description>'`.
  This is the ONLY thing that triggers a major version bump — the kind alone
  never does (per SemVer, major = incompatible change). When unsure whether a
  change breaks compatibility, ask the user instead of guessing.
- Fragments are YAML files in `.changes/unreleased/`; stage them with your commit.
- `chlog check` fails the build when a fragment is missing — never skip it.
<!-- chlog:end -->
