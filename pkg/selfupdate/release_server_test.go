package selfupdate_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rios0rios0/cliforge/pkg/selfupdate"
)

const (
	owner          = "acme"
	repo           = "tool"
	binaryName     = "tool"
	currentVersion = "1.0.0"
	latestVersion  = "1.1.0"
)

// release is what a fake GitHub serves as the latest release of owner/repo.
type release struct {
	version string
	// assetName is the archive the release lists; empty means this platform's.
	assetName string
	archive   []byte
	// digest is the digest the API reports for the archive: empty means its real
	// SHA-256, and noDigest reports none at all.
	digest   string
	noDigest bool
	// checksums is the checksums.txt published beside the archive: empty means
	// one listing its real SHA-256, and noChecksums publishes none at all.
	checksums   string
	noChecksums bool
	// latestStatus, assetStatus and checksumsStatus override the HTTP status of
	// the release lookup, of the download and of checksums.txt; zero means 200.
	latestStatus    int
	assetStatus     int
	checksumsStatus int
}

// releaseServer fakes the GitHub API and the release downloads for owner/repo,
// counting the requests it answers.
type releaseServer struct {
	*httptest.Server

	lookups         atomic.Int32
	downloads       atomic.Int32
	checksumFetches atomic.Int32
}

// newReleaseServer serves rel as the latest release until the test ends.
func newReleaseServer(t *testing.T, rel release) *releaseServer {
	t.Helper()
	if rel.assetName == "" {
		rel.assetName = selfupdate.ReleaseAssetName(binaryName, rel.version)
	}
	if rel.digest == "" {
		rel.digest = "sha256:" + sha256Hex(rel.archive)
	}
	if rel.checksums == "" {
		rel.checksums = checksumsLine(sha256Hex(rel.archive), rel.assetName)
	}
	server := &releaseServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/"+owner+"/"+repo+"/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		server.lookups.Add(1)
		if rel.latestStatus != 0 {
			w.WriteHeader(rel.latestStatus)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag_name": rel.version,
			"assets":   releaseAssets(server.URL, rel),
		})
	})
	mux.HandleFunc("GET /assets/{name}", func(w http.ResponseWriter, _ *http.Request) {
		server.downloads.Add(1)
		if rel.assetStatus != 0 {
			w.WriteHeader(rel.assetStatus)
			return
		}
		_, _ = w.Write(rel.archive)
	})
	mux.HandleFunc("GET /checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		server.checksumFetches.Add(1)
		if rel.checksumsStatus != 0 {
			w.WriteHeader(rel.checksumsStatus)
			return
		}
		_, _ = w.Write([]byte(rel.checksums))
	})
	server.Server = httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// releaseAssets lists the assets of rel the way the GitHub API does, with the
// archive's digest null when the release reports none.
func releaseAssets(serverURL string, rel release) []map[string]any {
	var digest any = rel.digest
	if rel.noDigest {
		digest = nil
	}
	assets := []map[string]any{{
		"name":                 rel.assetName,
		"browser_download_url": serverURL + "/assets/" + rel.assetName,
		"digest":               digest,
	}}
	if !rel.noChecksums {
		assets = append(assets, map[string]any{
			"name":                 "checksums.txt",
			"browser_download_url": serverURL + "/checksums.txt",
			"digest":               "sha256:" + sha256Hex([]byte(rel.checksums)),
		})
	}
	return assets
}

// sha256Hex is the SHA-256 digest of data in hex.
func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// checksumsLine is the line GoReleaser writes into checksums.txt for file.
func checksumsLine(digest, file string) string {
	return digest + "  " + file + "\n"
}

// binaryFileName is the file name binaryName has on the running platform.
func binaryFileName() string {
	if runtime.GOOS == "windows" {
		return binaryName + ".exe"
	}
	return binaryName
}

// releaseArchive builds the archive a release ships for the running platform —
// a .zip on Windows, a .tar.gz elsewhere — holding one file named fileName.
func releaseArchive(t *testing.T, fileName, content string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if runtime.GOOS == "windows" {
		archive := zip.NewWriter(&buffer)
		writer, err := archive.Create(fileName)
		require.NoError(t, err)
		_, err = writer.Write([]byte(content))
		require.NoError(t, err)
		require.NoError(t, archive.Close())
		return buffer.Bytes()
	}
	compressed := gzip.NewWriter(&buffer)
	archive := tar.NewWriter(compressed)
	require.NoError(t, archive.WriteHeader(&tar.Header{
		Name: fileName, Mode: 0o755, Size: int64(len(content)), Typeflag: tar.TypeReg,
	}))
	_, err := archive.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, archive.Close())
	require.NoError(t, compressed.Close())
	return buffer.Bytes()
}

// installedBinary creates the stand-in for the running binary, holding content,
// and returns its path.
func installedBinary(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), binaryFileName())
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// readFile returns the content of the file at path.
func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)
	return string(content)
}

// entriesIn lists the names in dir.
func entriesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// identityOf captures the identity of the file at path through an open handle.
// On Windows [os.Stat] defers reading it until a comparison and then reads
// whatever sits at the path by then, which would compare a replaced file with
// itself.
func identityOf(t *testing.T, path string) os.FileInfo {
	t.Helper()
	file, err := os.Open(filepath.Clean(path))
	require.NoError(t, err)
	defer func() { require.NoError(t, file.Close()) }()
	info, err := file.Stat()
	require.NoError(t, err)
	return info
}
