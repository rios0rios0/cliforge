package selfupdate_test

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/selfupdate"
)

// estOffsetSeconds is the UTC offset of US Eastern Standard Time.
const estOffsetSeconds = -5 * 60 * 60

func TestShouldCheckForUpdates(t *testing.T) {
	t.Parallel()

	t.Run("should return false when binary was modified today", func(t *testing.T) {
		t.Parallel()
		// given
		now := time.Date(2026, 4, 4, 15, 30, 0, 0, time.UTC)
		modTime := time.Date(2026, 4, 4, 8, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.False(t, result)
	})

	t.Run("should return true when binary was modified yesterday", func(t *testing.T) {
		t.Parallel()
		// given
		now := time.Date(2026, 4, 4, 15, 30, 0, 0, time.UTC)
		modTime := time.Date(2026, 4, 3, 23, 59, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.True(t, result)
	})

	t.Run("should return true when binary was modified a week ago", func(t *testing.T) {
		t.Parallel()
		// given
		now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
		modTime := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.True(t, result)
	})

	t.Run("should return false when binary was modified at start of today", func(t *testing.T) {
		t.Parallel()
		// given
		now := time.Date(2026, 4, 4, 23, 59, 59, 0, time.UTC)
		modTime := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.False(t, result)
	})

	t.Run("should handle timezone differences correctly", func(t *testing.T) {
		t.Parallel()
		// given
		eastern := time.FixedZone("EST", estOffsetSeconds)
		now := time.Date(2026, 4, 4, 2, 0, 0, 0, eastern)
		// modTime is April 4 04:00 UTC, which is April 3 23:00 EST
		modTime := time.Date(2026, 4, 4, 4, 0, 0, 0, time.UTC)

		// when
		result := selfupdate.ShouldCheckForUpdates(modTime, now)

		// then
		assert.True(t, result)
	})
}

// checkFixture is everything an update-check test controls: a release server,
// a binary last modified the day before the fixed clock, and a private cache
// directory.
type checkFixture struct {
	server   *releaseServer
	binary   string
	cacheDir string
	now      time.Time
}

// newCheckFixture serves rel as the latest release, with "now" at noon on a fixed
// day so that "today" and "yesterday" never straddle midnight.
func newCheckFixture(t *testing.T, rel release) *checkFixture {
	t.Helper()
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)
	binary := installedBinary(t, "current release")
	yesterday := now.AddDate(0, 0, -1)
	require.NoError(t, os.Chtimes(binary, yesterday, yesterday))
	return &checkFixture{
		server:   newReleaseServer(t, rel),
		binary:   binary,
		cacheDir: t.TempDir(),
		now:      now,
	}
}

// command builds a command whose update check runs against the fixture and
// finishes its lookup before CheckForUpdates returns.
func (f *checkFixture) command(version string, opts ...selfupdate.TestOption) *selfupdate.Command {
	return f.commandNamed(binaryName, version, opts...)
}

// commandNamed is command for a binary called name.
func (f *checkFixture) commandNamed(name, version string, opts ...selfupdate.TestOption) *selfupdate.Command {
	return selfupdate.NewTestCommand(owner, repo, name, version, append([]selfupdate.TestOption{
		selfupdate.WithAPIBaseURL(f.server.URL),
		selfupdate.WithExecutable(f.binary),
		selfupdate.WithCacheDir(f.cacheDir),
		selfupdate.WithClock(f.now),
		selfupdate.WithSynchronousLookups(),
	}, opts...)...)
}

// statePath is where the update check keeps the given state file.
func (f *checkFixture) statePath(name string) string {
	return filepath.Join(f.cacheDir, binaryName, name)
}

// seed writes a state file holding content, last modified at modified.
func (f *checkFixture) seed(t *testing.T, name, content string, modified time.Time) {
	t.Helper()
	path := f.statePath(name)
	// 0o700 mirrors the mode the production code gives this directory and is the
	// tightest a directory can be while remaining traversable.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	require.NoError(t, os.Chtimes(path, modified, modified))
}

// markedToday reports whether the day is marked as checked.
func (f *checkFixture) markedToday(t *testing.T) bool {
	t.Helper()
	info, err := os.Stat(f.statePath(selfupdate.UpdateCheckMarkerFilename))
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	require.NoError(t, err)
	return !selfupdate.ShouldCheckForUpdates(info.ModTime(), f.now)
}

// attempts returns the recorded count of lookups started today.
func (f *checkFixture) attempts(t *testing.T) int {
	t.Helper()
	count, err := strconv.Atoi(readFile(t, f.statePath(selfupdate.UpdateCheckAttemptsFilename)))
	require.NoError(t, err)
	return count
}

func TestCheckForUpdates(t *testing.T) {
	t.Parallel()

	t.Run("should look up the latest release and mark the day as checked when none answered today", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Equal(t, int32(1), fixture.server.lookups.Load())
		assert.True(t, fixture.markedToday(t))
		assert.Equal(t, 1, fixture.attempts(t))
	})

	t.Run("should mark the day as checked when the current version is already the latest", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: currentVersion})

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Equal(t, int32(1), fixture.server.lookups.Load())
		assert.True(t, fixture.markedToday(t))
	})

	t.Run("should leave the day unmarked when the lookup fails", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{latestStatus: http.StatusForbidden})

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Equal(t, int32(1), fixture.server.lookups.Load())
		assert.False(t, fixture.markedToday(t))
		assert.Equal(t, 1, fixture.attempts(t))
	})

	t.Run("should look up again when an earlier lookup today did not answer", func(t *testing.T) {
		t.Parallel()
		// given -- the first command of the day exited before its lookup returned
		fixture := newCheckFixture(t, release{version: latestVersion})
		fixture.seed(t, selfupdate.UpdateCheckAttemptsFilename, "1", fixture.now)

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Equal(t, int32(1), fixture.server.lookups.Load())
		assert.True(t, fixture.markedToday(t))
		assert.Equal(t, 2, fixture.attempts(t))
	})

	t.Run("should look up on the next run when the previous run exited before its lookup answered", func(t *testing.T) {
		t.Parallel()
		// given -- a quick command that exits before its lookup returns
		fixture := newCheckFixture(t, release{version: latestVersion})
		fixture.command(currentVersion, selfupdate.WithAbandonedLookups()).CheckForUpdates()

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Equal(t, int32(1), fixture.server.lookups.Load())
		assert.True(t, fixture.markedToday(t))
		assert.Equal(t, 2, fixture.attempts(t))
	})

	t.Run("should not look up when a lookup already answered today", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})
		fixture.seed(t, selfupdate.UpdateCheckMarkerFilename, "", fixture.now.Add(-time.Hour))

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
		assert.NoFileExists(t, fixture.statePath(selfupdate.UpdateCheckAttemptsFilename))
	})

	t.Run("should not look up when today's attempts are used up", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})
		limit := strconv.Itoa(selfupdate.MaxUpdateCheckAttemptsPerDay)
		fixture.seed(t, selfupdate.UpdateCheckAttemptsFilename, limit, fixture.now)

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
		assert.Equal(t, selfupdate.MaxUpdateCheckAttemptsPerDay, fixture.attempts(t))
	})

	t.Run("should start a new count when the last attempt was on an earlier day", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})
		limit := strconv.Itoa(selfupdate.MaxUpdateCheckAttemptsPerDay)
		fixture.seed(t, selfupdate.UpdateCheckAttemptsFilename, limit, fixture.now.AddDate(0, 0, -1))

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Equal(t, int32(1), fixture.server.lookups.Load())
		assert.Equal(t, 1, fixture.attempts(t))
	})
}

