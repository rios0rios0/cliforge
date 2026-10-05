package selfupdate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/platform"
	"github.com/rios0rios0/cliforge/pkg/selfupdate"
	"github.com/rios0rios0/cliforge/pkg/test/builders"
	"github.com/rios0rios0/cliforge/pkg/test/doubles"
)

// faultyOS builds the OSFaultStub the builder describes.
func faultyOS(t *testing.T, builder *builders.OSFaultStubBuilder) platform.OS {
	t.Helper()
	stub, ok := builder.Build().(*doubles.OSFaultStub)
	require.True(t, ok)
	return stub
}

// newRelease creates the extracted binary of a release, holding "new release",
// in a directory of its own.
func newRelease(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), binaryFileName())
	require.NoError(t, os.WriteFile(path, []byte("new release"), 0o600))
	return path
}

// isBackup reports whether path names a backup an update made.
func isBackup(path string) bool {
	return strings.Contains(filepath.Base(path), ".backup-")
}

// backupsIn lists the backups beside binary.
func backupsIn(t *testing.T, binary string) []string {
	t.Helper()
	var backups []string
	for _, name := range entriesIn(t, filepath.Dir(binary)) {
		if isBackup(name) {
			backups = append(backups, name)
		}
	}
	return backups
}

func TestInstallBinary(t *testing.T) {
	t.Parallel()

	t.Run("should swap in the new binary and remove the backup when every step succeeds", func(t *testing.T) {
		t.Parallel()
		// given
		current := installedBinary(t, "old release")
		incoming := newRelease(t)

		// when
		err := selfupdate.InstallBinary(platform.GetOS(), incoming, current)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, current))
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(current)))
	})

	t.Run("should leave the current binary untouched when the new binary cannot be staged", func(t *testing.T) {
		t.Parallel()
		// given
		current := installedBinary(t, "old release")
		incoming := newRelease(t)
		system := faultyOS(t, builders.NewOSFaultStubBuilder().
			WithFailingMove(func(_, dst string) bool { return strings.HasSuffix(dst, ".new") }))

		// when
		err := selfupdate.InstallBinary(system, incoming, current)

		// then
		require.ErrorContains(t, err, "failed to stage new binary")
		assert.Equal(t, "old release", readFile(t, current))
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(current)))
	})

	t.Run("should leave the current binary untouched when it cannot be backed up", func(t *testing.T) {
		t.Parallel()
		// given
		current := installedBinary(t, "old release")
		incoming := newRelease(t)
		system := faultyOS(t, builders.NewOSFaultStubBuilder().
			WithFailingMove(func(src, _ string) bool { return src == current }))

		// when
		err := selfupdate.InstallBinary(system, incoming, current)

		// then
		require.ErrorContains(t, err, "failed to backup current binary")
		assert.Equal(t, "old release", readFile(t, current))
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(current)),
			"the staged binary must be discarded")
	})

	t.Run("should restore the current binary when the new binary cannot be installed", func(t *testing.T) {
		t.Parallel()
		// given
		current := installedBinary(t, "old release")
		incoming := newRelease(t)
		system := faultyOS(t, builders.NewOSFaultStubBuilder().
			WithFailingMove(func(src, dst string) bool { return strings.HasSuffix(src, ".new") && dst == current }))

		// when
		err := selfupdate.InstallBinary(system, incoming, current)

		// then
		require.ErrorContains(t, err, "failed to install new binary")
		assert.Equal(t, "old release", readFile(t, current))
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(current)),
			"the staged binary must be discarded and the backup moved back")
	})

	t.Run("should succeed and keep the backup when the backup cannot be removed", func(t *testing.T) {
		t.Parallel()
		// given -- as on Windows, where the backup is the image of a running process
		current := installedBinary(t, "old release")
		incoming := newRelease(t)
		system := faultyOS(t, builders.NewOSFaultStubBuilder().WithFailingRemove(isBackup))

		// when
		err := selfupdate.InstallBinary(system, incoming, current)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, current))
		backups := backupsIn(t, current)
		require.Len(t, backups, 1)
		assert.Equal(t, "old release", readFile(t, filepath.Join(filepath.Dir(current), backups[0])))
	})
}

func TestInstallBinaryWithEarlierBackups(t *testing.T) {
	t.Parallel()

	t.Run("should remove the backups earlier updates left behind", func(t *testing.T) {
		t.Parallel()
		// given
		current := installedBinary(t, "old release")
		for _, name := range []string{binaryFileName() + ".backup", binaryFileName() + ".backup-1700000000"} {
			require.NoError(t, os.WriteFile(filepath.Join(filepath.Dir(current), name), []byte("older"), 0o600))
		}
		incoming := newRelease(t)

		// when
		err := selfupdate.InstallBinary(platform.GetOS(), incoming, current)

		// then
		require.NoError(t, err)
		assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(current)))
	})

	t.Run("should install the release when a backup an earlier update left cannot be removed", func(t *testing.T) {
		t.Parallel()
		// given -- the backup is still running a binary, so Windows keeps it
		current := installedBinary(t, "old release")
		running := filepath.Join(filepath.Dir(current), binaryFileName()+".backup-1700000000")
		require.NoError(t, os.WriteFile(running, []byte("still running"), 0o600))
		incoming := newRelease(t)
		system := faultyOS(t, builders.NewOSFaultStubBuilder().
			WithFailingRemove(func(path string) bool { return path == running }).
			WithFailingMove(func(_, dst string) bool { return dst == running }))

		// when
		err := selfupdate.InstallBinary(system, incoming, current)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, current))
		assert.Equal(t, "still running", readFile(t, running))
	})

	t.Run("should leave files that only resemble backups untouched", func(t *testing.T) {
		t.Parallel()
		// given
		current := installedBinary(t, "old release")
		lookalikes := []string{
			binaryFileName() + ".backup-old",
			binaryFileName() + ".backup-",
			binaryFileName() + ".backupx",
			"other" + ".backup-1700000000",
		}
		for _, name := range lookalikes {
			require.NoError(t, os.WriteFile(filepath.Join(filepath.Dir(current), name), []byte("keep"), 0o600))
		}
		incoming := newRelease(t)

		// when
		err := selfupdate.InstallBinary(platform.GetOS(), incoming, current)

		// then
		require.NoError(t, err)
		assert.ElementsMatch(t, append([]string{binaryFileName()}, lookalikes...), entriesIn(t, filepath.Dir(current)))
	})
}
