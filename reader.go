package guardlog

import (
	"context"

	"github.com/peer-calls/log"

	"github.com/juju/errors"
)

type ReaderID string

type Reader interface {
	ReaderID() ReaderID
	ReadLogs(context.Context, ReadLogsParams) error
}

type ReaderParams struct {
	ReaderID ReaderID
	Logger   log.Logger
}

type ReadLogsParams struct {
	State State
	Ch    chan<- Message
}

func (w ReadLogsParams) Send(ctx context.Context, message Message) error {
	select {
	case w.Ch <- message:
		return nil
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}
}
