package reader

import (
	"context"
	"sync"
	"time"

	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/jeremija/taily/pipeline"
	"github.com/jeremija/taily/processor"
	"github.com/jeremija/taily/types"
	"github.com/jeremija/taily/watcher"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

// Docker is a Reader that can read Docker events.
type Docker struct {
	params DockerParams
}

// Assert that Docker implements types.Reader.
var _ types.Reader = &Docker{}

// NewDocker creates a new instance of Docker.
func NewDocker(params DockerParams) *Docker {
	params.Logger = params.Logger.WithNamespaceAppended("docker")

	params.Logger = types.LoggerWithReaderID(params.Logger, params.ReaderID)

	return &Docker{
		params: params,
	}
}

// DockerParams contains parameters for NewDocker.
type DockerParams struct {
	types.ReaderParams                   // ReaderParams contains common reader params.
	Client             *client.Client    // Client is the docker client to use.
	Persister          types.Persister   // Persister to load/save container state.
	NewProcessor       processor.Factory // NewProcessor creates a Processor for all messages.
}

// formatDockerSince formats a ts for the ContainerLogs and Events Since
// argument.
func formatDockerSince(ts time.Time) string {
	if !ts.IsZero() {
		// TODO figure out the correct time format.
		return ts.Format("2006-01-02T15:04:05.999999999Z")
	}

	return ""
}

// ReaderID implements Reader.
func (d *Docker) ReaderID() types.ReaderID {
	return d.params.ReaderID
}

// ReadLogs implements Reader.
func (d *Docker) ReadLogs(ctx context.Context, params types.ReadLogsParams) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	defer wg.Wait()

	state := params.State

	since := formatDockerSince(state.Timestamp)

	eventsCh, errCh := d.params.Client.Events(ctx, dtypes.EventsOptions{
		Since: since,
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "event",
				Value: "start",
			},
			filters.KeyValuePair{
				Key:   "event",
				Value: "stop",
			},
		),
	})

	containers, err := d.params.Client.ContainerList(ctx, dtypes.ContainerListOptions{})
	if err != nil {
		return errors.Trace(err)
	}

	type containerWithDone struct {
		// cancel context.CancelFunc()
		done      <-chan struct{}
		container *DockerContainer
	}

	dockerContainers := map[string]*containerWithDone{}

	containerDoneCh := make(chan string)

	watchContainer := func(containerID string) {
		if _, ok := dockerContainers[containerID]; ok {
			return
		}

		dcDaemonID := d.params.ReaderID + types.ReaderID(":"+containerID)

		watcherParams := d.params.ReaderParams
		watcherParams.ReaderID = dcDaemonID

		dockerContainerParams := DockerContainerParams{
			ReaderParams: watcherParams,
			Client:       d.params.Client,
			ContainerID:  containerID,
		}

		logger := d.params.Logger.WithCtx(log.Ctx{
			"container_id": containerID,
		})

		dc := NewDockerContainer(dockerContainerParams)

		done := make(chan struct{})

		prevContainer := dockerContainers[containerID]

		dockerContainers[containerID] = &containerWithDone{
			done:      done,
			container: dc,
		}

		wg.Add(1)

		go func() {
			defer wg.Done()
			defer close(done)

			defer func() {
				select {
				case containerDoneCh <- containerID:
				case <-ctx.Done():
				}
			}()

			if prevContainer != nil {
				logger.Info("Waiting for previous container to terminate", nil)

				select {
				case <-prevContainer.done:
					logger.Info("Previous container terminated", nil)
				case <-ctx.Done():
					logger.Error("Context canceled", ctx.Err(), nil)
					return
				}
			}

			watcherParams := watcher.Params{
				Persister:    d.params.Persister,
				Reader:       dc,
				Logger:       logger,
				InitialState: types.State{}, // TODO make this configurable.
			}

			dw := watcher.New(watcherParams)

			pline := pipeline.New(pipeline.Params{
				Logger:       logger,
				Watcher:      dw,
				NewProcessor: d.params.NewProcessor,
				BufferSize:   0,
			})

			if err := pline.ProcessPipeline(ctx); err != nil {
				if !types.IsError(err, context.Canceled) {
					logger.Error("Watch failed", err, nil)
					return
				}
			}

			logger.Info("Watch done", nil)
		}()
	}

	removeContainer := func(containerID string) {
		delete(dockerContainers, containerID)
	}

	for _, container := range containers {
		watchContainer(container.ID)
	}

	readerID := d.params.ReaderID

	for {
		select {
		case ev, ok := <-eventsCh:
			if !ok { // Not sure if necessary
				eventsCh = nil
				continue
			}

			timestamp := time.Unix(0, ev.TimeNano).UTC()

			message := types.NewMessage(timestamp, readerID, "Container "+ev.Action, types.Fields{
				"action":       ev.Action,
				"container_id": ev.Actor.ID,
			})

			if err := params.Send(ctx, message); err != nil {
				return errors.Trace(err)
			}

			switch ev.Action {
			case "start":
				containerID := ev.Actor.ID
				watchContainer(containerID)

			case "stop":
				// Do not remove the container here so we can process the logs until
				// the shutdown. Instead, we'll remove it once containerDoneCh is
				// written to.

			default:
				return errors.Errorf("unexpected action: %q", ev.Action)
			}
		case containerID := <-containerDoneCh:
			removeContainer(containerID)
		case err := <-errCh:
			return errors.Trace(err)
		case <-ctx.Done():
			return errors.Trace(err)
		}
	}
}
