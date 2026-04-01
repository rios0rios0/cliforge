//go:build !windows

package platform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	osOrwxGrxUx      = 0o755
	operationTimeout = 30 * time.Second
)

// OSUnix implements OS for Unix-like systems.
type OSUnix struct{}

func (it *OSUnix) Download(url, tempFilePath string) error {
	return DownloadFile(url, tempFilePath)
}

func (it *OSUnix) Extract(tempFilePath, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	unzipCmd := exec.CommandContext(ctx, "unzip", "-o", tempFilePath, "-d", destPath)
	unzipCmd.Stderr = os.Stderr
	unzipCmd.Stdout = os.Stdout
	err := unzipCmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to perform decompressing using 'zip': %w", err)
	}
	return err
}

func (it *OSUnix) Move(tempFilePath, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	mvCmd := exec.CommandContext(ctx, "mv", tempFilePath, destPath)
	err := mvCmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to perform moving folder using 'mv': %w", err)
	}
	return err
}

func (it *OSUnix) Remove(tempFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	rmCmd := exec.CommandContext(ctx, "rm", tempFilePath)
	err := rmCmd.Run()
	if err != nil {
		err = fmt.Errorf("failed to perform deleting folder using 'rm': %w", err)
	}
	return err
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
