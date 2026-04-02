package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rios0rios0/cliforge/platform"
)

const (
	fetchTimeout     = 30 * time.Second
	githubAPIBaseURL = "https://api.github.com"
	windowsOS        = "windows"
)

// GitHubRelease represents a GitHub release response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// fetchLatestRelease fetches the latest release from GitHub and returns
// the version string, the download URL for the current platform, and any error.
func fetchLatestRelease(owner, repo, binaryName string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBaseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error fetching release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading response body: %w", err)
	}

	var release GitHubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", "", fmt.Errorf("error parsing release JSON: %w", err)
	}

	version := strings.TrimPrefix(release.TagName, "v")

	p := platform.GetInfo()
	ext := "tar.gz"
	if p.GetOSString() == windowsOS {
		ext = "zip"
	}
	expectedAssetName := fmt.Sprintf(
		"%s-%s-%s-%s.%s", binaryName, version, p.GetOSString(), p.GetArchString(), ext,
	)

	for _, asset := range release.Assets {
		if asset.Name == expectedAssetName {
			return version, asset.BrowserDownloadURL, nil
		}
	}

	return "", "", fmt.Errorf("no asset %q found for platform %s", expectedAssetName, p.GetPlatformString())
}
