package taily

import (
	"context"

	"github.com/juju/errors"
)

// ProcessorAny is an Processor that feeds every message to Action.
type ProcessorAny struct {
	ProcessorNoTick

	action Action
}

// Assert that ProcessorLog implements Processor.
var _ Processor = &ProcessorAny{}

// NewProcessorAny creates a new instance of ProcessorLog.
func NewProcessorAny(action Action) *ProcessorAny {
	return &ProcessorAny{
		action: action,
	}
}

// ProcessMessage implements Processor.
func (p *ProcessorAny) ProcessMessage(ctx context.Context, message Message) error {
	err := p.action.PerformAction(ctx, []Message{message})

	return errors.Trace(err)
}
