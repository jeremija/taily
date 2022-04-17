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
	Logger    log.Logger
	NoClose   bool
}

func NewDaemonWatcher(params DaemonWatcherParams) *DaemonWatcher {
	return &DaemonWatcher{
		params: params,
	}
}

func (dw *DaemonWatcher) WatchDaemon(ctx context.Context, ch chan<- Message) error {
	if !dw.params.NoClose {
		defer close(ch)
	}

	watcherID := dw.params.Watcher.WatcherID()

	logger := dw.params.Logger

	logger.Info("Watch daemon STARTED", nil)
	defer logger.Info("Watch daemon DONE", nil)

	state, err := dw.params.Persister.LoadState(ctx, watcherID)
	if err != nil {
		return errors.Trace(err)
	}

	logger.Info("Loaded state", log.Ctx{
		"state": state.String(),
	})

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

		logger.Info("Saving state", nil)

		if err := dw.params.Persister.SaveState(ctx2, watcherID, state); err != nil {
			logger.Error("Saving state", err, nil)
		}
	}()

	count := 0

	// Ignore old messages.
	if state.SameTimestamp > 0 {
	loop:
		for msg := range localCh {
			if msg.Timestamp.Equal(state.Timestamp) {
				count++

				if count <= state.SameTimestamp {
					continue
				}
			}

			state = state.WithTimestamp(msg.Timestamp).WithCursor(msg.Cursor)

			select {
			case ch <- msg:
				break loop
			case <-ctx.Done():
				return errors.Trace(err)
			}
		}
	}

	for msg := range localCh {
		select {
		case ch <- msg:
			state = state.WithTimestamp(msg.Timestamp).WithCursor(msg.Cursor)
		case <-ctx.Done():
			return errors.Trace(err)
		}
	}

	return errors.Trace(<-errCh)
}

func (dw *DaemonWatcher) WatchDaemonAsync(ctx context.Context, ch chan<- Message) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- errors.Trace(dw.WatchDaemon(ctx, ch))
	}()

	return errCh
}
