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

- **Self-Update**: Check for and install updates from GitHub Releases with dry-run, force, and interactive confirmation support
- **Platform Abstraction**: Cross-platform file operations (Unix/Windows) for download, extract, move, and permissions
- **Version Comparison**: Semantic version comparison with dev-build awareness

## Installation

```bash
go get github.com/rios0rios0/cliforge
```

## Usage

```go
import "github.com/rios0rios0/cliforge/selfupdate"

cmd := selfupdate.NewCommand("owner", "repo", "binary-name", currentVersion)
err := cmd.Execute(dryRun, force)
```

The self-update command expects release assets named `{binary}-{version}-{os}-{arch}.tar.gz` (`.zip` on Windows), which matches the GoReleaser default naming convention.

## Contributing

Contributions are welcome. See CONTRIBUTING.md for guidelines.

## License

See [LICENSE](LICENSE) file for details.
