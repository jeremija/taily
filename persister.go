package guardlog

import (
	"context"
)

type Persister interface {
	LoadState(context.Context, ReaderID) (State, error)
	SaveState(context.Context, ReaderID, State) error
}
