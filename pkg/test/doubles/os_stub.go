package doubles

import "github.com/rios0rios0/cliforge/pkg/platform"

// Compile-time interface check.
var _ platform.OS = (*OSStub)(nil)

// OSStub implements platform.OS for testing.
type OSStub struct {
	DownloadErr       error
	ExtractErr        error
	MoveErr           error
	RemoveErr         error
	MakeExecutableErr error
}

func (s *OSStub) Download(_, _ string) error    { return s.DownloadErr }
func (s *OSStub) Extract(_, _ string) error     { return s.ExtractErr }
func (s *OSStub) Move(_, _ string) error        { return s.MoveErr }
func (s *OSStub) Remove(_ string) error         { return s.RemoveErr }
func (s *OSStub) MakeExecutable(_ string) error { return s.MakeExecutableErr }
