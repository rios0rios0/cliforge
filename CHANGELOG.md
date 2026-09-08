# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This file is not edited by hand. Every change writes its own fragment under
`.changes/unreleased/` with [chlog](https://github.com/luizjhonata/chlog), and a release compiles
the pending fragments into a version section here — so two branches each adding an entry no
longer touch the same lines, and a rebase that used to conflict on this file now conflicts on
nothing.

## [Unreleased]

## [0.4.5] - 2026-09-08

### Changed

- changed both `chlog new` examples in the AI-assistant instruction block of `CLAUDE.md` and `.github/copilot-instructions.md` to `--body '<past-tense description>'`: changelog bodies here are written in simple past tense, and the body is single-quoted because it carries backticks that a double-quoted shell argument would command-substitute, and added the line telling the reader to write an apostrophe inside the single-quoted body as `'\''`, since bodies here carry possessives, and switched the 4 other hand-written `chlog new` examples in `CONTRIBUTING.md`, `.github/pull_request_template.md`, `.github/pull_request_template/default.md`, and `.github/skills/code-review/SKILL.md` to the same single-quoted body argument

## [0.4.4] - 2026-09-03

### Changed

- changed the Go module dependencies to their latest versions

## [0.4.3] - 2026-09-02

### Changed

- changed the Go version to `1.27.1` and updated all module dependencies

## [0.4.2] - 2026-08-29

### Changed

- changed the Go module dependencies to their latest versions

## [0.4.1] - 2026-08-28

### Changed

- changed the Claude workflows to call the reusable workflows in `rios0rios0/pipelines` instead of `rios0rios0/.github`, which is where every other reusable workflow and composite action already lives, and renamed them to `claude-review.yaml` and `claude-mention.yaml`, matching the `reusable-claude-review.yaml` / `reusable-claude-mention.yaml` definitions they call
- changed the Go module dependencies to their latest versions

### Fixed

- changed the Claude workflow callers to single-quote every `types:` sequence entry, per the account YAML standard, and renamed the changelog fragment added by the previous change to chlog's documented `<unix-nanoseconds>-<four hex characters>.yaml` form -- its former suffix was not hexadecimal. The mention workflow is now named `Claude Mention` with a matching `claude-mention` job id.
- restored the `.changes/unreleased/` directory with a `.gitkeep`, so the release tooling keeps recognising this project as [chlog](https://github.com/luizjhonata/chlog)-based after a release consumes the last fragment. Git tracks files rather than directories, so the bump commit that removed the final fragment removed the directory too, and the next run read the empty `[Unreleased]` section as "nothing to release"
- restored the `id-token: write` permission on both Claude workflow callers. Without it the caller grants less than the reusable workflow declares, which GitHub rejects before the job starts -- runs ended in `startup_failure`. The action needs the scope because `setupGitHubToken()` exchanges a GitHub OIDC token for the GitHub App token it posts with, unless a `github_token` is passed explicitly.

### Removed

- removed the unused `id-token: write` permission from the Claude workflow callers, and changed `claude-review.yaml`'s display name to `Claude Review` so it matches its file name and its `Claude Mention` sibling. `anthropics/claude-code-action` needs `id-token: write` only for workload identity federation or the Bedrock / Vertex / Foundry OIDC paths; these authenticate with `claude_code_oauth_token`, so the scope allowed minting OIDC tokens for any audience without ever being used.

## [0.4.0] - 2026-08-26

### Added

- added a tailored `code-review` skill under `.github/skills/` so GitHub Copilot reviews changes against the [rios0rios0/guide](https://github.com/rios0rios0/guide/wiki) standards and this repository's own load-bearing invariants

### Changed

- changed the changelog to [chlog](https://github.com/luizjhonata/chlog) fragments: a change now writes its own YAML file under `.changes/unreleased/` through `chlog new --kind <Kind> --body "..."`, and `CHANGELOG.md` is GENERATED from them at release time by `chlog batch auto && chlog merge`. That is the one thing a single shared file cannot do — two branches each adding an entry no longer touch the same lines, so a rebase that used to conflict on `CHANGELOG.md` now conflicts on nothing. The `[Unreleased]` section was empty, so nothing had to be carried across. AutoBump already reads the fragments directly, so the release flow is unchanged.
- changed the Go module dependencies to their latest versions

### Fixed

- fixed the `main` pipeline, which every repository's `sast:gitleaks` job had been failing since the code-review skill landed: the skill's own security bullet listed credential prefixes verbatim to warn against writing them, and the scanner's second pass matches those prefixes on their own, so the warning tripped the rule it was describing. The bullet now names the vendors instead, and the commit that carried the original wording is allowlisted by fingerprint in `.gitleaksignore`, because the scan walks the whole history reachable from `HEAD` and no edit at the tip can clear a past commit. No credential was ever committed.

## [0.3.19] - 2026-08-25

### Changed

- changed the Go module dependencies to their latest versions

## [0.3.18] - 2026-08-24

### Changed

- changed the Go module dependencies to their latest versions
- changed the Go version to `1.27.0` and updated all module dependencies

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

