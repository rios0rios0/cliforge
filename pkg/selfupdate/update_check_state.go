package selfupdate

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"
)

// stateFileMode keeps the update check's state private to the user.
const stateFileMode = 0o600

// checkedToday reports whether a lookup already answered today.
func checkedToday(markerPath string, now time.Time) bool {
	info, err := os.Stat(markerPath)
	return err == nil && !ShouldCheckForUpdates(info.ModTime(), now)
}

// claimUpdateCheckAttempt counts one more lookup against today's limit and
// reports whether it may start. It refuses when the limit is reached, and also
// when the count cannot be read or written: a lookup it cannot count is one
// nothing would stop from repeating on every run.
func claimUpdateCheckAttempt(attemptsPath string, now time.Time) bool {
	attempts, ok := attemptsMadeToday(attemptsPath, now)
	if !ok {
		logger.Debugf("failed to read the update check attempts in %s, skipping update check", attemptsPath)
		return false
	}
	if attempts >= maxUpdateCheckAttemptsPerDay {
		logger.Debug("update check attempts for today used up, skipping")
		return false
	}
	if err := writeAttempts(attemptsPath, attempts+1, now); err != nil {
		logger.Debugf("failed to record the update check attempt, skipping update check: %v", err)
		return false
	}
	return true
}

// attemptsMadeToday returns how many lookups started today, and false when the
// count cannot be read. A count from an earlier day, or none at all, is zero, so
// even a count that cannot be read only holds for the rest of the day.
func attemptsMadeToday(attemptsPath string, now time.Time) (int, bool) {
	info, err := os.Stat(attemptsPath)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, true
	}
	if err != nil {
		return 0, false
	}
	if ShouldCheckForUpdates(info.ModTime(), now) {
		return 0, true
	}
	content, err := os.ReadFile(filepath.Clean(attemptsPath))
	if err != nil {
		return 0, false
	}
	attempts, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil || attempts < 0 {
		return 0, false
	}
	return attempts, true
}

// writeAttempts records attempts as the count of lookups started on the day of
// now.
func writeAttempts(attemptsPath string, attempts int, now time.Time) error {
	if err := ensureStateDir(filepath.Dir(attemptsPath)); err != nil {
		return err
	}
	if err := os.WriteFile(attemptsPath, []byte(strconv.Itoa(attempts)), stateFileMode); err != nil {
		return err
	}
	return os.Chtimes(attemptsPath, now, now)
}

// touchFile creates the file (and any missing parent directories) if it does
// not exist and sets both its access and modification times to now.
func touchFile(path string, now time.Time) error {
	if err := ensureStateDir(filepath.Dir(path)); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, stateFileMode)
	if err != nil {
		return err
	}
	if closeErr := file.Close(); closeErr != nil {
		return closeErr
	}
	return os.Chtimes(path, now, now)
}

// ensureStateDir creates the directory the update check keeps its state in.
func ensureStateDir(dir string) error {
	// The state directory is private to the current user, so no group or other
	// access is granted. 0o700 is the tightest mode a directory can use: without
	// the owner execute (search) bit the files inside it are unreachable.
	// Semgrep applies its file threshold of 0o600 to directory creation too,
	// which no usable directory mode can satisfy.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	return os.MkdirAll(dir, 0o700)
}
