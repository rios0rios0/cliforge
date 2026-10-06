package selfupdate

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rios0rios0/cliforge/pkg/platform"
	logger "github.com/sirupsen/logrus"
)

// Command checks for and applies updates from GitHub releases.
type Command struct {
	owner          string
	repo           string
	binaryName     string
	currentVersion string
	// apiBaseURL is the GitHub API the latest release is looked up on.
	apiBaseURL string
	// executable resolves the binary an update replaces.
	executable func() (string, error)
	// cacheDir resolves the directory the daily update check keeps its state under.
	cacheDir func() (string, error)
	// now tells the time; the daily update check compares calendar days.
	now func() time.Time
	// background runs the update check's lookup without holding up the command.
	background func(task func())
}

// NewCommand creates a new Command parameterized for a specific CLI tool.
func NewCommand(owner, repo, binaryName, currentVersion string) *Command {
	return &Command{
		owner:          owner,
		repo:           repo,
		binaryName:     binaryName,
		currentVersion: currentVersion,
		apiBaseURL:     githubAPIBaseURL,
		executable:     resolveExecutable,
		cacheDir:       os.UserCacheDir,
		now:            time.Now,
		background:     runInBackground,
	}
}

// Execute checks for updates and applies them if available.
func (it *Command) Execute(dryRun, force bool) error {
	logger.Infof("Checking for %s updates...", it.binaryName)
	logger.Infof("Current %s version: %s", it.binaryName, it.currentVersion)

	asset, err := fetchLatestRelease(it.apiBaseURL, it.owner, it.repo, it.binaryName)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	latestVersion := asset.version

	logger.Infof("Latest %s version: %s", it.binaryName, latestVersion)

	comparison := CompareVersions(it.currentVersion, latestVersion)
	switch {
	case comparison < 0:
		if dryRun {
			logger.Infof("Dry run: Would update %s from %s to %s", it.binaryName, it.currentVersion, latestVersion)
			logger.Infof("Download URL: %s", asset.url)
			return nil
		}

		if !force && !it.promptForUpdate(latestVersion) {
			logger.Info("Update cancelled by user")
			return nil
		}

		logger.Infof("Updating %s from %s to %s...", it.binaryName, it.currentVersion, latestVersion)
		return it.performUpdate(asset)

	case comparison == 0:
		logger.Infof("%s is already up to date", it.binaryName)
		return nil

	default:
		logger.Infof("Current %s version %s is newer than latest available %s",
			it.binaryName, it.currentVersion, latestVersion)
		return nil
	}
}

func (it *Command) promptForUpdate(latestVersion string) bool {
	logger.Infof("%s version %s is available (current: %s)", it.binaryName, latestVersion, it.currentVersion)
	logger.Info("Do you want to update? [y/N]: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			logger.Errorf("Error reading input: %v", err)
		}
		return false
	}
	response := strings.TrimSpace(strings.ToLower(scanner.Text()))

	return response == "y" || response == "yes"
}

// performUpdate downloads asset, checks it against every digest the release
// states for it, and installs the binary it holds in place of the running one.
// Nothing is downloaded when the release states no usable digest, and nothing is
// extracted from an archive that does not match.
func (it *Command) performUpdate(asset releaseAsset) error {
	currentOS := platform.GetOS()

	currentExe, err := it.executable()
	if err != nil {
		return err
	}

	digests, err := releaseDigests(asset)
	if err != nil {
		return fmt.Errorf("refusing to install %s: %w", asset.name, err)
	}

	tempDir, err := os.MkdirTemp("", fmt.Sprintf("%s-update-*", it.binaryName))
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			logger.Warnf("Failed to cleanup temp directory %s: %v", tempDir, removeErr)
		}
	}()

	tempArchive := filepath.Join(tempDir, fmt.Sprintf("%s-archive", it.binaryName))

	logger.Info("Downloading new version...")
	err = currentOS.Download(asset.url, tempArchive)
	if err != nil {
		return fmt.Errorf("failed to download new version: %w", err)
	}

	logger.Info("Verifying the download...")
	err = verifyArchive(tempArchive, digests)
	if err != nil {
		return fmt.Errorf("refusing to install %s: %w", asset.name, err)
	}

	logger.Info("Extracting archive...")
	err = extractArchive(tempArchive, tempDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	resolvedBinaryName := it.binaryName
	if platform.GetInfo().GetOSString() == windowsOS {
		resolvedBinaryName = it.binaryName + ".exe"
	}
	extractedBinary := filepath.Join(tempDir, resolvedBinaryName)
	if _, statErr := os.Stat(extractedBinary); errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("binary %q not found in extracted archive", resolvedBinaryName)
	}

	err = currentOS.MakeExecutable(extractedBinary)
	if err != nil {
		return fmt.Errorf("failed to make downloaded file executable: %w", err)
	}

	err = installBinary(currentOS, extractedBinary, currentExe)
	if err != nil {
		return err
	}

	logger.Infof("%s has been successfully updated!", it.binaryName)
	logger.Infof("Please restart your terminal or run '%s version' to verify the update", it.binaryName)

	return nil
}

// resolveExecutable returns the path of the running binary through any symlink:
// the file an update replaces.
func resolveExecutable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get current executable path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}
	return resolved, nil
}

// runInBackground runs task in a goroutine of its own.
func runInBackground(task func()) {
	go task()
}
