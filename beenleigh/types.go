package beenleigh

import (
	"io"
)

type RunCloser interface {
	Run(Logic)
	io.Closer
}

type RunFunc func(logic Logic, closer <-chan struct{})

type runFuncHandle struct {
	RunFunc
	closeCh chan struct{}
}

func NewRunHandler(runner RunFunc) RunCloser {
	return &runFuncHandle{runner, make(chan struct{})}
}

func (rfh *runFuncHandle) Run(l Logic) {
	rfh.RunFunc(l, rfh.closeCh)
}

func (rfh *runFuncHandle) Close() error {
	close(rfh.closeCh)
	return nil
}

type MethodEvent struct{}

// A specification for creation of new objects.
// Type should be a type available for creation by the logic implementation.
// Data is a context sensitive string, which syntax depends on the type.
// Emitter is a integer, identifying the emitter number to link the new object to.
type Spec struct {
	Emitter int
	Data    string
}
