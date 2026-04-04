//go:build unit

package selfupdate_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
