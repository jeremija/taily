package processor

import (
	"context"
	"time"
)

type NoTick struct{}

func (d NoTick) Tick(context.Context, time.Time) error {
	return nil
}
