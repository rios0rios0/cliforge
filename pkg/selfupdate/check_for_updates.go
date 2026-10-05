package selfupdate

import (
	"os"
	"path/filepath"
	"time"

	logger "github.com/sirupsen/logrus"
)

const (
	// updateCheckMarkerFilename marks the day a lookup last answered.
	updateCheckMarkerFilename = "last_update_check"
	// updateCheckAttemptsFilename counts the lookups started today; its
	// modification time says which day the count belongs to.
	updateCheckAttemptsFilename = "update_check_attempts"
	// maxUpdateCheckAttemptsPerDay bounds the lookups a day can start while none
	// of them answers: a command that exits before its lookup returns, a machine
	// that is offline, or an API that is rate-limiting. It keeps a short-lived
	// caller such as a container health probe from turning into a stream of API
	// calls, while the next run after an unanswered lookup still retries at once.
	maxUpdateCheckAttemptsPerDay = 5
)

// ShouldCheckForUpdates determines whether an update check should be performed
// based on a reference timestamp. Returns false if the timestamp falls on the
// same calendar day as now (in now's timezone). This is used both for the
// binary's modification time and for the per-day update-check state files.
func ShouldCheckForUpdates(binaryModTime, now time.Time) bool {
	tY, tM, tD := binaryModTime.In(now.Location()).Date()
	nY, nM, nD := now.Date()
	return tY != nY || tM != nM || tD != nD
}

// CheckForUpdates checks if a newer version of the binary is available on GitHub
// and logs a warning if so. It is designed to be called on CLI startup, and it
// never blocks or fails it: the lookup runs in the background, and its errors are
// only logged at debug level.
//
// It looks at most once a day, and a day counts as checked only once a lookup
// has answered, so a command that exits before its lookup returns leaves the
// check to the next command instead of spending it. Up to
// maxUpdateCheckAttemptsPerDay lookups may start in a day. The check is skipped
// for development builds and for a binary modified today, and the lookup is
// skipped altogether when the state that enforces those limits, kept under the
// user's cache directory, cannot be used, rather than run unthrottled.
func (it *Command) CheckForUpdates() {
	if it.currentVersion == devVersion {
		logger.Debug("development build detected, skipping update check")
		return
	}

	now := it.now()
	if !it.binaryModifiedBeforeToday(now) {
		return
	}

	stateDir, ok := it.updateCheckDir()
	if !ok {
		return
	}
	markerPath := filepath.Join(stateDir, updateCheckMarkerFilename)
	if checkedToday(markerPath, now) {
		logger.Debug("update check already performed today, skipping")
		return
	}
	if !claimUpdateCheckAttempt(filepath.Join(stateDir, updateCheckAttemptsFilename), now) {
		return
	}

	it.background(func() { it.lookUpLatestVersion(markerPath) })
}

// lookUpLatestVersion asks for the latest release, warns when it is newer than
// the running one, and only then marks the day as checked. A process that dies
// between the two warns again later rather than losing the warning.
func (it *Command) lookUpLatestVersion(markerPath string) {
	latestVersion, err := fetchLatestVersion(it.apiBaseURL, it.owner, it.repo)
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

	if err = touchFile(markerPath, it.now()); err != nil {
		logger.Debugf("failed to update check marker %s: %v", markerPath, err)
	}
}

// binaryModifiedBeforeToday reports whether the running binary predates today:
// one built or installed today is as current as it gets.
func (it *Command) binaryModifiedBeforeToday(now time.Time) bool {
	executable, err := it.executable()
	if err != nil {
		logger.Debugf("%v, skipping update check", err)
		return false
	}
	info, err := os.Stat(executable)
	if err != nil {
		logger.Debugf("failed to stat executable: %v", err)
		return false
	}
	if !ShouldCheckForUpdates(info.ModTime(), now) {
		logger.Debug("binary was modified today, skipping update check")
		return false
	}
	return true
}

// updateCheckDir returns the directory the update check keeps its state in,
// named after the binary under the user cache directory, and false when there is
// none it can use.
func (it *Command) updateCheckDir() (string, bool) {
	cacheDir, err := it.cacheDir()
	if err != nil {
		logger.Debugf("failed to resolve user cache directory, skipping update check: %v", err)
		return "", false
	}
	// The binary name becomes one path element, so a name such as "..", "../evil"
	// or "/abs" must not reach outside the cache directory.
	name := it.binaryName
	if !filepath.IsLocal(name) || filepath.Base(name) != name {
		logger.Debugf("invalid binary name for the update check state: %q", name)
		return "", false
	}
	return filepath.Join(cacheDir, name), true
}
