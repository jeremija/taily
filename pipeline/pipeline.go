package pipeline

import (
	"context"
	"time"

	"github.com/jeremija/taily/processor"
	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

type AsyncWatcher interface {
	WatchAsync(context.Context, chan<- types.Message) <-chan error
}

// Pipeline starts a Watcher and feeds all message to the Processor.
type Pipeline struct {
	params Params
}

// Params contains parameters for NewPipeline.
type Params struct {
	Logger       log.Logger        // Logger is used for logging errors.
	Watcher      AsyncWatcher      // Watcher is used to start watching.
	NewProcessor processor.Factory // NewProcessor creates a processor for all message.
	BufferSize   int               // BufferSize is the message buffer size. Defaults to 0.
}

// New creates a new instance of Pipeline.
func New(params Params) *Pipeline {
	return &Pipeline{
		params: params,
	}
}

// ProcessPipeline starts the watch and feeds all messages to Processor.
func (p *Pipeline) ProcessPipeline(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	processor, err := p.params.NewProcessor()
	if err != nil {
		return errors.Trace(err)
	}

	ch := make(chan types.Message, p.params.BufferSize)

	errCh := p.params.Watcher.WatchAsync(ctx, ch)

	tick := time.NewTicker(time.Second) // TODO make this configurable and mockable.

	defer tick.Stop()

loop:
	for {
		select {
		case message, ok := <-ch:
			if !ok {
				break loop
			}

			if err := processor.ProcessMessage(ctx, message); err != nil {
				// Do not exit if we fail to process. Doing so would just stop
				// reading logs altogether.
				p.params.Logger.Error("Failed to process message", err, nil)
			}
		case ts := <-tick.C:
			if err := processor.Tick(ctx, ts); err != nil {
				p.params.Logger.Error("Failed to send tick", err, nil)
			}
		}
	}

	return errors.Trace(<-errCh)
}
