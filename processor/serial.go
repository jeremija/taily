package processor

import (
	"context"
	"time"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// Serial implements Processor by procesing the messages in sequence until
// the end, or until an error is reached.
type Serial []types.Processor

// Assert that Processors implements Processor.
var _ types.Processor = Serial{}

// ProcessMessage implements Processor.
func (p Serial) ProcessMessage(ctx context.Context, message types.Message) error {
	for _, proc := range p {
		if err := proc.ProcessMessage(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

// Tick implements Processor.
func (p Serial) Tick(ctx context.Context, now time.Time) error {
	for _, proc := range p {
		if err := proc.Tick(ctx, now); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}
