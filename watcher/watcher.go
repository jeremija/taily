package watcher

import (
	"context"
	"time"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

// Watcher is a component that wraps a Reader and takes care of loading and
// storing state before an after calling Reader.ReadLogs. It ensures that each
// Reader never reads the same messages that were previously read.
type Watcher struct {
	params Params
}

func New(params Params) *Watcher {
	return &Watcher{
		params: params,
	}
}

// Params contains parameters for NewWatcher.
type Params struct {
	Persister    types.Persister // Persister to load and store state with.
	Reader       types.Reader    // Reader to read logs from.
	Logger       log.Logger      // Logger to use.
	InitialState types.State
}

// watch calls ReadLogs and prevents duplicate messages from being read.
func (w *Watcher) watch(ctx context.Context, state types.State, ch chan<- types.Message) (types.State, error) {
	localCh := make(chan types.Message)
	errCh := make(chan error, 1)

	params := types.ReadLogsParams{
		State: state,
		Ch:    localCh,
	}

	go func() {
		defer close(localCh)

		errCh <- errors.Trace(w.params.Reader.ReadLogs(ctx, params))
	}()

	count := 0

	send := func(message types.Message) error {
		select {
		case ch <- message:
			return nil
		case <-ctx.Done():
			return errors.Trace(ctx.Err())
		}
	}

	// Ignore old messages.
	if state.NumMessages > 0 {
		for message := range localCh {
			if message.Timestamp.Equal(state.Timestamp) {
				count++

				if count <= state.NumMessages {
					continue
				}
			}

			state = state.WithTimestamp(message.Timestamp).WithCursor(message.Cursor)

			if err := send(message); err != nil {
				return state, errors.Trace(err)
			}

			break
		}
	}

	for message := range localCh {
		state = state.WithTimestamp(message.Timestamp).WithCursor(message.Cursor)

		if err := send(message); err != nil {
			return state, errors.Trace(err)
		}
	}

	return state, errors.Trace(<-errCh)
}

// persistState persists the Reader's state. It uses a separate context so that
// we can still persist the state upon shutdown (e.g. SIGTERM).
func (w *Watcher) persistState(state types.State) {
	// Use a different context because we still want to be able to persist state
	// on shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	readerID := w.params.Reader.ReaderID()
	logger := w.params.Logger.WithCtx(log.Ctx{
		"state": state.String(),
	})

	// TODO perhaps it would be wiser to call SaveState only after we've
	// successfully processed the message. On the other hand, failure to
	// process could hang the processing indefinitely in case we reached a part
	// that we cannot process
	if err := w.params.Persister.SaveState(ctx, readerID, state); err != nil {
		logger.Error("Saving state", err, nil)
	} else {
		logger.Info("Saved state", nil)
	}
}

// Watch loads the state and invokes the Reader.ReadLogs. It persists the state
// after the reading is done. The ch will be closed after reading is complete.
func (w *Watcher) Watch(ctx context.Context, ch chan<- types.Message) (err error) {
	defer close(ch)

	readerID := w.params.Reader.ReaderID()
	logger := w.params.Logger

	logger.Info("Watch daemon STARTED", nil)
	defer logger.Info("Watch daemon DONE", nil)

	state, err := w.params.Persister.LoadState(ctx, readerID)
	if err != nil {
		return errors.Trace(err)
	}

	if state == (types.State{}) {
		state = w.params.InitialState
	}

	logger.Info("Loaded state", log.Ctx{
		"state": state.String(),
	})

	// Persist state at the end, regardless if we encountered an error or not.

	state, err = w.watch(ctx, state, ch)

	w.persistState(state)

	return errors.Trace(err)
}

// WatchAsync calls Watch in a separate goroutine and returns a channel that
// will return an error upon completion. The resulting channel is buffered so
// it does not need to be read from.
func (w *Watcher) WatchAsync(ctx context.Context, ch chan<- types.Message) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- errors.Trace(w.Watch(ctx, ch))
	}()

	return errCh
}
