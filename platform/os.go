package platform

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	downloadTimeout = 10 * time.Minute
)

// OS defines the interface for platform-specific file operations.
type OS interface {
	Download(url, tempFilePath string) error
	Extract(tempFilePath, destPath string) error
	Move(tempFilePath, destPath string) error
	Remove(tempFilePath string) error
	MakeExecutable(filePath string) error
}

// DownloadFile provides a common implementation for downloading files via HTTP.
func DownloadFile(url, tempFilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: HTTP %d %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", tempFilePath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write downloaded content to file: %w", err)
	}

	return nil
}
