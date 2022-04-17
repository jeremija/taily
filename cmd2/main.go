package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/client"
	"github.com/jeremija/guardlog"
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

	watcher := guardlog.NewDockerContainer(guardlog.DockerContainerParams{
		WatcherParams: guardlog.WatcherParams{
			WatcherID: "test",
			Logger:    logger,
		},
		Client:      docker,
		ContainerID: "7cc68f2887f2",
	})

	dw := guardlog.NewDaemonWatcher(guardlog.DaemonWatcherParams{
		Persister: guardlog.NewPersisterNoop(),
		Watcher:   watcher,
		Logger:    logger,
	})

	ch := make(chan guardlog.Message)
	errCh := dw.WatchDaemonAsync(ctx, ch)

	for msg := range ch {
		fmt.Println(msg)
	}

	if err := <-errCh; err != nil {
		panic(err)
	}
}
