//go:build !windows

package platform

import (
	"fmt"
	"os"
	"syscall"
)

const (
	osOrwxGrxUx = 0o755

	// errCrossDevice is the error a rename fails with when its two paths sit on
	// different filesystems, such as a tmpfs /tmp and the disk a binary lives on.
	errCrossDevice = syscall.EXDEV
)

// OSUnix implements OS for Unix-like systems.
type OSUnix struct{}

func (it *OSUnix) Download(url, tempFilePath string) error {
	return DownloadFile(url, tempFilePath)
}

func (it *OSUnix) Extract(tempFilePath, destPath string) error {
	if err := extractZip(tempFilePath, destPath); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	return nil
}

func (it *OSUnix) Move(tempFilePath, destPath string) error {
	if err := moveFile(tempFilePath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil
}

func (it *OSUnix) Remove(tempFilePath string) error {
	if err := os.Remove(tempFilePath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

func (it *OSUnix) MakeExecutable(filePath string) error {
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	err := os.Chmod(filePath, osOrwxGrxUx)
	if err != nil {
		err = fmt.Errorf("failed to perform change binary permissions using 'chmod': %w", err)
	}
	return err
}

// GetOS returns the platform-specific OS implementation.
func GetOS() *OSUnix {
	return &OSUnix{}
}
