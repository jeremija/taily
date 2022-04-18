package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jeremija/taily/config"
	"github.com/jeremija/taily/factory"
	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

var GitDescribe = ""

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

	logger.Info("Starting", log.Ctx{
		"version": GitDescribe,
	})

	cfg, err := config.NewFromEnv("TAILY_CONFIG")
	if err != nil {
		return errors.Trace(err)
	}

	go func() {
		<-ctx.Done()

		logger.Info("Tearing down", nil)
	}()

	pipelines, err := factory.NewPipelines(logger, cfg)
	if err != nil {
		return errors.Trace(err)
	}

	errCh := make(chan error, len(pipelines))

	for i := range pipelines {
		pipeline := pipelines[i]

		go func() {
			errCh <- errors.Trace(pipeline.ProcessPipeline(ctx))
		}()
	}

	numErrors := 0

	for i := 0; i < cap(errCh); i++ {
		if err := <-errCh; err != nil {
			if types.IsError(err, context.Canceled) {
				logger.Info("Watcher complete", nil)
			} else {
				numErrors++
				logger.Error("Watcher failed", err, nil)
			}
		}
	}

	if numErrors > 0 {
		return errors.Errorf("there were errors: %d", numErrors)
	}

	return nil
}
