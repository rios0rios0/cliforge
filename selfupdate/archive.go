package selfupdate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/rios0rios0/cliforge/platform"
)

const (
	extractTimeout = 30 * time.Second
)

// extractArchive extracts the downloaded archive to the destination directory.
func extractArchive(archivePath, destDir string) error {
	p := platform.GetInfo()

	// On Windows, use the OS abstraction (PowerShell Expand-Archive) for .zip files
	if p.GetOSString() == windowsOS {
		currentOS := platform.GetOS()
		return currentOS.Extract(archivePath, destDir)
	}

	// On Unix, extract .tar.gz using tar
	ctx, cancel := context.WithTimeout(context.Background(), extractTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tar", "-xzf", archivePath, "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract tar.gz archive: %w", err)
	}
	return nil
}
