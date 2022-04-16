package guardlog

import (
	"context"
	"time"

	"github.com/peer-calls/log"

	"github.com/juju/errors"
)

type WatcherID string

type Watcher interface {
	WatcherID() WatcherID
	Watch(context.Context, WatchParams) error
}

type WatcherParams struct {
	WatcherID WatcherID
	Logger    log.Logger
}

type WatchParams struct {
	State State
	Ch    chan<- Message
}

func (w WatchParams) Send(ctx context.Context, message Message) error {
	select {
	case w.Ch <- message:
		return nil
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}
}

type DaemonWatcher struct {
	params DaemonWatcherParams
}

type DaemonWatcherParams struct {
	Persister Persister
	Watcher   Watcher
}

func NewDaemonWatcher(params DaemonWatcherParams) *DaemonWatcher {
	return &DaemonWatcher{
		params: params,
	}
}

func (dw *DaemonWatcher) WatchDaemon(ctx context.Context, ch chan<- Message) error {
	defer close(ch)

	watcherID := dw.params.Watcher.WatcherID()

	state, err := dw.params.Persister.LoadState(ctx, watcherID)
	if err != nil {
		return errors.Trace(err)
	}

	localCh := make(chan Message)
	errCh := make(chan error, 1)

	params := WatchParams{
		State: state,
		Ch:    localCh,
	}

	go func() {
		defer close(localCh)

		errCh <- errors.Trace(dw.params.Watcher.Watch(ctx, params))
	}()

	defer func() {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		dw.params.Persister.SaveState(ctx2, watcherID, state)
	}()

	for msg := range localCh {
		select {
		case ch <- msg:
			state = state.WithTimestamp(msg.Timestamp)
		case <-ctx.Done():
			return errors.Trace(err)
		}
	}

	return errors.Trace(<-errCh)
}
