package action

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// Log is an implementation of Action that just writes all messages to
// output.
type Log struct {
	formatter types.Formatter
	output    io.Writer

	mu   sync.Mutex
	pool *pool
}

var _ types.Action = &Log{}

// NewLog creates a new instance of ActionLog.
func NewLog(formatter types.Formatter, output io.Writer) *Log {
	return &Log{
		formatter: formatter,
		output:    output,
		pool:      newPool(),
	}
}

// PerformAction implements Action.
func (a *Log) PerformAction(ctx context.Context, messages []types.Message) error {
	buffer := a.pool.Get()
	defer a.pool.Put(buffer)

	var err error

	if err := formatMessage(a.formatter, messages, buffer); err != nil {
		return errors.Trace(err)
	}

	a.mu.Lock()
	_, err = io.Copy(a.output, buffer)
	a.mu.Unlock()

	return errors.Trace(err)
}

func formatMessage(f types.Formatter, messages []types.Message, b *bytes.Buffer) error {
	for _, message := range messages {
		if err := f.Format(b, message); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
