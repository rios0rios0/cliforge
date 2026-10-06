package selfupdate

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	logger "github.com/sirupsen/logrus"
)

const (
	// checksumsAssetName is the manifest GoReleaser publishes beside a release's
	// archives, one "<sha256>  <file>" line for each.
	checksumsAssetName = "checksums.txt"
	// sha256DigestPrefix opens the digest GitHub reports for an asset it hashed
	// with SHA-256.
	sha256DigestPrefix = "sha256:"
	// githubDigestSource names where the digest GitHub computed comes from.
	githubDigestSource = "the digest GitHub reports"
)

// errNoDigest reports a release that states no digest for its archive, leaving
// nothing to check the download against.
var errNoDigest = errors.New("the release states no SHA-256 digest for it")

// statedDigest is a SHA-256 digest a release states for its archive, in
// lowercase hex, and where the release states it.
type statedDigest struct {
	source string
	digest string
}

// releaseDigests returns every SHA-256 digest the release states for asset: the
// one GitHub computed when the archive was uploaded, and the archive's line in
// the checksums.txt published beside it. The two travel apart, the first in the
// API answer and the second from the download host, so a release carrying both
// is checked against each. It fails when the release states neither, and when
// checksums.txt cannot be read, leaves the archive out, or lists something that
// is not a SHA-256 digest for it: a release whose own manifest does not vouch
// for its archive is not one to install.
func releaseDigests(asset releaseAsset) ([]statedDigest, error) {
	var digests []statedDigest

	if reported, ok := strings.CutPrefix(asset.digest, sha256DigestPrefix); ok {
		digest, err := parseSHA256(reported)
		if err != nil {
			return nil, fmt.Errorf("GitHub reports a malformed digest for it: %w", err)
		}
		digests = append(digests, statedDigest{source: githubDigestSource, digest: digest})
	} else if asset.digest != "" {
		logger.Debugf("Skipping the digest %q GitHub reports for %s: only SHA-256 is checked", asset.digest, asset.name)
	}

	if asset.checksumsURL != "" {
		manifest, err := fetchChecksums(asset.checksumsURL)
		if err != nil {
			return nil, err
		}
		listed, ok := checksumFor(manifest, asset.name)
		if !ok {
			return nil, fmt.Errorf("%s does not list it", checksumsAssetName)
		}
		digest, err := parseSHA256(listed)
		if err != nil {
			return nil, fmt.Errorf("%s lists a malformed digest for it: %w", checksumsAssetName, err)
		}
		digests = append(digests, statedDigest{source: checksumsAssetName, digest: digest})
	}

	if len(digests) == 0 {
		return nil, errNoDigest
	}

	return digests, nil
}

// checksumFor returns the digest manifest lists for name. GoReleaser writes the
// sha256sum format: "<hex>  <name>", or "<hex> *<name>" for a binary-mode entry.
func checksumFor(manifest, name string) (string, bool) {
	for line := range strings.Lines(manifest) {
		digest, file, ok := strings.Cut(strings.TrimRight(line, "\r\n"), " ")
		if !ok {
			continue
		}
		file = strings.TrimPrefix(strings.TrimPrefix(file, " "), "*")
		if file == name {
			return digest, true
		}
	}

	return "", false
}

// parseSHA256 returns digest as lowercase hex, failing unless it is the hex
// encoding of a SHA-256 sum.
func parseSHA256(digest string) (string, error) {
	sum, err := hex.DecodeString(digest)
	if err != nil || len(sum) != sha256.Size {
		return "", fmt.Errorf("%q is not a SHA-256 digest", digest)
	}

	return hex.EncodeToString(sum), nil
}

// verifyArchive hashes the archive at path and fails unless it matches every
// digest in want.
func verifyArchive(path string, want []statedDigest) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open the download: %w", err)
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to hash the download: %w", err)
	}
	got := hex.EncodeToString(hash.Sum(nil))

	for _, stated := range want {
		if got != stated.digest {
			return fmt.Errorf("its SHA-256 %s does not match %s, %s", got, stated.source, stated.digest)
		}
	}

	return nil
}
