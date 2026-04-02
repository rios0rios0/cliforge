package selfupdate

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rios0rios0/cliforge/pkg/platform"
	logger "github.com/sirupsen/logrus"
)

// Command checks for and applies updates from GitHub releases.
type Command struct {
	owner          string
	repo           string
	binaryName     string
	currentVersion string
}

// NewCommand creates a new Command parameterized for a specific CLI tool.
func NewCommand(owner, repo, binaryName, currentVersion string) *Command {
	return &Command{
		owner:          owner,
		repo:           repo,
		binaryName:     binaryName,
		currentVersion: currentVersion,
	}
}

// Execute checks for updates and applies them if available.
func (it *Command) Execute(dryRun, force bool) error {
	logger.Infof("Checking for %s updates...", it.binaryName)
	logger.Infof("Current %s version: %s", it.binaryName, it.currentVersion)

	latestVersion, downloadURL, err := fetchLatestRelease(it.owner, it.repo, it.binaryName)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	logger.Infof("Latest %s version: %s", it.binaryName, latestVersion)

	comparison := CompareVersions(it.currentVersion, latestVersion)
	switch {
	case comparison < 0:
		if dryRun {
			logger.Infof("Dry run: Would update %s from %s to %s", it.binaryName, it.currentVersion, latestVersion)
			logger.Infof("Download URL: %s", downloadURL)
			return nil
		}

		if !force && !it.promptForUpdate(latestVersion) {
			logger.Info("Update cancelled by user")
			return nil
		}

		logger.Infof("Updating %s from %s to %s...", it.binaryName, it.currentVersion, latestVersion)
		return it.performUpdate(downloadURL)

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

func (it *Command) performUpdate(downloadURL string) error {
	currentOS := platform.GetOS()

	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
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
	err = currentOS.Download(downloadURL, tempArchive)
	if err != nil {
		return fmt.Errorf("failed to download new version: %w", err)
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
	if _, statErr := os.Stat(extractedBinary); os.IsNotExist(statErr) {
		return fmt.Errorf("binary %q not found in extracted archive", resolvedBinaryName)
	}

	err = currentOS.MakeExecutable(extractedBinary)
	if err != nil {
		return fmt.Errorf("failed to make downloaded file executable: %w", err)
	}

	backupFile := currentExe + ".backup"
	err = currentOS.Move(currentExe, backupFile)
	if err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	err = currentOS.Move(extractedBinary, currentExe)
	if err != nil {
		if restoreErr := currentOS.Move(backupFile, currentExe); restoreErr != nil {
			logger.Errorf("Failed to restore backup: %v", restoreErr)
		}
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	err = currentOS.Remove(backupFile)
	if err != nil {
		logger.Warnf("Failed to remove backup file %s: %v", backupFile, err)
	}

	logger.Infof("%s has been successfully updated!", it.binaryName)
	logger.Infof("Please restart your terminal or run '%s version' to verify the update", it.binaryName)

	return nil
}
