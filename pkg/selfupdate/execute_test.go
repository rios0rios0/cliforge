package selfupdate_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/selfupdate"
)

// commandFor builds a command that updates binary from server.
func commandFor(server *releaseServer, binary string) *selfupdate.Command {
	return selfupdate.NewTestCommand(owner, repo, binaryName, currentVersion,
		selfupdate.WithAPIBaseURL(server.URL),
		selfupdate.WithExecutable(binary),
	)
}

func TestExecute(t *testing.T) {
	t.Parallel()

	t.Run("should replace the binary with the latest release when a newer version is published", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		before := identityOf(t, binary)
		server := newReleaseServer(t, release{
			version: latestVersion,
			archive: releaseArchive(t, binaryFileName(), "new release"),
		})

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, binary))
		assert.False(t, os.SameFile(before, identityOf(t, binary)), "the release must arrive as a new file")
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(binary)),
			"no staged binary or backup may be left behind")
	})

	t.Run("should leave the binary untouched when the archive does not contain it", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		server := newReleaseServer(t, release{
			version: latestVersion,
			archive: releaseArchive(t, "README.md", "not the binary"),
		})

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "not found in extracted archive")
		assert.Equal(t, "old release", readFile(t, binary))
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(binary)))
	})

	t.Run("should leave the binary untouched when the download fails", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		server := newReleaseServer(t, release{version: latestVersion, assetStatus: http.StatusBadGateway})

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "failed to download")
		assert.Equal(t, "old release", readFile(t, binary))
	})

	t.Run("should not download when the binary is already up to date", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "current release")
		server := newReleaseServer(t, release{version: currentVersion})

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, int32(1), server.lookups.Load())
		assert.Zero(t, server.downloads.Load())
		assert.Zero(t, server.checksumFetches.Load())
		assert.Equal(t, "current release", readFile(t, binary))
	})

	t.Run("should not download when running dry", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		server := newReleaseServer(t, release{version: latestVersion})

		// when
		err := commandFor(server, binary).Execute(true, false)

		// then
		require.NoError(t, err)
		assert.Zero(t, server.downloads.Load())
		assert.Zero(t, server.checksumFetches.Load())
		assert.Equal(t, "old release", readFile(t, binary))
	})

	t.Run("should return an error when the release has no asset for this platform", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		server := newReleaseServer(t, release{version: latestVersion, assetName: "tool-1.1.0-plan9-mips.tar.gz"})

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "no asset")
		assert.Zero(t, server.downloads.Load())
	})

	t.Run("should return an error when the latest release cannot be fetched", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		server := newReleaseServer(t, release{latestStatus: http.StatusForbidden})

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "failed to fetch latest release")
		assert.Equal(t, "old release", readFile(t, binary))
	})
}
