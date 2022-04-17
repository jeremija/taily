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
	pool sync.Pool
}

var _ types.Action = &Log{}

// NewLog creates a new instance of ActionLog.
func NewLog(formatter types.Formatter, output io.Writer) *Log {
	return &Log{
		formatter: formatter,
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		output: output,
	}
}

// PerformAction implements Action.
func (a *Log) PerformAction(ctx context.Context, messages []types.Message) error {
	buffer := a.pool.Get().(*bytes.Buffer)

	var err error

	for _, message := range messages {
		if err = a.formatter.Format(buffer, message); err != nil {
			break
		}
	}

	if err == nil {
		a.mu.Lock()
		_, err = io.Copy(a.output, buffer)
		a.mu.Unlock()
	}

	buffer.Reset()
	a.pool.Put(buffer)

	return errors.Trace(err)
}
