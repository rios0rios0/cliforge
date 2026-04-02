package builders

import (
	testkit "github.com/rios0rios0/testkit/pkg/test"

	"github.com/rios0rios0/cliforge/test/doubles"
)

// OSStubBuilder builds OSStub instances using the builder pattern.
type OSStubBuilder struct {
	*testkit.BaseBuilder

	downloadErr       error
	extractErr        error
	moveErr           error
	removeErr         error
	makeExecutableErr error
}

// NewOSStubBuilder creates a new builder with default values.
func NewOSStubBuilder() *OSStubBuilder {
	return &OSStubBuilder{BaseBuilder: testkit.NewBaseBuilder()}
}

func (b *OSStubBuilder) WithDownloadErr(err error) *OSStubBuilder {
	b.downloadErr = err
	return b
}

func (b *OSStubBuilder) WithExtractErr(err error) *OSStubBuilder {
	b.extractErr = err
	return b
}

func (b *OSStubBuilder) WithMoveErr(err error) *OSStubBuilder {
	b.moveErr = err
	return b
}

func (b *OSStubBuilder) WithRemoveErr(err error) *OSStubBuilder {
	b.removeErr = err
	return b
}

func (b *OSStubBuilder) WithMakeExecutableErr(err error) *OSStubBuilder {
	b.makeExecutableErr = err
	return b
}

func (b *OSStubBuilder) Build() any {
	return &doubles.OSStub{
		DownloadErr:       b.downloadErr,
		ExtractErr:        b.extractErr,
		MoveErr:           b.moveErr,
		RemoveErr:         b.removeErr,
		MakeExecutableErr: b.makeExecutableErr,
	}
}
