package guardlog

import (
	"context"
	"time"

	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

type Watcher struct {
	params WatcherParams
}

type WatcherParams struct {
	Persister Persister  // Persister to load and store state with.
	Reader    Reader     // Reader to read logs from.
	Logger    log.Logger // Logger to use.
	NoClose   bool       //NoClose will prevent Watch from closing ch on exit.
}

func NewWatcher(params WatcherParams) *Watcher {
	return &Watcher{
		params: params,
	}
}

func (dw *Watcher) watch(ctx context.Context, state State, ch chan<- Message) (State, error) {
	localCh := make(chan Message)
	errCh := make(chan error, 1)

	params := ReadLogsParams{
		State: state,
		Ch:    localCh,
	}

	go func() {
		defer close(localCh)

		errCh <- errors.Trace(dw.params.Reader.ReadLogs(ctx, params))
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
				return state, errors.Trace(ctx.Err())
			}
		}
	}

	for msg := range localCh {
		select {
		case ch <- msg:
			state = state.WithTimestamp(msg.Timestamp).WithCursor(msg.Cursor)
		case <-ctx.Done():
			return state, errors.Trace(ctx.Err())
		}
	}

	return state, errors.Trace(<-errCh)
}

func (dw *Watcher) persistState(state State) {
	// Use a different context because we still want to be able to persist state
	// on shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	readerID := dw.params.Reader.ReaderID()
	logger := dw.params.Logger.WithCtx(log.Ctx{
		"state": state.String(),
	})

	// TODO perhaps it would be wiser to call SaveState only after we've
	// successfully processed the message. On the other hand, failure to
	// process could hang the processing indefinitely in case we reached a part
	// that we cannot process
	if err := dw.params.Persister.SaveState(ctx, readerID, state); err != nil {
		logger.Error("Saving state", err, nil)
	} else {
		logger.Info("Saved state", nil)
	}
}

func (dw *Watcher) Watch(ctx context.Context, ch chan<- Message) error {
	if !dw.params.NoClose {
		defer close(ch)
	}

	readerID := dw.params.Reader.ReaderID()
	logger := dw.params.Logger

	logger.Info("Watch daemon STARTED", nil)
	defer logger.Info("Watch daemon DONE", nil)

	state, err := dw.params.Persister.LoadState(ctx, readerID)
	if err != nil {
		return errors.Trace(err)
	}

	logger.Info("Loaded state", log.Ctx{
		"state": state.String(),
	})

	// Persist state at the end, regardless if we encountered an error or not.

	state, err = dw.watch(ctx, state, ch)

	dw.persistState(state)

	return errors.Trace(err)
}

func (dw *Watcher) WatchAsync(ctx context.Context, ch chan<- Message) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- errors.Trace(dw.Watch(ctx, ch))
	}()

	return errCh
}
