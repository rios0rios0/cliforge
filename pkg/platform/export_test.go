package platform

// ErrCrossDevice is the error a rename fails with on this platform when its two
// paths sit on different volumes.
const ErrCrossDevice = errCrossDevice

// MoveFileWith moves src to dst over the given rename, so tests can make the
// rename cross volumes.
func MoveFileWith(rename func(oldPath, newPath string) error, src, dst string) error {
	return moveFileWith(rename, src, dst)
}

// ExtractZipWithLimit unpacks the archive with its total unpacked size capped at
// limit.
func ExtractZipWithLimit(archivePath, destDir string, limit int64) error {
	return extractZipWithLimit(archivePath, destDir, limit)
}
