# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.17] - 2026-08-17

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.16] - 2026-08-16

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.15] - 2026-08-15

### Changed

- changed the Go version to `1.26.6` and updated all module dependencies

### Security

- changed the update-check marker directory mode from `0o750` to `0o700`, removing group access from a directory that is private to the current user

## [0.3.14] - 2026-07-14

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.13] - 2026-07-13

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.12] - 2026-07-10

### Changed

- changed the Go version to `1.26.5` and updated all module dependencies

### Security

- replaced `secrets: inherit` with an explicit `CLAUDE_CODE_OAUTH_TOKEN` secret in the Claude workflow callers, following the least-privilege principle

## [0.3.11] - 2026-06-09

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.10] - 2026-06-03

### Changed

- changed the Go version to `1.26.4` and updated all module dependencies

## [0.3.9] - 2026-05-25

### Changed

- refreshed `CLAUDE.md` and `.github/copilot-instructions.md` to correct `OSUnix` description (`unzip`, not `tar`; `os.Chmod`, not shell `chmod`)

## [0.3.8] - 2026-05-22

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.7] - 2026-05-19

### Changed

- changed the Go module dependencies to their latest versions
- refreshed `CLAUDE.md` and `.github/copilot-instructions.md` to fix `go test` commands (added missing `-tags unit` flag) and correct `os_windows.go` build-constraint description

## [0.3.6] - 2026-05-08

### Changed

- changed the Go version to `1.26.3` and updated all module dependencies

## [0.3.5] - 2026-04-29

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.4] - 2026-04-28

### Changed

- refreshed `CLAUDE.md` and `.github/copilot-instructions.md` to document the daily update-check throttle via marker file added in v0.3.3

## [0.3.3] - 2026-04-23

### Changed

- changed `CheckForUpdates` to throttle the GitHub API call to once per day across all CLI invocations by persisting a `last_update_check` marker file under the user's cache directory (`os.UserCacheDir()`)

## [0.3.2] - 2026-04-16

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.1] - 2026-04-15

### Changed

- changed the Go version to `1.26.2` and updated all module dependencies

## [0.3.0] - 2026-04-14

### Added

- added `CheckForUpdates` method for automatic version checking on CLI startup
- added `ShouldCheckForUpdates` function to skip checks when binary was modified today

### Changed

- changed the Go module dependencies to their latest versions

## [0.2.0] - 2026-04-03

### Added

- added CI/CD workflow, Makefile, contributing guide, and copilot instructions
- added test infrastructure with `OSStub` double and `OSStubBuilder`
- added unit tests for `CompareVersions` function

### Changed

- changed `PlatformInfo` to `Info` and `SelfUpdateCommand` to `Command` to avoid Go stuttering lint violations
- changed project structure to use `pkg/` directory for CI pipeline compatibility

## [0.1.0] - 2026-04-01

### Added

- added `CompareVersions` for semantic version comparison
- added `platform` package with cross-platform OS abstraction (Unix/Windows)
- added `selfupdate` package with parameterized self-update command for GitHub releases

### Changed

- changed the Go module dependencies to their latest versions

