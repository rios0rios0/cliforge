package selfupdate

import "github.com/rios0rios0/cliforge/pkg/platform"

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
