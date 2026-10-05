package selfupdate_test

import (
	"os"
	"testing"
	"time"
)

const (
	// runAsReleaseEnv makes this test binary stand in for a release that keeps
	// running: started with it set, the binary only waits to be killed.
	runAsReleaseEnv = "CLIFORGE_TEST_RUN_AS_RELEASE"
	// releaseLifetime bounds how long such a stand-in waits if nobody kills it.
	releaseLifetime = 2 * time.Minute
)

func TestMain(m *testing.M) {
	if os.Getenv(runAsReleaseEnv) == "1" {
		time.Sleep(releaseLifetime)
		return
	}
	os.Exit(m.Run())
}
