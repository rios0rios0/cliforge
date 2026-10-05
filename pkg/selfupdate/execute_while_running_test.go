package selfupdate_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/platform"
	"github.com/rios0rios0/cliforge/pkg/selfupdate"
)

// runningRelease copies this test binary to where a release would be installed,
// starts it, and returns its path: a binary that is running while it is updated.
// The process is killed when the test ends.
func runningRelease(t *testing.T) string {
	t.Helper()
	self, err := os.Executable()
	require.NoError(t, err)
	content, err := os.ReadFile(filepath.Clean(self))
	require.NoError(t, err)
	binary := filepath.Join(t.TempDir(), binaryFileName())
	require.NoError(t, os.WriteFile(binary, content, 0o600))
	require.NoError(t, platform.GetOS().MakeExecutable(binary))

	process := exec.CommandContext(t.Context(), binary)
	process.Env = append(os.Environ(), runAsReleaseEnv+"=1")
	require.NoError(t, process.Start())
	t.Cleanup(func() {
		_ = process.Process.Kill()
		_ = process.Wait()
	})
	return binary
}

// update runs a forced update of binary, built as version, to the release server
// publishes.
func update(t *testing.T, binary, version string, rel release) error {
	t.Helper()
	server := newReleaseServer(t, rel)
	return selfupdate.NewTestCommand(owner, repo, binaryName, version,
		selfupdate.WithAPIBaseURL(server.URL),
		selfupdate.WithExecutable(binary),
	).Execute(false, true)
}

// TestExecuteWhileRunning is deliberately not parallel. It writes binaries and
// then starts them, and on Linux a start fails with ETXTBSY if another test forks
// a process while such a binary is still open for writing.
func TestExecuteWhileRunning(t *testing.T) {
	t.Run("should replace the binary while it is running", func(t *testing.T) {
		// given
		binary := runningRelease(t)

		// when
		err := update(t, binary, currentVersion, release{
			version: latestVersion,
			archive: releaseArchive(t, binaryFileName(), "release 1.1.0"),
		})

		// then
		require.NoError(t, err)
		assert.Equal(t, "release 1.1.0", readFile(t, binary))
	})

	t.Run("should install a later release while an earlier one still runs from its backup", func(t *testing.T) {
		// given -- after this first update the running process is the backup, which
		// Windows can neither delete nor replace until that process exits
		binary := runningRelease(t)
		require.NoError(t, update(t, binary, currentVersion, release{
			version: latestVersion,
			archive: releaseArchive(t, binaryFileName(), "release 1.1.0"),
		}))

		// when
		err := update(t, binary, latestVersion, release{
			version: "1.2.0",
			archive: releaseArchive(t, binaryFileName(), "release 1.2.0"),
		})

		// then
		require.NoError(t, err)
		assert.Equal(t, "release 1.2.0", readFile(t, binary))
	})
}
