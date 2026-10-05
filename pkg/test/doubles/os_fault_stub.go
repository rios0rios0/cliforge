package doubles

import "github.com/rios0rios0/cliforge/pkg/platform"

// Compile-time interface check.
var _ platform.OS = (*OSFaultStub)(nil)

// OSFaultStub implements platform.OS over a real one, failing only the moves and
// removals its predicates select, so a test can drive a self-update into one
// failure while every other file operation still happens for real.
type OSFaultStub struct {
	OS          platform.OS
	FailsMove   func(src, dst string) bool
	FailsRemove func(path string) bool
	Err         error
}

func (s *OSFaultStub) Download(url, path string) error { return s.OS.Download(url, path) }
func (s *OSFaultStub) Extract(archive, dest string) error {
	return s.OS.Extract(archive, dest)
}

func (s *OSFaultStub) Move(src, dst string) error {
	if s.FailsMove != nil && s.FailsMove(src, dst) {
		return s.Err
	}
	return s.OS.Move(src, dst)
}

func (s *OSFaultStub) Remove(path string) error {
	if s.FailsRemove != nil && s.FailsRemove(path) {
		return s.Err
	}
	return s.OS.Remove(path)
}

func (s *OSFaultStub) MakeExecutable(path string) error { return s.OS.MakeExecutable(path) }
