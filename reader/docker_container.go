package reader

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

// DockerContainer is a Reader that can read logs from a single docker
// container.
type DockerContainer struct {
	params DockerContainerParams
}

// Assert that DockerContainer implements Reader.
var _ types.Reader = &DockerContainer{}

// NewDockerContainer creates a new instance of DockerContainer.
func NewDockerContainer(params DockerContainerParams) *DockerContainer {
	params.Logger = params.Logger.WithNamespaceAppended("docker_container")

	params.Logger = types.LoggerWithReaderID(params.Logger, params.ReaderID)

	params.Logger = params.Logger.WithCtx(log.Ctx{
		"docker_container_id": params.ContainerID,
	})

	return &DockerContainer{
		params: params,
	}
}

// DockerContainer contains parameters for NewDockerContainer.
type DockerContainerParams struct {
	types.ReaderParams                // ReaderParams contains common reader params.
	Client             *client.Client // Client is the docker client to use.
	ContainerID        string         // ContainerID to read logs from.
}

// ReaderID implements Reader.
func (d *DockerContainer) ReaderID() types.ReaderID {
	return d.params.ReaderID
}

// ReadLogs implements Reader.
func (d *DockerContainer) ReadLogs(ctx context.Context, params types.ReadLogsParams) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	containerID := d.params.ContainerID
	state := params.State

	inspect, err := d.params.Client.ContainerInspect(ctx, containerID)
	if err != nil {
		return errors.Trace(err)
	}

	isTTY := inspect.Config.Tty

	reader, err := d.params.Client.ContainerLogs(ctx, containerID, dtypes.ContainerLogsOptions{
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

	readerID := d.params.ReaderID
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
			WatcherID:      readerID,
			ContainerID:    containerID,
			Source:         types.SourceStdout,
			ReadLogsParams: params,
			Reader:         stdout,
		}

		errCh <- errors.Trace(scanDockerContainerLogs(ctx, p))
	}()

	go func() {
		p := ScanDockerContainerLogsParams{
			WatcherID:      readerID,
			ContainerID:    containerID,
			Source:         types.SourceStderr,
			ReadLogsParams: params,
			Reader:         stderr,
		}

		errCh <- errors.Trace(scanDockerContainerLogs(ctx, p))
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
	WatcherID      types.ReaderID
	ContainerID    string
	Source         types.Source
	ReadLogsParams types.ReadLogsParams
	Reader         io.Reader
}

func scanDockerContainerLogs(ctx context.Context, params ScanDockerContainerLogsParams) error {
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

		message := types.NewMessage(timestamp.UTC(), params.WatcherID, split[1], types.Fields{
			"container_id": params.ContainerID,
		})
		message.Source = params.Source

		if err := params.ReadLogsParams.Send(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}

	err := scanner.Err()

	if !types.IsError(err, io.ErrClosedPipe) {
		return errors.Trace(err)
	}

	return nil
}
