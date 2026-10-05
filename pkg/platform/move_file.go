package platform

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// moveFile moves src to dst, replacing dst when it exists. It renames when it can
// and copies only when the two paths sit on different volumes, which a rename
// cannot cross.
func moveFile(src, dst string) error {
	return moveFileWith(os.Rename, src, dst)
}

// moveFileWith is moveFile over the given rename, so that tests can stand in for
// a rename that crosses volumes.
func moveFileWith(rename func(oldPath, newPath string) error, src, dst string) error {
	err := rename(src, dst)
	if err == nil || !isCrossDevice(err) {
		return err
	}
	return copyAcrossVolumes(src, dst)
}

// isCrossDevice reports whether a rename failed because its two paths sit on
// different volumes.
func isCrossDevice(err error) bool {
	return errors.Is(err, errCrossDevice)
}

// copyAcrossVolumes moves src to dst when they sit on different volumes.
//
// The copy is written to a temporary file beside dst and renamed over it, so dst
// only ever holds a complete file, and always a new one. Rewriting the file
// already at dst would keep its identity, which consumers compare to tell
// whether an update installed anything, and fails outright on Linux while that
// file is a running binary. src keeps its permissions and modification time, as
// it would through a rename, and is removed only once dst is in place.
func copyAcrossVolumes(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("cannot copy %s across volumes: not a regular file", src)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create a temporary file beside %s: %w", dst, err)
	}
	tmpPath := tmp.Name()
	if err = publishCopy(tmp, src, info, dst); err != nil {
		if removeErr := os.Remove(tmpPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return errors.Join(err, removeErr)
		}
		return err
	}
	return os.Remove(src)
}

// publishCopy fills tmp with the content of src, gives it src's permissions and
// modification time, and renames it to dst.
func publishCopy(tmp *os.File, src string, info os.FileInfo, dst string) error {
	if err := fillFrom(tmp, src); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), info.Mode().Perm()); err != nil {
		return err
	}
	if err := os.Chtimes(tmp.Name(), time.Time{}, info.ModTime()); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dst)
}

// fillFrom copies the content of the file at src into dst, flushes it to disk and
// closes dst.
func fillFrom(dst *os.File, src string) error {
	in, err := os.Open(filepath.Clean(src))
	if err != nil {
		return errors.Join(err, dst.Close())
	}
	defer func() { _ = in.Close() }()

	if _, err = io.Copy(dst, in); err != nil {
		return errors.Join(err, dst.Close())
	}
	if err = dst.Sync(); err != nil {
		return errors.Join(err, dst.Close())
	}
	return dst.Close()
}
