# cliforge

cliforge is a shared Go library providing self-update and platform abstraction for CLI tools that distribute binaries via GitHub Releases. It is consumed by downstream CLI projects via Go module imports. This is a **library**, not a standalone binary -- there is no `main.go` or CLI.

Always reference these instructions first and fall back to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Test

- Install dependencies: `go mod download`
- Build (compile check): `go build ./...`
- Run tests: `make test` (preferred) or `go test ./...` for quick checks during development.
- Run linting: `make lint` -- NEVER run `golangci-lint` directly.
- Run security analysis: `make sast` -- NEVER run `gitleaks`, `semgrep`, `trivy`, `hadolint`, or `codeql` directly.
- Tidy dependencies: `go mod tidy`

### Linting, Testing, and SAST with Makefile

This project uses the [rios0rios0/pipelines](https://github.com/rios0rios0/pipelines) repository for shared CI/CD scripts. The `Makefile` imports these scripts via `SCRIPTS_DIR`. Always use `make` targets:

```bash
make lint    # golangci-lint via pipeline scripts
make test    # unit + integration tests via pipeline scripts
make sast    # CodeQL, Semgrep, Trivy, Hadolint, Gitleaks
```

### Important: This Is a Library

- There is **no `main` package**, no CLI, and no `make build` or `make run` targets.
- Changes must be validated by compiling (`go build ./...`) and running tests (`make test`).
- Any breaking change to exported types or interfaces affects downstream consumer projects.

## Architecture

The project provides two packages, both consumed as library imports:

### Repository Structure

```
cliforge/
├── platform/
│   ├── os.go              # OS interface: Download, Extract, Move, Remove, MakeExecutable
│   ├── os_unix.go         # OSUnix implementation (tar, mv, rm, chmod) -- build tag: !windows
│   ├── os_windows.go      # OSWindows implementation (PowerShell) -- build tag: windows
│   └── platform.go        # Info: normalizes runtime.GOOS/GOARCH (Android -> Linux mapping)
├── selfupdate/
│   ├── selfupdate.go      # Command: NewCommand(owner, repo, binary, version), Execute(dryRun, force)
│   ├── github.go          # fetchLatestRelease: GitHub API call, asset matching by {binary}-{version}-{os}-{arch}.{ext}
│   ├── version.go         # CompareVersions: semver comparison, "dev" always older, zero-padding
│   └── archive.go         # extractArchive: delegates to tar (Unix) or platform.OS.Extract (Windows)
├── test/
│   ├── doubles/
│   │   └── os_stub.go     # OSStub: stub implementing platform.OS with configurable errors
│   └── builders/
│       └── os_stub_builder.go  # OSStubBuilder: fluent builder for OSStub
├── Makefile               # Imports pipeline scripts (lint, test, sast)
├── go.mod                 # Module: github.com/rios0rios0/cliforge
└── .github/
    └── workflows/default.yaml  # CI/CD pipeline (delegates to rios0rios0/pipelines go-library workflow)
```

### Key Types

| Type                | Package      | Purpose                                                                              |
|---------------------|--------------|--------------------------------------------------------------------------------------|
| `OS`                | `platform`   | Interface: `Download`, `Extract`, `Move`, `Remove`, `MakeExecutable`                 |
| `OSUnix`            | `platform`   | Unix implementation via shell commands (`tar`, `mv`, `rm`, `chmod`)                  |
| `OSWindows`         | `platform`   | Windows implementation via PowerShell                                                |
| `Info`      | `platform`   | Normalizes `runtime.GOOS`/`runtime.GOARCH` (handles Android-to-Linux mapping)        |
| `Command` | `selfupdate` | Main public API: check for updates from GitHub releases, download, backup, replace   |
| `CompareVersions`   | `selfupdate` | Semver comparison; `"dev"` always older; pads unequal-length versions with zeros      |
| `GitHubRelease`     | `selfupdate` | JSON mapping for GitHub release API response                                         |

### Dependency Flow

```
Consumer CLI tool
  -> selfupdate.Command
       -> platform.OS (interface, injected per OS via build tags)
       -> selfupdate.CompareVersions (pure function)
       -> selfupdate.fetchLatestRelease (HTTP + JSON)
       -> logrus (structured logging)
```

## Testing

### Standards

- All tests follow **BDD** structure with `// given`, `// when`, `// then` comment blocks.
- Test descriptions use `"should ... when ..."` format via `t.Run()` subtests.
- Unit tests must run in **parallel** using `t.Parallel()` + `t.Run()`.
- All tests use `testify` (`assert`/`require`) -- never bare `t.Error`/`t.Fatal`.

### Test Infrastructure

`test/doubles/` contains stub implementations:

| Stub     | Implements    |
|----------|---------------|
| `OSStub` | `platform.OS` |

`test/builders/` provides builder-pattern helpers for constructing stubs in tests.

### Running Tests

```bash
make test             # Full test suite via pipeline scripts (ALWAYS use this)
go test ./...         # Quick compile + test check during development (acceptable)
```

## Validation

### After Making Changes

1. `go build ./...` -- must compile with zero errors
2. `make lint` -- must report 0 issues
3. `make test` -- all tests must pass
4. `make sast` -- should report no new findings

## Common Development Commands

```bash
# Full validation cycle
go build ./... && make lint && make test

# Quick test cycle during development
go test ./...

# Full security + quality gate
make lint && make test && make sast
```