func TestCheckForUpdatesSkips(t *testing.T) {
	t.Parallel()

	t.Run("should not look up when the binary was modified today", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})
		require.NoError(t, os.Chtimes(fixture.binary, fixture.now, fixture.now))

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
	})

	t.Run("should not look up when the build is a development build", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})

		// when
		fixture.command("dev").CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
	})

	t.Run("should not look up when the cache directory cannot be resolved", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})

		// when
		fixture.command(currentVersion, selfupdate.WithoutCacheDir(errors.New("fixture: no home"))).CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
	})

	t.Run("should not look up when the attempt cannot be recorded", func(t *testing.T) {
		t.Parallel()
		// given -- a file where the state directory belongs
		fixture := newCheckFixture(t, release{version: latestVersion})
		require.NoError(t, os.WriteFile(filepath.Join(fixture.cacheDir, binaryName), []byte("x"), 0o600))

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
	})

	t.Run("should not look up when the attempts cannot be read", func(t *testing.T) {
		t.Parallel()
		// given
		fixture := newCheckFixture(t, release{version: latestVersion})
		fixture.seed(t, selfupdate.UpdateCheckAttemptsFilename, "not a number", fixture.now)

		// when
		fixture.command(currentVersion).CheckForUpdates()

		// then
		assert.Zero(t, fixture.server.lookups.Load())
	})

	for _, name := range []string{"..", "", "../escaped", "nested/name"} {
		t.Run(
			"should not look up when the binary name "+strconv.Quote(name)+" cannot name a directory",
			func(t *testing.T) {
				t.Parallel()
				// given
				fixture := newCheckFixture(t, release{version: latestVersion})

				// when
				fixture.commandNamed(name, currentVersion).CheckForUpdates()

				// then
				assert.Zero(t, fixture.server.lookups.Load())
				assert.Empty(t, entriesIn(t, fixture.cacheDir))
			},
		)
	}
}
