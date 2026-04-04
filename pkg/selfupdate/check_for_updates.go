package selfupdate

import (
	"os"
	"time"

	logger "github.com/sirupsen/logrus"
)

// ShouldCheckForUpdates determines whether an update check should be performed
// based on the binary's modification time. Returns false if the binary was
// modified today (same calendar day as now in now's timezone), indicating
// the binary was recently installed or updated.
func ShouldCheckForUpdates(binaryModTime, now time.Time) bool {
	tY, tM, tD := binaryModTime.In(now.Location()).Date()
	nY, nM, nD := now.Date()
	return tY != nY || tM != nM || tD != nD
}

// CheckForUpdates checks if a newer version of the binary is available on GitHub
// and logs a warning if so. It is designed to be called on CLI startup.
// If the binary was modified today, the check is skipped entirely.
// Errors are logged at debug level and never returned.
func (it *Command) CheckForUpdates() {
	execPath, err := os.Executable()
	if err != nil {
		logger.Debugf("failed to get executable path: %v", err)
		return
	}

	info, err := os.Stat(execPath)
	if err != nil {
		logger.Debugf("failed to stat executable: %v", err)
		return
	}

	if !ShouldCheckForUpdates(info.ModTime(), time.Now()) {
		logger.Debug("binary was modified today, skipping update check")
		return
	}

	latestVersion, _, err := fetchLatestRelease(it.owner, it.repo, it.binaryName)
	if err != nil {
		logger.Debugf("failed to fetch latest release: %v", err)
		return
	}

	if CompareVersions(it.currentVersion, latestVersion) < 0 {
		logger.Warnf(
			"A new version of %s is available: %s (current: %s). "+
				"Run the self-update command to upgrade.",
			it.binaryName, latestVersion, it.currentVersion,
		)
	}
}
