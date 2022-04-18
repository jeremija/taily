package mock

import (
	"context"
	"sync"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// Reader is a component that can be used for mocking.
type Reader struct {
	id       types.ReaderID
	acceptCh chan *ReadCtx
}

// NewReader creates a new instance of Reader.
func NewReader(id types.ReaderID) *Reader {
	return &Reader{
		id:       id,
		acceptCh: make(chan *ReadCtx),
	}
}

// Assert that Reader implements types.Reader.
var _ types.Reader = &Reader{}

// ReaderID implements types.Reader.
func (r *Reader) ReaderID() types.ReaderID {
	return r.id
}

// ReadLogs implements types.Reader.
func (r *Reader) ReadLogs(ctx context.Context, params types.ReadLogsParams) error {
	readCtx := newReadCtx(params)
	select {
	case r.acceptCh <- readCtx:
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}

	select {
	case <-readCtx.Done():
		return nil
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}
}

// Accept must be called by the user of this mock to ReadLogs. The resulting
// ReadCtx can be used to mock messages.
func (r *Reader) Accept(ctx context.Context) (*ReadCtx, error) {
	select {
	case readCtx := <-r.acceptCh:
		return readCtx, nil
	case <-ctx.Done():
		return nil, errors.Trace(ctx.Err())
	}
}

// newReadCtx creates a new instance of ReadCtx.
func newReadCtx(params types.ReadLogsParams) *ReadCtx {
	return &ReadCtx{
		ch:   params.Ch,
		done: make(chan struct{}),
	}
}

// ReadCtx is a mocked reading context that can be used for mocking messages.
type ReadCtx struct {
	ch   chan<- types.Message
	done chan struct{}
	once sync.Once
}

func (r *ReadCtx) Done() <-chan struct{} {
	return r.done
}

// Close must be called to terminate the ReadCtx. It closes the channel
// returned by done. It can be called multiple times.
func (r *ReadCtx) Close() {
	r.once.Do(func() {
		close(r.done)
	})
}

// MockMessage sends a mock message.
func (r *ReadCtx) MockMessage(ctx context.Context, message types.Message) error {
	select {
	case r.ch <- message:
		return nil
	case <-r.done:
		return errors.Errorf("read context is done")
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}
}
