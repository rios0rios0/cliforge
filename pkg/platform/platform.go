package platform

import (
	"runtime"
	"strings"
)

// Info holds OS and architecture information.
type Info struct {
	OS   string
	Arch string
}

// GetInfo returns the current operating system and architecture.
func GetInfo() Info {
	return Info{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// GetPlatformString returns a formatted platform string in the format OS_ARCH.
func (p Info) GetPlatformString() string {
	return p.OS + "_" + p.Arch
}

// GetArchString returns the architecture string normalized for release binaries.
func (p Info) GetArchString() string {
	// Handle Android architecture which includes "android_" prefix
	if after, ok := strings.CutPrefix(p.Arch, "android_"); ok {
		return after
	}
	return p.Arch
}

// GetOSString returns the OS string normalized for release binaries.
func (p Info) GetOSString() string {
	// Android uses Linux binaries
	if p.OS == "android" {
		return "linux"
	}
	return p.OS
}
