package taily

import (
	"context"

	"github.com/peer-calls/log"

	"github.com/juju/errors"
)

// ReaderID is a unique ID of a reader.
// TODO ensure that when we crete multiple instances of the same readers they
// actually get assigned an unique ID.
type ReaderID string

// Reader describes a component that can read logs.
type Reader interface {
	// ReaderID returns the reader's ID.
	ReaderID() ReaderID
	// ReadLogs reads logs until context is done, or an error is encountered.
	// Implementations must not close the ReadLogsParams.Ch as that is done
	// conditionally in Watcher.
	ReadLogs(context.Context, ReadLogsParams) error
}

// ReaderParams contains common parameters for all Reader implementations.
type ReaderParams struct {
	ReaderID ReaderID
	Logger   log.Logger
}

// ReadLogsParams contains parameters for Reader.ReadLogs.
type ReadLogsParams struct {
	State State          // State for resuming reading.
	Ch    chan<- Message // Ch is a channel to write the messages to.
}

// ReadLogsParams is a convenience wrapper that tries to send to Ch until the
// ctx is done.
func (w ReadLogsParams) Send(ctx context.Context, message Message) error {
	select {
	case w.Ch <- message:
		return nil
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}
}
