package platform_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/platform"
)

const (
	// sourceMode is the permission set MakeExecutable gives a binary on Unix.
	sourceMode = 0o755
	// ownerExecuteBit is the permission bit that lets the owner run a file.
	ownerExecuteBit = 0o100
)

// writeFile creates a file holding content in a fresh directory and returns its
// path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// identityOf captures the identity of the file at path through an open handle.
// On Windows [os.Stat] defers reading it until a comparison and then reads
// whatever sits at the path by then, which would compare a replaced file with
// itself.
func identityOf(t *testing.T, path string) os.FileInfo {
	t.Helper()
	file, err := os.Open(filepath.Clean(path))
	require.NoError(t, err)
	defer func() { require.NoError(t, file.Close()) }()
	info, err := file.Stat()
	require.NoError(t, err)
	return info
}

// crossingVolumes is a rename that fails the way one between two volumes does.
func crossingVolumes(oldPath, newPath string) error {
	return &os.LinkError{Op: "rename", Old: oldPath, New: newPath, Err: platform.ErrCrossDevice}
}

// readFile returns the content of the file at path.
func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return string(content)
}

// entriesIn lists the names in dir.
func entriesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestOSMove(t *testing.T) {
	t.Parallel()

	t.Run("should move the file when source and destination share a volume", func(t *testing.T) {
		t.Parallel()
		// given
		src := writeFile(t, "new", "new release")
		dst := filepath.Join(t.TempDir(), "binary")

		// when
		err := platform.GetOS().Move(src, dst)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, dst))
		assert.NoFileExists(t, src)
	})

	t.Run("should replace the destination when it already exists", func(t *testing.T) {
		t.Parallel()
		// given
		src := writeFile(t, "new", "new release")
		dst := writeFile(t, "binary", "old release")

		// when
		err := platform.GetOS().Move(src, dst)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, dst))
		assert.NoFileExists(t, src)
	})

	t.Run("should return an error when the source does not exist", func(t *testing.T) {
		t.Parallel()
		// given
		src := filepath.Join(t.TempDir(), "missing")
		dst := filepath.Join(t.TempDir(), "binary")

		// when
		err := platform.GetOS().Move(src, dst)

		// then
		require.Error(t, err)
		assert.NoFileExists(t, dst)
	})
}

func TestMoveFileAcrossVolumes(t *testing.T) {
	t.Parallel()

	t.Run("should copy the file when the rename crosses volumes", func(t *testing.T) {
		t.Parallel()
		// given
		src := writeFile(t, "new", "new release")
		dst := filepath.Join(t.TempDir(), "binary")

		// when
		err := platform.MoveFileWith(crossingVolumes, src, dst)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, dst))
		assert.NoFileExists(t, src)
	})

	t.Run("should keep the permissions and modification time when the rename crosses volumes", func(t *testing.T) {
		t.Parallel()
		// given
		src := writeFile(t, "new", "new release")
		require.NoError(t, platform.GetOS().MakeExecutable(src))
		modified := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		require.NoError(t, os.Chtimes(src, modified, modified))
		dst := filepath.Join(t.TempDir(), "binary")

		// when
		err := platform.MoveFileWith(crossingVolumes, src, dst)

		// then
		require.NoError(t, err)
		info, statErr := os.Stat(dst)
		require.NoError(t, statErr)
		assert.True(t, info.ModTime().Equal(modified), "modification time %s", info.ModTime())
		if runtime.GOOS != "windows" {
			assert.Equal(t, os.FileMode(sourceMode), info.Mode().Perm())
		}
	})

	t.Run("should leave a new file at the destination when the rename crosses volumes", func(t *testing.T) {
		t.Parallel()
		// given
		src := writeFile(t, "new", "new release")
		dst := writeFile(t, "binary", "old release")
		before := identityOf(t, dst)

		// when
		err := platform.MoveFileWith(crossingVolumes, src, dst)

		// then
		require.NoError(t, err)
		assert.False(t, os.SameFile(before, identityOf(t, dst)), "the old file must not be rewritten in place")
	})

	t.Run("should keep the source when the copy cannot be put in place", func(t *testing.T) {
		t.Parallel()
		// given -- a directory at the destination makes the final rename fail
		src := writeFile(t, "new", "new release")
		parent := t.TempDir()
		dst := filepath.Join(parent, "binary")
		// 0o700 is the tightest mode a directory can have while staying traversable;
		// semgrep applies its 0o600 file threshold to directories too.
		// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
		require.NoError(t, os.Mkdir(dst, 0o700))
		require.NoError(t, os.WriteFile(filepath.Join(dst, "keep"), []byte("x"), 0o600))

		// when
		err := platform.MoveFileWith(crossingVolumes, src, dst)

		// then
		require.Error(t, err)
		assert.Equal(t, "new release", readFile(t, src))
		assert.Equal(t, []string{"binary"}, entriesIn(t, parent), "no temporary copy may be left behind")
	})

	t.Run("should refuse to copy anything but a regular file across volumes", func(t *testing.T) {
		t.Parallel()
		// given
		src := t.TempDir()
		dst := filepath.Join(t.TempDir(), "binary")

		// when
		err := platform.MoveFileWith(crossingVolumes, src, dst)

		// then
		require.ErrorContains(t, err, "not a regular file")
		assert.NoFileExists(t, dst)
	})

	t.Run("should not copy when the rename fails for another reason", func(t *testing.T) {
		t.Parallel()
		// given
		denied := errors.New("fixture permission failure")
		src := writeFile(t, "new", "new release")
		dst := filepath.Join(t.TempDir(), "binary")

		// when
		err := platform.MoveFileWith(func(_, _ string) error { return denied }, src, dst)

		// then
		require.ErrorIs(t, err, denied)
		assert.NoFileExists(t, dst)
		assert.FileExists(t, src)
	})
}

func TestOSRemove(t *testing.T) {
	t.Parallel()

	t.Run("should remove the file when it exists", func(t *testing.T) {
		t.Parallel()
		// given
		path := writeFile(t, "binary.backup", "old release")

		// when
		err := platform.GetOS().Remove(path)

		// then
		require.NoError(t, err)
		assert.NoFileExists(t, path)
	})

	t.Run("should return an error when the file does not exist", func(t *testing.T) {
		t.Parallel()
		// given
		path := filepath.Join(t.TempDir(), "missing")

		// when
		err := platform.GetOS().Remove(path)

		// then
		require.Error(t, err)
	})
}

func TestOSMakeExecutable(t *testing.T) {
	t.Parallel()

	t.Run("should leave the file executable when it was not", func(t *testing.T) {
		t.Parallel()
		// given
		path := writeFile(t, "binary", "release")

		// when
		err := platform.GetOS().MakeExecutable(path)

		// then
		require.NoError(t, err)
		if runtime.GOOS != "windows" {
			info, statErr := os.Stat(path)
			require.NoError(t, statErr)
			assert.NotZero(t, info.Mode().Perm()&ownerExecuteBit, "the owner must be able to run it")
		}
	})
}
