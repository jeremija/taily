package guardlog

import (
	"context"
)

type Persister interface {
	LoadState(context.Context, WatcherID) (State, error)
	SaveState(context.Context, WatcherID, State) error
}
