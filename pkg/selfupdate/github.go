package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rios0rios0/cliforge/pkg/platform"
)

const (
	fetchTimeout     = 30 * time.Second
	githubAPIBaseURL = "https://api.github.com"
	windowsOS        = "windows"
	// maxChecksumsSize bounds the checksums.txt read into memory. GoReleaser
	// writes one line per artifact, so a real one is a few hundred bytes.
	maxChecksumsSize = 1 << 20
)

// GitHubRelease represents a GitHub release response.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		// Digest is the "<algorithm>:<hex>" digest GitHub computed when the asset
		// was uploaded, empty when it reports none.
		Digest string `json:"digest"`
	} `json:"assets"`
}

// releaseAsset is the archive the latest release ships for the running platform,
// together with what the release states about its content.
type releaseAsset struct {
	version string
	name    string
	url     string
	// digest is the digest GitHub reports for the archive, empty when it reports
	// none.
	digest string
	// checksumsURL locates the checksums.txt published beside the archive, empty
	// when the release has none.
	checksumsURL string
}

// fetchGitHubRelease fetches the latest release metadata from the GitHub API at
// apiBaseURL.
func fetchGitHubRelease(apiBaseURL, owner, repo string) (*GitHubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", apiBaseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	var release GitHubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return nil, fmt.Errorf("error parsing release JSON: %w", err)
	}

	return &release, nil
}

// fetchLatestVersion fetches only the latest release version from GitHub,
// without requiring a platform-specific asset to exist.
func fetchLatestVersion(apiBaseURL, owner, repo string) (string, error) {
	release, err := fetchGitHubRelease(apiBaseURL, owner, repo)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// fetchLatestRelease fetches the latest release from GitHub and returns the
// archive it ships for the current platform.
func fetchLatestRelease(apiBaseURL, owner, repo, binaryName string) (releaseAsset, error) {
	release, err := fetchGitHubRelease(apiBaseURL, owner, repo)
	if err != nil {
		return releaseAsset{}, err
	}

	version := strings.TrimPrefix(release.TagName, "v")

	p := platform.GetInfo()
	asset := releaseAsset{version: version, name: releaseAssetName(binaryName, version, p)}

	found := false
	for _, candidate := range release.Assets {
		switch candidate.Name {
		case asset.name:
			asset.url = candidate.BrowserDownloadURL
			asset.digest = candidate.Digest
			found = true
		case checksumsAssetName:
			asset.checksumsURL = candidate.BrowserDownloadURL
		}
	}
	if !found {
		return releaseAsset{}, fmt.Errorf("no asset %q found for platform %s", asset.name, p.GetPlatformString())
	}

	return asset, nil
}

// fetchChecksums downloads the checksums.txt at url.
func fetchChecksums(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error downloading %s: %w", checksumsAssetName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s returned status %d", checksumsAssetName, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxChecksumsSize+1))
	if err != nil {
		return "", fmt.Errorf("error reading %s: %w", checksumsAssetName, err)
	}
	if len(body) > maxChecksumsSize {
		return "", fmt.Errorf("%s is larger than %d bytes", checksumsAssetName, maxChecksumsSize)
	}

	return string(body), nil
}

// releaseAssetName is the name GoReleaser gives the archive of binaryName at
// version for platform p: {binary}-{version}-{os}-{arch}.tar.gz, or .zip on
// Windows.
func releaseAssetName(binaryName, version string, p platform.Info) string {
	ext := "tar.gz"
	if p.GetOSString() == windowsOS {
		ext = "zip"
	}
	return fmt.Sprintf("%s-%s-%s-%s.%s", binaryName, version, p.GetOSString(), p.GetArchString(), ext)
}
