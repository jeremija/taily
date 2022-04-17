package guardlog

import (
	"context"
	"fmt"
)

// PersisterNoop is a stupid implementation of Persister which does nothing. It
// can be used for testing so that the daemon always reads something.
type PersisterNoop struct{}

// Assert that PersiterNoop implements Persister.
var _ Persister = PersisterNoop{}

// NewPersisterNoop creates a new intance of PersisterNoop.
func NewPersisterNoop() PersisterNoop {
	return PersisterNoop{}
}

// LoadState implements Persister.
func (n PersisterNoop) LoadState(ctx context.Context, watcherID ReaderID) (State, error) {
	return State{}, nil
}

// SaveState implements Persister.
func (n PersisterNoop) SaveState(ctx context.Context, watcherID ReaderID, state State) error {
	fmt.Println("SaveState", watcherID, state)
	return nil
}
