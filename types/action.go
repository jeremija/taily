package types

import "context"

// Action defines an action to perform when a match is found.
type Action interface {
	// PerformAction does something with matched messages.
	PerformAction(ctx context.Context, messages []Message) error
}
