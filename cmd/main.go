package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeremija/guardlog"
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

	config, err := guardlog.NewConfigFromEnv("GUARDLOG_CONFIG")
	if err != nil {
		return errors.Trace(err)
	}

	logger := log.NewFromEnv("GUARDLOG_LOG")

	persister, err := guardlog.NewPersisterFromConfig(config.Persister)
	if err != nil {
		return errors.Trace(err)
	}

	processorsMap, err := guardlog.NewProcessorsMap(config.Processors)
	if err != nil {
		return errors.Trace(err)
	}

	errCh := make(chan error, len(config.Watchers))

	for _, config := range config.Watchers {
		processor, err := guardlog.NewProcessorsFromMap(processorsMap, config.Processors)
		if err != nil {
			errCh <- errors.Trace(err)
			continue
		}

		watcher, err := guardlog.NewReaderFromConfig(logger, persister, config)
		if err != nil {
			errCh <- errors.Trace(err)
			continue
		}

		dw := guardlog.NewWatcher(guardlog.WatcherParams{
			Persister:    persister,
			Reader:       watcher,
			Logger:       logger,
			InitialState: config.InitialState,
		})

		go func() {
			ch := make(chan guardlog.Message)
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

	var firstErr error

	for i := 0; i < len(config.Watchers); i++ {
		err := <-errCh

		if firstErr == nil && err != nil {
			firstErr = errors.Trace(err)
		}
	}

	return errors.Trace(firstErr)
}
