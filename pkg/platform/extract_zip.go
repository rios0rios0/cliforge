package platform

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// maxExtractedSize caps how much an archive may unpack to. A release holds
	// one binary, so anything larger is a corrupt or hostile archive.
	maxExtractedSize int64 = 512 << 20
	// extractedFileMode leaves extracted files private; the self-update makes the
	// binary executable afterwards.
	extractedFileMode = 0o600
	// extractedDirMode is the tightest mode a directory can have and still be
	// traversed.
	extractedDirMode = 0o700
)

// errArchiveTooLarge reports an archive that unpacks to more than the limit.
var errArchiveTooLarge = errors.New("archive unpacks to more than the size limit")

// extractZip unpacks the zip archive at archivePath into destDir.
func extractZip(archivePath, destDir string) error {
	return extractZipWithLimit(archivePath, destDir, maxExtractedSize)
}

// extractZipWithLimit is extractZip with the total unpacked size capped at limit.
//
// Every entry is written through an [os.Root] opened on destDir, with its name
// converted by [filepath.Localize], so no entry can land outside destDir: names
// that climb out with "..", absolute names and names Windows cannot represent are
// all refused, and so are symbolic links and anything else that is not a regular
// file or a directory.
func extractZipWithLimit(archivePath, destDir string, limit int64) error {
	reader, err := zip.OpenReader(filepath.Clean(archivePath))
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = reader.Close() }()

	root, err := os.OpenRoot(destDir)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", destDir, err)
	}
	defer func() { _ = root.Close() }()

	remaining := limit
	for _, entry := range reader.File {
		written, entryErr := extractZipEntry(root, entry, remaining)
		if entryErr != nil {
			return fmt.Errorf("failed to extract %q: %w", entry.Name, entryErr)
		}
		remaining -= written
	}
	return nil
}

// extractZipEntry writes one archive entry under root and returns how many bytes
// it wrote, refusing to write more than remaining.
func extractZipEntry(root *os.Root, entry *zip.File, remaining int64) (int64, error) {
	name, err := filepath.Localize(strings.TrimSuffix(entry.Name, "/"))
	if err != nil {
		return 0, err
	}
	mode := entry.Mode()
	if mode.IsDir() {
		return 0, root.MkdirAll(name, extractedDirMode)
	}
	if !mode.IsRegular() {
		return 0, fmt.Errorf("unsupported entry type %s", mode.Type())
	}
	if dir := filepath.Dir(name); dir != "." {
		if err = root.MkdirAll(dir, extractedDirMode); err != nil {
			return 0, err
		}
	}

	content, err := entry.Open()
	if err != nil {
		return 0, err
	}
	defer func() { _ = content.Close() }()

	out, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, extractedFileMode)
	if err != nil {
		return 0, err
	}
	// Copying one byte past the limit is how an entry that is too large shows up.
	written, err := io.CopyN(out, content, remaining+1)
	closeErr := out.Close()
	if err != nil && !errors.Is(err, io.EOF) {
		return written, errors.Join(err, closeErr)
	}
	if written > remaining {
		return written, errArchiveTooLarge
	}
	return written, closeErr
}
