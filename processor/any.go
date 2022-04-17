package processor

import (
	"context"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// Any is an Processor that feeds every message to Action.
type Any struct {
	NoTick

	action types.Action
}

// Assert that ProcessorLog implements types.Processor.
var _ types.Processor = &Any{}

// NewAny creates a new instance of ProcessorLog.
func NewAny(action types.Action) *Any {
	return &Any{
		action: action,
	}
}

// ProcessMessage implements Processor.
func (p *Any) ProcessMessage(ctx context.Context, message types.Message) error {
	err := p.action.PerformAction(ctx, []types.Message{message})

	return errors.Trace(err)
}
