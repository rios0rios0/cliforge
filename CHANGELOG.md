# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- added `CheckForUpdates` method for automatic version checking on CLI startup
- added `ShouldCheckForUpdates` function to skip checks when binary was modified today

### Changed

- changed the Go module dependencies to their latest versions
- changed the Go version to `1.26.2` and updated all module dependencies

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

