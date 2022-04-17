package guardlog

import (
	"bytes"
	"io"
	"sync"

	"github.com/juju/errors"
)

// ProcessorLog is an implementation of Processor that just
// writes the read messages to output. Be careful when using
// this processor if you're also watching the output of this
// process (e.g. in a docker container or journald), because it
// will print every message read, so it could get stuck in a
// loop if it's reading its own log messages.
type ProcessorLog struct {
	formatter Formatter
	output    io.Writer

	mu   sync.Mutex
	pool sync.Pool
}

// NewProcessorLog creates a new instance of ProcessorLog.
func NewProcessorLog(formatter Formatter, output io.Writer) *ProcessorLog {
	return &ProcessorLog{
		formatter: formatter,
		pool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		output: output,
	}
}

// ProcessMessage implements Processor.
func (p *ProcessorLog) ProcessMessage(message Message) error {
	buffer := p.pool.Get().(*bytes.Buffer)

	if err := p.formatter.Format(buffer, message); err != nil {
		return errors.Trace(err)
	}

	p.mu.Lock()
	_, err := io.Copy(p.output, buffer)
	p.mu.Unlock()

	buffer.Reset()
	p.pool.Put(buffer)

	return errors.Trace(err)
}
