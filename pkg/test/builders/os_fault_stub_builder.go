package builders

import (
	"errors"

	testkit "github.com/rios0rios0/testkit/pkg/test"

	"github.com/rios0rios0/cliforge/pkg/platform"
	"github.com/rios0rios0/cliforge/pkg/test/doubles"
)

// errInjectedFault is what a selected operation fails with unless WithErr says
// otherwise.
var errInjectedFault = errors.New("injected fault")

// OSFaultStubBuilder builds OSFaultStub instances using the builder pattern. It
// wraps the real OS of the running platform unless WithDelegate names another.
type OSFaultStubBuilder struct {
	*testkit.BaseBuilder

	delegate    platform.OS
	failsMove   func(src, dst string) bool
	failsRemove func(path string) bool
	err         error
}

// NewOSFaultStubBuilder creates a builder that fails nothing.
func NewOSFaultStubBuilder() *OSFaultStubBuilder {
	return &OSFaultStubBuilder{BaseBuilder: testkit.NewBaseBuilder()}
}

// WithDelegate sets the OS every operation that is not failed goes through.
func (b *OSFaultStubBuilder) WithDelegate(delegate platform.OS) *OSFaultStubBuilder {
	b.delegate = delegate
	return b
}

// WithFailingMove fails the moves the predicate selects.
func (b *OSFaultStubBuilder) WithFailingMove(fails func(src, dst string) bool) *OSFaultStubBuilder {
	b.failsMove = fails
	return b
}

// WithFailingRemove fails the removals the predicate selects.
func (b *OSFaultStubBuilder) WithFailingRemove(fails func(path string) bool) *OSFaultStubBuilder {
	b.failsRemove = fails
	return b
}

// WithErr sets the error a failed operation returns.
func (b *OSFaultStubBuilder) WithErr(err error) *OSFaultStubBuilder {
	b.err = err
	return b
}

func (b *OSFaultStubBuilder) Build() any {
	delegate := b.delegate
	if delegate == nil {
		delegate = platform.GetOS()
	}
	err := b.err
	if err == nil {
		err = errInjectedFault
	}
	return &doubles.OSFaultStub{
		OS:          delegate,
		FailsMove:   b.failsMove,
		FailsRemove: b.failsRemove,
		Err:         err,
	}
}
