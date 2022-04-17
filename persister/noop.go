package persister

import (
	"context"
	"fmt"

	"github.com/jeremija/taily/types"
)

// Noop is a stupid implementation of Persister which does nothing. It
// can be used for testing so that the daemon always reads something.
type Noop struct{}

// Assert that Noop implements types.Persister.
var _ types.Persister = Noop{}

// NewNoop creates a new intance of Noop.
func NewNoop() Noop {
	return Noop{}
}

// LoadState implements Persister.
func (n Noop) LoadState(ctx context.Context, readerID types.ReaderID) (types.State, error) {
	return types.State{}, nil
}

// SaveState implements Persister.
func (n Noop) SaveState(ctx context.Context, readerID types.ReaderID, state types.State) error {
	fmt.Println("SaveState", readerID, state)
	return nil
}
