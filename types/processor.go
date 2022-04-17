package types

import (
	"context"
	"time"
)

// Processor is a message processor.
type Processor interface {
	// ProcessMessage processes the message read.
	ProcessMessage(context.Context, Message) error
	Tick(context.Context, time.Time) error
}
