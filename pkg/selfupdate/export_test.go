package selfupdate

import (
	"time"

	"github.com/rios0rios0/cliforge/pkg/platform"
)

const (
	// UpdateCheckMarkerFilename marks the day a lookup last answered.
	UpdateCheckMarkerFilename = updateCheckMarkerFilename
	// UpdateCheckAttemptsFilename counts the lookups started today.
	UpdateCheckAttemptsFilename = updateCheckAttemptsFilename
	// MaxUpdateCheckAttemptsPerDay bounds the lookups a day can start.
	MaxUpdateCheckAttemptsPerDay = maxUpdateCheckAttemptsPerDay
)

// TestOption adjusts a Command built by NewTestCommand.
type TestOption func(*Command)

// NewTestCommand builds a Command the way NewCommand does and then applies opts,
// so that tests can point it at a fake release server and a stand-in binary.
func NewTestCommand(owner, repo, binaryName, currentVersion string, opts ...TestOption) *Command {
	command := NewCommand(owner, repo, binaryName, currentVersion)
	for _, opt := range opts {
		opt(command)
	}
	return command
}

// WithAPIBaseURL looks the latest release up on apiBaseURL instead of the GitHub
// API.
func WithAPIBaseURL(apiBaseURL string) TestOption {
	return func(command *Command) { command.apiBaseURL = apiBaseURL }
}

// WithExecutable makes path the binary an update replaces.
func WithExecutable(path string) TestOption {
	return func(command *Command) {
		command.executable = func() (string, error) { return path, nil }
	}
}

// InstallBinary puts newBinary in place of currentExe through system.
func InstallBinary(system platform.OS, newBinary, currentExe string) error {
	return installBinary(system, newBinary, currentExe)
}

// ReleaseAssetName is the name of the archive a release of binaryName at version
// ships for the running platform.
func ReleaseAssetName(binaryName, version string) string {
	return releaseAssetName(binaryName, version, platform.GetInfo())
}

// WithCacheDir keeps the update check's state under dir.
func WithCacheDir(dir string) TestOption {
	return func(command *Command) {
		command.cacheDir = func() (string, error) { return dir, nil }
	}
}

// WithoutCacheDir makes the user cache directory unresolvable, failing with err.
func WithoutCacheDir(err error) TestOption {
	return func(command *Command) {
		command.cacheDir = func() (string, error) { return "", err }
	}
}

// WithClock fixes the time the update check sees at now.
func WithClock(now time.Time) TestOption {
	return func(command *Command) {
		command.now = func() time.Time { return now }
	}
}

// WithSynchronousLookups runs the update check's lookup before CheckForUpdates
// returns.
func WithSynchronousLookups() TestOption {
	return func(command *Command) {
		command.background = func(task func()) { task() }
	}
}

// WithAbandonedLookups starts no lookup at all, the way a command that exits
// before its lookup returns leaves it unfinished.
func WithAbandonedLookups() TestOption {
	return func(command *Command) {
		command.background = func(func()) {}
	}
}
