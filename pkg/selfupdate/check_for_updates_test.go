//go:build unit

package selfupdate_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/selfupdate"
)

func TestShouldCheckForUpdates(t *testing.T) {
	t.Parallel()

	t.Run("should return false when binary was modified today", func(t *testing.T) {
		// given
		now := time.Date(2026, 4, 4, 15, 30, 0, 0, time.UTC)
		modTime := time.Date(2026, 4, 4, 8, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.False(t, result)
	})

	t.Run("should return true when binary was modified yesterday", func(t *testing.T) {
		// given
		now := time.Date(2026, 4, 4, 15, 30, 0, 0, time.UTC)
		modTime := time.Date(2026, 4, 3, 23, 59, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.True(t, result)
	})

	t.Run("should return true when binary was modified a week ago", func(t *testing.T) {
		// given
		now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
		modTime := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.True(t, result)
	})

	t.Run("should return false when binary was modified at start of today", func(t *testing.T) {
		// given
		now := time.Date(2026, 4, 4, 23, 59, 59, 0, time.UTC)
		modTime := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.False(t, result)
	})

	t.Run("should handle timezone differences correctly", func(t *testing.T) {
		// given
		eastern := time.FixedZone("EST", -5*3600)
		now := time.Date(2026, 4, 4, 2, 0, 0, 0, eastern)
		// modTime is April 4 04:00 UTC, which is April 3 23:00 EST
		modTime := time.Date(2026, 4, 4, 4, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.True(t, result)
	})
}

// TestCheckForUpdatesDailyThrottle cannot use t.Parallel because it mutates
// the XDG_CACHE_HOME environment variable via t.Setenv.
func TestCheckForUpdatesDailyThrottle(t *testing.T) {
	t.Run("should not fire HTTP call when marker file was touched today", func(t *testing.T) {
		// given
		cacheDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", cacheDir)

		binaryName := "throttle-test-binary"
		markerDir := filepath.Join(cacheDir, binaryName)
		require.NoError(t, os.MkdirAll(markerDir, 0o755))
		markerPath := filepath.Join(markerDir, "last_update_check")
		file, err := os.Create(markerPath)
		require.NoError(t, err)
		require.NoError(t, file.Close())
		require.NoError(t, os.Chtimes(markerPath, time.Now(), time.Now()))

		cmd := selfupdate.NewCommand("owner-that-does-not-exist", "repo-that-does-not-exist", binaryName, "0.0.1")

		// when
		cmd.CheckForUpdates()

		// then
		info, statErr := os.Stat(markerPath)
		require.NoError(t, statErr)
		assert.WithinDuration(t, time.Now(), info.ModTime(), 5*time.Second)
	})
}
