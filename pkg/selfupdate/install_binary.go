package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	logger "github.com/sirupsen/logrus"

	"github.com/rios0rios0/cliforge/pkg/platform"
)

const (
	// stagedBinarySuffix names the copy of the new release placed beside the
	// running binary before the swap.
	stagedBinarySuffix = ".new"
	// backupSuffix starts the name the running binary is moved to during the swap.
	backupSuffix = ".backup"
)

// installBinary puts newBinary in place of currentExe through system.
//
// The new binary is first staged beside currentExe. That is the only move that
// can cross volumes, so it runs while nothing has been touched yet, and it leaves
// the swap itself to two renames within one directory. currentExe is then moved
// to a backup and the staged binary moved into its place, always arriving as a
// new file rather than written over the old one, so callers can tell an install
// happened by comparing file identities. Every backup gets its own name: Windows
// can neither delete nor replace the image of a running process, and a binary
// still running from an earlier backup (a daemon, say) must not block this one.
func installBinary(system platform.OS, newBinary, currentExe string) error {
	staged := stagedBinaryPath(currentExe)
	if err := system.Move(newBinary, staged); err != nil {
		return fmt.Errorf("failed to stage new binary: %w", err)
	}

	removeStaleBackups(system, currentExe)

	backup := backupBinaryPath(currentExe, time.Now())
	if err := system.Move(currentExe, backup); err != nil {
		discardStaged(system, staged)
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := system.Move(staged, currentExe); err != nil {
		if restoreErr := system.Move(backup, currentExe); restoreErr != nil {
			return fmt.Errorf(
				"failed to install new binary: %w; restoring the previous one from %s also failed: %w",
				err, backup, restoreErr,
			)
		}
		discardStaged(system, staged)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	if err := system.Remove(backup); err != nil {
		logger.Debugf("Kept the previous binary at %s, which the next update removes: %v", backup, err)
	}
	return nil
}

// discardStaged removes a staged binary that an update did not install.
func discardStaged(system platform.OS, staged string) {
	if err := system.Remove(staged); err != nil {
		logger.Debugf("Failed to remove staged binary %s: %v", staged, err)
	}
}

// removeStaleBackups deletes the backups earlier updates left beside currentExe.
// It is best-effort: Windows keeps the backup of a binary that was still running
// at the time, and it can only go once that process has exited.
func removeStaleBackups(system platform.OS, currentExe string) {
	dir := filepath.Dir(currentExe)
	entries, err := os.ReadDir(dir)
	if err != nil {
		logger.Debugf("Failed to look for old backups in %s: %v", dir, err)
		return
	}
	binaryFileName := filepath.Base(currentExe)
	for _, entry := range entries {
		if entry.IsDir() || !isBackupOf(entry.Name(), binaryFileName) {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if removeErr := system.Remove(path); removeErr != nil {
			logger.Debugf("Kept old backup %s: %v", path, removeErr)
		}
	}
}

// isBackupOf reports whether name is a backup an update made of binaryFileName:
// the "<name>.backup" of earlier releases, or a "<name>.backup-<digits>".
func isBackupOf(name, binaryFileName string) bool {
	rest, ok := strings.CutPrefix(name, binaryFileName+backupSuffix)
	if !ok {
		return false
	}
	if rest == "" {
		return true
	}
	digits, ok := strings.CutPrefix(rest, "-")
	if !ok || digits == "" {
		return false
	}
	_, err := strconv.ParseUint(digits, 10, 64)
	return err == nil
}

// stagedBinaryPath is where the new release waits beside currentExe for the swap.
func stagedBinaryPath(currentExe string) string {
	return currentExe + stagedBinarySuffix
}

// backupBinaryPath is the unique name currentExe is moved to by an update at the
// given time.
func backupBinaryPath(currentExe string, at time.Time) string {
	return currentExe + backupSuffix + "-" + strconv.FormatInt(at.UnixNano(), 10)
}
