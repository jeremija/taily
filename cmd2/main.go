package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/jeremija/taily"
	"github.com/peer-calls/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)
	defer cancel()

	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	logger := log.NewFromEnv("GUARDLOG_LOG")

	watcher := taily.NewDockerContainer(taily.DockerContainerParams{
		ReaderParams: taily.ReaderParams{
			ReaderID: "test",
			Logger:   logger,
		},
		Client:      docker,
		ContainerID: "7cc68f2887f2",
	})

	dw := taily.NewWatcher(taily.WatcherParams{
		Persister: taily.NewPersisterNoop(),
		Reader:    watcher,
		Logger:    logger,
	})

	ch := make(chan taily.Message)
	errCh := dw.WatchAsync(ctx, ch)

	for msg := range ch {
		fmt.Println(msg)
	}

	if err := <-errCh; err != nil {
		panic(err)
	}
}
