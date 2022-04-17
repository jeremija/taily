package guardlog

import (
	"context"
	"fmt"
)

type PersisterNoop struct{}

var _ Persister = PersisterNoop{}

func NewPersisterNoop() PersisterNoop {
	return PersisterNoop{}
}

func (n PersisterNoop) LoadState(ctx context.Context, watcherID ReaderID) (State, error) {
	return State{}, nil
}

func (n PersisterNoop) SaveState(ctx context.Context, watcherID ReaderID, state State) error {
	fmt.Println("SaveState", watcherID, state)
	return nil
}
