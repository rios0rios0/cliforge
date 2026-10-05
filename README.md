<h1 align="center">cliforge</h1>
<p align="center">
    <a href="https://github.com/rios0rios0/cliforge/releases/latest">
        <img src="https://img.shields.io/github/release/rios0rios0/cliforge.svg?style=for-the-badge&logo=github" alt="Latest Release"/></a>
    <a href="https://github.com/rios0rios0/cliforge/blob/main/LICENSE">
        <img src="https://img.shields.io/github/license/rios0rios0/cliforge.svg?style=for-the-badge&logo=github" alt="License"/></a>
    <a href="https://pkg.go.dev/github.com/rios0rios0/cliforge"><img src="https://img.shields.io/badge/go-reference-007d9c?style=for-the-badge&logo=go" alt="Go Reference"/></a>
</p>

Shared Go library providing self-update and platform abstraction for CLI tools that distribute binaries via GitHub Releases.

## Features

- **Self-Update**: Check for and install updates from GitHub Releases with dry-run, force, and interactive confirmation support. Passive `CheckForUpdates` startup checks look at most once a day, and a day only counts as checked once a lookup has answered, so a command that exits before its lookup returns leaves the check to the next one; no more than 5 lookups start in a day, and none at all when the state under the user's cache directory cannot be kept
- **Platform Abstraction**: Cross-platform file operations for download, extract, move, and permissions, in pure Go on Unix and Windows alike: moves fall back to a copy across volumes, and zip extraction refuses entries that would land outside the destination
- **Safe binary replacement**: the new release is staged beside the running binary and swapped in with renames, under a unique backup name, so an update succeeds on Windows even while an earlier release is still running
- **Version Comparison**: Semantic version comparison with dev-build awareness

## Installation

```bash
go get github.com/rios0rios0/cliforge
```

## Usage

```go
import "github.com/rios0rios0/cliforge/pkg/selfupdate"

cmd := selfupdate.NewCommand("owner", "repo", "binary-name", currentVersion)
err := cmd.Execute(dryRun, force)
```

The self-update command expects release assets named `{binary}-{version}-{os}-{arch}.tar.gz` (`.zip` on Windows), which matches the GoReleaser default naming convention.

## Contributing

Contributions are welcome. See CONTRIBUTING.md for guidelines.

## License

See [LICENSE](LICENSE) file for details.
