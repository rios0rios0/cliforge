package platform

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// errCrossDevice is the error a rename fails with when its two paths sit on
// different volumes. The syscall package's EXDEV is a made-up value on Windows
// that no rename ever returns.
const errCrossDevice = windows.ERROR_NOT_SAME_DEVICE

// OSWindows implements OS for Windows systems.
type OSWindows struct{}

func (it *OSWindows) Download(url, tempFilePath string) error {
	return DownloadFile(url, tempFilePath)
}

func (it *OSWindows) Extract(tempFilePath, destPath string) error {
	if err := extractZip(tempFilePath, destPath); err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}
	return nil
}

func (it *OSWindows) Move(tempFilePath, destPath string) error {
	if err := moveFile(tempFilePath, destPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}
	return nil
}

func (it *OSWindows) Remove(tempFilePath string) error {
	if err := os.Remove(tempFilePath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}
	return nil
}

func (it *OSWindows) MakeExecutable(_ string) error {
	return nil // Windows doesn't need to explicitly make files executable
}

// GetOS returns the platform-specific OS implementation.
func GetOS() *OSWindows {
	return &OSWindows{}
}
