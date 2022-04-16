package guardlog

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

type Docker struct {
	params DockerParams
}

func NewDocker(params DockerParams) *Docker {
	params.Logger = params.Logger.WithNamespaceAppended("docker")

	params.Logger = params.Logger.WithCtx(log.Ctx{
		"daemon_id": params.WatcherID,
	})

	return &Docker{
		params: params,
	}
}

type DockerParams struct {
	WatcherParams
	Client    *client.Client
	Persister Persister
}

func formatDockerSince(ts time.Time) string {
	if !ts.IsZero() {
		// TODO figure out the correct time format.
		return ts.Format("2006-01-02T15:04:05.999999999Z")
	}

	return ""
}

func (d *Docker) WatcherID() WatcherID {
	return d.params.WatcherID
}

func (d *Docker) Watch(ctx context.Context, params WatchParams) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	state := params.State

	since := formatDockerSince(state.Timestamp)

	eventsCh, errCh := d.params.Client.Events(ctx, types.EventsOptions{
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

	containers, err := d.params.Client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return errors.Trace(err)
	}

	type containerWithDone struct {
		// cancel context.CancelFunc()
		done      <-chan struct{}
		container *DockerContainer
	}

	dockerContainers := map[string]*containerWithDone{}

	watchContainer := func(containerID string) {
		if _, ok := dockerContainers[containerID]; ok {
			return
		}

		dcDaemonID := d.params.WatcherID + WatcherID(":"+containerID)

		daemonParams := d.params.WatcherParams
		daemonParams.WatcherID = dcDaemonID

		dockerContainerParams := DockerContainerParams{
			WatcherParams: d.params.WatcherParams,
			Client:        d.params.Client,
			ContainerID:   containerID,
		}

		logger := d.params.Logger.WithCtx(log.Ctx{
			"container_id": containerID,
		})

		dc := NewDockerContainer(dockerContainerParams)

		// ctx2, cancel2 := context.WithCancel(ctx)
		done := make(chan struct{})

		prevContainer := dockerContainers[containerID]

		dockerContainers[containerID] = &containerWithDone{
			done:      done,
			container: dc,
		}

		go func() {
			defer close(done)

			if prevContainer != nil {
				logger.Info("Waiting for previous container to terminate", nil)

				select {
				case <-prevContainer.done:
				case <-ctx.Done():
					logger.Error("Context canceled", ctx.Err(), nil)
					return
				}
			}

			dwParams := DaemonWatcherParams{
				Persister: d.params.Persister,
				Watcher:   dc,
			}

			dw := NewDaemonWatcher(dwParams)

			go func() {
				logger.Info("Watching", nil)

				if err := dw.WatchDaemon(ctx, d.params.Logger, params.Ch); err != nil {
					logger.Error("Watch failed", err, nil)

					return
				}

				logger.Info("Watch done", nil)
			}()
		}()
	}

	// removeContainer := func(containerID string) {
	// 	delete(dockerContainers, containerID)
	// }

	for _, container := range containers {
		watchContainer(container.ID)
	}

	for {
		select {
		case ev, ok := <-eventsCh:
			if !ok { // Not sure if necessary
				eventsCh = nil
				continue
			}

			d.params.Logger.Info(fmt.Sprintf("Received event: %+v", ev), nil)

			message := Message{
				Timestamp: time.Unix(0, ev.TimeNano).UTC(),
				Fields: map[string]string{
					"action":       ev.Action,
					"container_id": ev.Actor.ID,
				},
				WatcherID: d.params.WatcherID,
			}

			if err := params.Send(ctx, message); err != nil {
				return errors.Trace(err)
			}

			switch ev.Action {
			case "start":
				containerID := ev.Actor.ID
				watchContainer(containerID)

			case "stop":
				// containerID := ev.Actor.ID
				// removeContainer(containerID)

			default:
				return errors.Errorf("unexpected action: %q", ev.Action)
			}
		case err := <-errCh:
			return errors.Trace(err)
		case <-ctx.Done():
			return errors.Trace(err)
		}
	}
}
