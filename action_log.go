package taily

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/juju/errors"
)

// ActionLog is an implementation of Action that just writes all messages to
// output.
type ActionLog struct {
	formatter Formatter
	output    io.Writer

	mu   sync.Mutex
	pool sync.Pool
}

var _ Action = &ActionLog{}

// NewActionLog creates a new instance of ActionLog.
func NewActionLog(formatter Formatter, output io.Writer) *ActionLog {
	return &ActionLog{
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
func (a *ActionLog) PerformAction(ctx context.Context, messages []Message) error {
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
