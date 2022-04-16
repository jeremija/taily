package guardlog

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

type DockerContainer struct {
	params DockerContainerParams
}

func NewDockerContainer(params DockerContainerParams) *DockerContainer {
	params.Logger = params.Logger.WithNamespaceAppended("docker_container")

	params.Logger = params.Logger.WithCtx(log.Ctx{
		"daemon_id":           params.WatcherID,
		"docker_container_id": params.ContainerID,
	})

	return &DockerContainer{
		params: params,
	}
}

type DockerContainerParams struct {
	WatcherParams
	Client      *client.Client
	ContainerID string
}

func (d *DockerContainer) WatcherID() WatcherID {
	return d.params.WatcherID
}

func (d *DockerContainer) Watch(ctx context.Context, params WatchParams) error {
	state := params.State

	inspect, err := d.params.Client.ContainerInspect(ctx, d.params.ContainerID)
	if err != nil {
		return errors.Trace(err)
	}

	isTTY := inspect.Config.Tty

	reader, err := d.params.Client.ContainerLogs(ctx, d.params.ContainerID, types.ContainerLogsOptions{
		Since:      formatDockerSince(state.Timestamp),
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
	})
	if err != nil {
		return errors.Trace(err)
	}

	defer reader.Close()

	var stdout, stderr io.Reader

	watcherID := d.params.WatcherID
	stdout = reader
	errCh := make(chan error, 2)

	if !isTTY {
		errCh := make(chan error, 3)

		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()

		go func() {
			defer stdoutReader.Close()
			defer stderrReader.Close()

			_, err := stdcopy.StdCopy(stdoutWriter, stderrWriter, reader)

			errCh <- errors.Trace(err)
		}()
	}

	go func() {
		errCh <- errors.Trace(Scan(ctx, watcherID, SourceStdout, params, stdout))
	}()

	go func() {
		errCh <- errors.Trace(Scan(ctx, watcherID, SourceStderr, params, stderr))
	}()

	for i := 0; i < 3; i++ {
		if readErr := <-errCh; err == nil && readErr != nil {
			err = errors.Trace(readErr)
		}
	}

	return errors.Trace(err)
}

func Scan(ctx context.Context, watcherID WatcherID, source Source, params WatchParams, reader io.Reader) error {
	if reader == nil {
		return nil
	}

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		split := strings.SplitN(line, " ", 2)

		if len(split) < 2 {
			return errors.Errorf("failed to process line: %q", line)
		}

		timestamp, err := time.Parse(jsonmessage.RFC3339NanoFixed, split[0])
		if err != nil {
			return errors.Trace(err)
		}

		message := Message{
			Timestamp: timestamp,
			Fields: map[string]string{
				"message": split[1],
			},
			Source:    source,
			WatcherID: watcherID,
		}

		if err := params.Send(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}

	return errors.Trace(scanner.Err())
}
