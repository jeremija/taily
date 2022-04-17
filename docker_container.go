package taily

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

// DockerContainer is a Reader that can read logs from a single docker
// container.
type DockerContainer struct {
	params DockerContainerParams
}

// Assert that DockerContainer implements Reader.
var _ Reader = &DockerContainer{}

// NewDockerContainer creates a new instance of DockerContainer.
func NewDockerContainer(params DockerContainerParams) *DockerContainer {
	params.Logger = params.Logger.WithNamespaceAppended("docker_container")

	params.Logger = LoggerWithReaderID(params.Logger, params.ReaderID)

	params.Logger = params.Logger.WithCtx(log.Ctx{
		"docker_container_id": params.ContainerID,
	})

	return &DockerContainer{
		params: params,
	}
}

// DockerContainer contains parameters for NewDockerContainer.
type DockerContainerParams struct {
	ReaderParams                // ReaderParams contains common reader params.
	Client       *client.Client // Client is the docker client to use.
	ContainerID  string         // ContainerID to read logs from.
}

// ReaderID implements Reader.
func (d *DockerContainer) ReaderID() ReaderID {
	return d.params.ReaderID
}

// ReadLogs implements Reader.
func (d *DockerContainer) ReadLogs(ctx context.Context, params ReadLogsParams) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	containerID := d.params.ContainerID
	state := params.State

	inspect, err := d.params.Client.ContainerInspect(ctx, containerID)
	if err != nil {
		return errors.Trace(err)
	}

	isTTY := inspect.Config.Tty

	reader, err := d.params.Client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
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

	watcherID := d.params.ReaderID
	stdout = reader
	errCh := make(chan error, 2)

	if !isTTY {
		errCh = make(chan error, 3)

		stdoutReader, stdoutWriter := io.Pipe()
		stderrReader, stderrWriter := io.Pipe()

		stdout = stdoutReader
		stderr = stderrReader

		go func() {
			defer stdoutReader.Close()
			defer stderrReader.Close()

			_, err := stdcopy.StdCopy(stdoutWriter, stderrWriter, reader)

			errCh <- errors.Trace(err)
		}()
	}

	go func() {
		p := ScanDockerContainerLogsParams{
			WatcherID:      watcherID,
			ContainerID:    containerID,
			Source:         SourceStdout,
			ReadLogsParams: params,
			Reader:         stdout,
		}

		errCh <- errors.Trace(ScanDockerContainerLogs(ctx, p))
	}()

	go func() {
		p := ScanDockerContainerLogsParams{
			WatcherID:      watcherID,
			ContainerID:    containerID,
			Source:         SourceStderr,
			ReadLogsParams: params,
			Reader:         stderr,
		}

		errCh <- errors.Trace(ScanDockerContainerLogs(ctx, p))
	}()

	var retErr error

	for i := 0; i < cap(errCh); i++ {
		if err := <-errCh; err != nil {
			if retErr == nil {
				retErr = errors.Trace(err)
			}
		}
	}

	return errors.Trace(retErr)
}

type ScanDockerContainerLogsParams struct {
	WatcherID      ReaderID
	ContainerID    string
	Source         Source
	ReadLogsParams ReadLogsParams
	Reader         io.Reader
}

func ScanDockerContainerLogs(ctx context.Context, params ScanDockerContainerLogsParams) error {
	if params.Reader == nil {
		return nil
	}

	scanner := bufio.NewScanner(params.Reader)

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

		message := NewMessage(timestamp.UTC(), params.WatcherID, split[1], Fields{
			"container_id": params.ContainerID,
		})
		message.Source = params.Source

		if err := params.ReadLogsParams.Send(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}

	err := scanner.Err()

	if !IsError(err, io.ErrClosedPipe) {
		return errors.Trace(err)
	}

	return nil
}
