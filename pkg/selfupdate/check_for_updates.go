package selfupdate

import (
	"os"
	"path/filepath"
	"time"

	logger "github.com/sirupsen/logrus"
)

const updateCheckMarkerFilename = "last_update_check"

// ShouldCheckForUpdates determines whether an update check should be performed
// based on a reference timestamp. Returns false if the timestamp falls on the
// same calendar day as now (in now's timezone). This is used both for the
// binary's modification time and for the per-day update-check marker file.
func ShouldCheckForUpdates(binaryModTime, now time.Time) bool {
	tY, tM, tD := binaryModTime.In(now.Location()).Date()
	nY, nM, nD := now.Date()
	return tY != nY || tM != nM || tD != nD
}

// CheckForUpdates checks if a newer version of the binary is available on GitHub
// and logs a warning if so. It is designed to be called on CLI startup.
// The check is skipped entirely when the current version is "dev", when the
// binary was modified today, or when an update check has already been performed
// today (tracked via a marker file under the user's cache directory). The
// network call runs in a goroutine to avoid blocking CLI startup. Errors are
// logged at debug level and never returned.
func (it *Command) CheckForUpdates() {
	if it.currentVersion == devVersion {
		logger.Debug("development build detected, skipping update check")
		return
	}

	now := time.Now()

	execPath, err := os.Executable()
	if err != nil {
		logger.Debugf("failed to get executable path: %v", err)
		return
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		logger.Debugf("failed to resolve executable symlinks: %v", err)
		return
	}

	info, err := os.Stat(execPath)
	if err != nil {
		logger.Debugf("failed to stat executable: %v", err)
		return
	}

	if !ShouldCheckForUpdates(info.ModTime(), now) {
		logger.Debug("binary was modified today, skipping update check")
		return
	}

	markerPath := it.updateCheckMarkerPath()
	if markerPath != "" {
		if markerInfo, statErr := os.Stat(markerPath); statErr == nil {
			if !ShouldCheckForUpdates(markerInfo.ModTime(), now) {
				logger.Debug("update check already performed today, skipping")
				return
			}
		}
		if touchErr := touchFile(markerPath, now); touchErr != nil {
			logger.Debugf("failed to update check marker %s: %v", markerPath, touchErr)
		}
	}

	go func() {
		latestVersion, fetchErr := fetchLatestVersion(it.owner, it.repo)
		if fetchErr != nil {
			logger.Debugf("failed to fetch latest release: %v", fetchErr)
			return
		}

		if CompareVersions(it.currentVersion, latestVersion) < 0 {
			logger.Warnf(
				"A new version of %s is available: %s (current: %s). "+
					"Run the self-update command to upgrade.",
				it.binaryName, latestVersion, it.currentVersion,
			)
		}
	}()
}

// updateCheckMarkerPath returns the path to the marker file used to track the
// last time an update check ran. Returns an empty string if the user cache
// directory cannot be resolved.
func (it *Command) updateCheckMarkerPath() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		logger.Debugf("failed to resolve user cache directory: %v", err)
		return ""
	}
	return filepath.Join(cacheDir, it.binaryName, updateCheckMarkerFilename)
}

// touchFile creates the file (and any missing parent directories) if it does
// not exist and sets both its access and modification times to now.
func touchFile(path string, now time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	if closeErr := file.Close(); closeErr != nil {
		return closeErr
	}
	return os.Chtimes(path, now, now)
}
