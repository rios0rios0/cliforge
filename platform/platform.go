package platform

import (
	"runtime"
	"strings"
)

// PlatformInfo holds OS and architecture information.
type PlatformInfo struct {
	OS   string
	Arch string
}

// GetPlatformInfo returns the current operating system and architecture.
func GetPlatformInfo() PlatformInfo {
	return PlatformInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// GetPlatformString returns a formatted platform string in the format OS_ARCH.
func (p PlatformInfo) GetPlatformString() string {
	return p.OS + "_" + p.Arch
}

// GetArchString returns the architecture string normalized for release binaries.
func (p PlatformInfo) GetArchString() string {
	// Handle Android architecture which includes "android_" prefix
	if after, ok := strings.CutPrefix(p.Arch, "android_"); ok {
		return after
	}
	return p.Arch
}

// GetOSString returns the OS string normalized for release binaries.
func (p PlatformInfo) GetOSString() string {
	// Android uses Linux binaries
	if p.OS == "android" {
		return "linux"
	}
	return p.OS
}
