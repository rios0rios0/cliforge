package selfupdate_test

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/selfupdate"
)

// assertUntouched checks that binary still holds the release it held before a
// refused update, with nothing left beside it.
func assertUntouched(t *testing.T, binary string) {
	t.Helper()
	assert.Equal(t, "old release", readFile(t, binary))
	assert.Equal(t, []string{binaryFileName()}, entriesIn(t, filepath.Dir(binary)))
}

func TestExecuteVerifiesTheDownload(t *testing.T) {
	t.Parallel()

	// newRelease is the latest release, whose archive holds "new release".
	newRelease := func(t *testing.T) release {
		t.Helper()
		return release{version: latestVersion, archive: releaseArchive(t, binaryFileName(), "new release")}
	}
	assetName := selfupdate.ReleaseAssetName(binaryName, latestVersion)
	otherDigest := sha256Hex([]byte("another archive"))

	t.Run("should install the release when only checksums.txt states its digest", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.noDigest = true
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, binary))
		assert.Equal(t, int32(1), server.checksumFetches.Load())
	})

	t.Run("should install the release when only GitHub reports its digest", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.noChecksums = true
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, binary))
		assert.Zero(t, server.checksumFetches.Load())
	})

	t.Run("should install the release when checksums.txt lists it in binary mode and in uppercase", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.noDigest = true
		rel.checksums = strings.ToUpper(sha256Hex(rel.archive)) + " *" + assetName + "\r\n"
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, binary))
	})

	t.Run("should install the release when GitHub reports its digest in another algorithm", func(t *testing.T) {
		t.Parallel()
		// given -- only SHA-256 is checked, so checksums.txt alone vouches for it
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.digest = "sha512:" + strings.Repeat("ab", 64)
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.NoError(t, err)
		assert.Equal(t, "new release", readFile(t, binary))
	})

	t.Run("should refuse the release when the archive does not match the digest GitHub reports", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.digest = "sha256:" + otherDigest
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "refusing to install "+assetName)
		require.ErrorContains(t, err, "does not match the digest GitHub reports, "+otherDigest)
		assertUntouched(t, binary)
	})

	t.Run("should refuse the release when the archive does not match checksums.txt", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.checksums = checksumsLine(otherDigest, assetName)
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "refusing to install "+assetName)
		require.ErrorContains(t, err, "does not match checksums.txt, "+otherDigest)
		assertUntouched(t, binary)
	})

	t.Run("should refuse the release without downloading it when it states no digest", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.noDigest = true
		rel.noChecksums = true
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "states no SHA-256 digest")
		assert.Zero(t, server.downloads.Load())
		assertUntouched(t, binary)
	})

	t.Run("should refuse the release without downloading it when checksums.txt leaves it out", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.checksums = checksumsLine(sha256Hex(rel.archive), "tool-1.1.0-plan9-mips.tar.gz")
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "checksums.txt does not list it")
		assert.Zero(t, server.downloads.Load())
		assertUntouched(t, binary)
	})

	t.Run("should refuse the release without downloading it when checksums.txt cannot be fetched", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.checksumsStatus = http.StatusNotFound
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "downloading checksums.txt returned status 404")
		assert.Zero(t, server.downloads.Load())
		assertUntouched(t, binary)
	})

	t.Run("should refuse the release without downloading it when a stated digest is malformed", func(t *testing.T) {
		t.Parallel()
		// given
		binary := installedBinary(t, "old release")
		rel := newRelease(t)
		rel.checksums = checksumsLine("not-a-digest", assetName)
		server := newReleaseServer(t, rel)

		// when
		err := commandFor(server, binary).Execute(false, true)

		// then
		require.ErrorContains(t, err, "checksums.txt lists a malformed digest for it")
		assert.Zero(t, server.downloads.Load())
		assertUntouched(t, binary)
	})

	t.Run(
		"should refuse the release without downloading it when GitHub reports a malformed digest",
		func(t *testing.T) {
			t.Parallel()
			// given
			binary := installedBinary(t, "old release")
			rel := newRelease(t)
			rel.digest = "sha256:" + otherDigest[:10]
			server := newReleaseServer(t, rel)

			// when
			err := commandFor(server, binary).Execute(false, true)

			// then
			require.ErrorContains(t, err, "GitHub reports a malformed digest for it")
			assert.Zero(t, server.downloads.Load())
			assertUntouched(t, binary)
		},
	)
}
