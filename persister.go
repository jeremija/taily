package guardlog

import "context"

type Persister interface {
	LoadState(context.Context, WatcherID) (State, error)
	SaveState(context.Context, WatcherID, State) error
}

type NoopPersister struct{}

var _ Persister = NoopPersister{}

func (n NoopPersister) LoadState(ctx context.Context, daemonID WatcherID) (State, error) {
	return State{}, nil
}

func (n NoopPersister) SaveState(ctx context.Context, daemonID WatcherID, state State) error {
	return nil
}
