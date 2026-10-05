package platform_test

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/platform"
)

// zipEntry is one file, directory or link to put in a test archive.
type zipEntry struct {
	name    string
	content string
	mode    fs.FileMode
}

// writeZip builds a zip archive of entries at path.
func writeZip(t *testing.T, path string, entries ...zipEntry) {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		mode := entry.mode
		if mode == 0 {
			mode = 0o644
		}
		header.SetMode(mode)
		writer, err := archive.CreateHeader(header)
		require.NoError(t, err)
		_, err = writer.Write([]byte(entry.content))
		require.NoError(t, err)
	}
	require.NoError(t, archive.Close())
	require.NoError(t, os.WriteFile(path, buffer.Bytes(), 0o600))
}

func TestOSExtract(t *testing.T) {
	t.Parallel()

	t.Run("should extract every file when the archive is valid", func(t *testing.T) {
		t.Parallel()
		// given
		archive := filepath.Join(t.TempDir(), "release.zip")
		writeZip(t, archive,
			zipEntry{name: "tool.exe", content: "binary"},
			zipEntry{name: "docs/", mode: fs.ModeDir | 0o755},
			zipEntry{name: "docs/README.md", content: "readme"},
		)
		dest := t.TempDir()

		// when
		err := platform.GetOS().Extract(archive, dest)

		// then
		require.NoError(t, err)
		assert.Equal(t, "binary", readFile(t, filepath.Join(dest, "tool.exe")))
		assert.Equal(t, "readme", readFile(t, filepath.Join(dest, "docs", "README.md")))
	})

	t.Run("should extract the archive when its file name has no extension", func(t *testing.T) {
		t.Parallel()
		// given -- the self-update downloads every release to "<binary>-archive"
		archive := filepath.Join(t.TempDir(), "tool-archive")
		writeZip(t, archive, zipEntry{name: "tool.exe", content: "binary"})
		dest := t.TempDir()

		// when
		err := platform.GetOS().Extract(archive, dest)

		// then
		require.NoError(t, err)
		assert.Equal(t, "binary", readFile(t, filepath.Join(dest, "tool.exe")))
	})

	t.Run("should create the parent directories of a file when the archive lists none", func(t *testing.T) {
		t.Parallel()
		// given
		archive := filepath.Join(t.TempDir(), "release.zip")
		writeZip(t, archive, zipEntry{name: "nested/deeper/tool.exe", content: "binary"})
		dest := t.TempDir()

		// when
		err := platform.GetOS().Extract(archive, dest)

		// then
		require.NoError(t, err)
		assert.Equal(t, "binary", readFile(t, filepath.Join(dest, "nested", "deeper", "tool.exe")))
	})

	t.Run("should return an error when the file is not a zip archive", func(t *testing.T) {
		t.Parallel()
		// given
		archive := writeFile(t, "release.zip", "not a zip")

		// when
		err := platform.GetOS().Extract(archive, t.TempDir())

		// then
		require.Error(t, err)
	})
}

func TestOSExtractRejectsUnsafeArchives(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		entry zipEntry
	}{
		{name: "an entry climbs out of the destination", entry: zipEntry{name: "../escaped.txt", content: "x"}},
		{name: "an entry climbs out from a subdirectory", entry: zipEntry{name: "docs/../../escaped.txt", content: "x"}},
		{name: "an entry has an absolute path", entry: zipEntry{name: "/escaped.txt", content: "x"}},
		{name: "an entry is a symbolic link", entry: zipEntry{name: "link", content: "/etc/passwd", mode: fs.ModeSymlink | 0o777}},
	} {
		t.Run("should reject the archive when "+tc.name, func(t *testing.T) {
			t.Parallel()
			// given
			base := t.TempDir()
			archive := filepath.Join(base, "release.zip")
			writeZip(t, archive, tc.entry)
			dest := filepath.Join(base, "dest")
			// 0o700 is the tightest mode a directory can have while staying traversable;
			// semgrep applies its 0o600 file threshold to directories too.
			// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
			require.NoError(t, os.Mkdir(dest, 0o700))

			// when
			err := platform.GetOS().Extract(archive, dest)

			// then
			require.Error(t, err)
			assert.Empty(t, entriesIn(t, dest))
			assert.ElementsMatch(t, []string{"release.zip", "dest"}, entriesIn(t, base))
		})
	}
}

func TestExtractZipWithLimit(t *testing.T) {
	t.Parallel()

	t.Run("should reject the archive when its content exceeds the size limit", func(t *testing.T) {
		t.Parallel()
		// given
		const limit = 10
		archive := filepath.Join(t.TempDir(), "release.zip")
		writeZip(t, archive, zipEntry{name: "tool.exe", content: "eleven char"})

		// when
		err := platform.ExtractZipWithLimit(archive, t.TempDir(), limit)

		// then
		require.ErrorContains(t, err, "size limit")
	})

	t.Run("should extract the archive when its content fills the size limit exactly", func(t *testing.T) {
		t.Parallel()
		// given
		const limit = 10
		archive := filepath.Join(t.TempDir(), "release.zip")
		writeZip(t, archive, zipEntry{name: "a", content: "12345"}, zipEntry{name: "b", content: "67890"})
		dest := t.TempDir()

		// when
		err := platform.ExtractZipWithLimit(archive, dest, limit)

		// then
		require.NoError(t, err)
		assert.Equal(t, "67890", readFile(t, filepath.Join(dest, "b")))
	})
}
