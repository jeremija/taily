package taily

import (
	"context"
	"time"
)

type ProcessorFactory func() (Processor, error)

// Processor is a message processor.
type Processor interface {
	// ProcessMessage processes the message read.
	ProcessMessage(context.Context, Message) error
	Tick(context.Context, time.Time) error
}

type ProcessorNoTick struct{}

func (d ProcessorNoTick) Tick(context.Context, time.Time) error {
	return nil
}
