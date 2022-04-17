package taily

import (
	"context"
	"time"

	"github.com/juju/errors"
)

// Processors implements Processor by procesing the messages in sequence until
// the end, or until an error is reached.
type Processors []Processor

// Assert that Processors implements Processor.
var _ Processor = Processors{}

// ProcessMessage implements Processor.
func (p Processors) ProcessMessage(ctx context.Context, message Message) error {
	for _, proc := range p {
		if err := proc.ProcessMessage(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// Tick implements Processor.
func (p Processors) Tick(ctx context.Context, now time.Time) error {
	return nil
}
