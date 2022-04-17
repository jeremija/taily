package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeremija/taily"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

func main() {
	if err := main2(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v", err)
		os.Exit(1)
	}
}

func main2(argv []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)
	defer cancel()

	logger := log.NewFromEnv("TAILY_LOG")
	config, err := taily.NewConfigFromEnv("TAILY_CONFIG")
	if err != nil {
		return errors.Trace(err)
	}

	persister, err := taily.NewPersisterFromConfig(config.Persister)
	if err != nil {
		return errors.Trace(err)
	}

	processorsMap, err := taily.NewProcessorsMap(config.Processors)
	if err != nil {
		return errors.Trace(err)
	}

	errCh := make(chan error, len(config.Watchers))

	for _, config := range config.Watchers {
		processor, err := taily.NewProcessorsFromMap(processorsMap, config.Processors)
		if err != nil {
			errCh <- errors.Trace(err)
			continue
		}

		watcher, err := taily.NewReaderFromConfig(logger, persister, config)
		if err != nil {
			errCh <- errors.Trace(err)
			continue
		}

		dw := taily.NewWatcher(taily.WatcherParams{
			Persister:    persister,
			Reader:       watcher,
			Logger:       logger,
			InitialState: config.InitialState,
		})

		go func() {
			ch := make(chan taily.Message)
			localErrCh := make(chan error, 1)

			go func() {
				localErrCh <- errors.Trace(dw.Watch(ctx, ch))
			}()

			for message := range ch {
				if err := processor.ProcessMessage(message); err != nil {
					logger.Error("Failed to process message", err, nil)
				}
			}

			// Post error only until we finished processing messages.
			errCh <- errors.Trace(<-localErrCh)
		}()
	}

	numErrors := 0

	for i := 0; i < len(config.Watchers); i++ {
		if err := <-errCh; err != nil {
			if taily.IsError(err, context.Canceled) {
				logger.Info("Watcher complete", nil)
			} else {
				numErrors++
				logger.Error("Watcher failed", err, nil)
			}
		}
	}

	if numErrors > 0 {
		return errors.Errorf("errors encountered: %d", numErrors)
	}

	return nil
}
