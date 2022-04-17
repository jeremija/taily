package guardlog

import (
	"context"
)

// Persister is a component for loading and reading reader state.
type Persister interface {
	// LoadState loads the reader state. When the state doees not exist, it must
	// return no error.
	LoadState(context.Context, ReaderID) (State, error)
	// SaveSave saves the reader state.
	SaveState(context.Context, ReaderID, State) error
}
