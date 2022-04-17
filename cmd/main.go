package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
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

	logger := log.New().
		WithConfig(log.NewConfig(log.ConfigMap{
			"**": log.LevelInfo,
		})).
		WithConfig(log.NewConfigFromString(os.Getenv("TAILY_LOG"))).
		WithNamespace("taily")

	config, err := taily.NewConfigFromEnv("TAILY_CONFIG")
	if err != nil {
		return errors.Trace(err)
	}

	var wg sync.WaitGroup

	defer wg.Wait()

	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()

		logger.Info("Tearing down", nil)
	}()

	actionsMap, err := taily.NewActionsMap(config.Actions)
	if err != nil {
		return errors.Trace(err)
	}

	persister, err := taily.NewPersisterFromConfig(config.Persister)
	if err != nil {
		return errors.Trace(err)
	}

	processorsMap, err := taily.NewProcessorsMap(config.Processors, actionsMap)
	if err != nil {
		return errors.Trace(err)
	}

	errCh := make(chan error, len(config.Watchers))

	for _, config := range config.Watchers {
		newProcessor, err := taily.NewProcessorsFromMap(processorsMap, config.Processors)
		if err != nil {
			errCh <- errors.Trace(err)
			continue
		}

		watcher, err := taily.NewReaderFromConfig(logger, persister, newProcessor, config)
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

		pipeline := taily.NewPipeline(taily.PipelineParams{
			Logger:       logger,
			Watcher:      dw,
			NewProcessor: newProcessor,
			BufferSize:   0,
		})

		go func() {
			errCh <- errors.Trace(pipeline.ProcessPipeline(ctx))
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
