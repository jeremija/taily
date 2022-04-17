package guardlog

import (
	"context"
	"runtime"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

type Journald struct {
	params JournaldParams
}

type JournaldParams struct {
	WatcherParams
	Journal *sdjournal.Journal
}

func NewJournald(params JournaldParams) *Journald {
	params.Logger = params.Logger.WithNamespaceAppended("journald")

	params.Logger = LoggerWithWatcherID(params.Logger, params.WatcherID)

	return &Journald{
		params: params,
	}
}

func (d *Journald) WatcherID() WatcherID {
	return d.params.WatcherID
}

func (d *Journald) Watch(ctx context.Context, params WatchParams) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	state := params.State

	if cursor := state.Cursor; cursor != "" {
		cursor := state.Cursor

		if err := d.params.Journal.SeekCursor(string(cursor)); err != nil {
			d.params.Logger.Error("Failed to seek cursor", err, log.Ctx{
				"cursor": cursor,
			})
		}
	} else if ts := state.Timestamp; !ts.IsZero() {
		usec := uint64(ts.UnixMicro())

		if err := d.params.Journal.SeekRealtimeUsec(usec); err != nil {
			d.params.Logger.Error("Failed to seek cursor", err, log.Ctx{
				"cursor": cursor,
			})
		}
	}

	waitForChange := func(ctx context.Context) error {
		for {
			if err := ctx.Err(); err != nil {
				return errors.Trace(err)
			}

			waitResult := d.params.Journal.Wait(time.Second)

			switch waitResult {
			case sdjournal.SD_JOURNAL_NOP: // No change.
				continue
			case sdjournal.SD_JOURNAL_APPEND, sdjournal.SD_JOURNAL_INVALIDATE:
				return nil
			default:
				if waitResult < 0 {
					return errors.Errorf("received error event: %v", waitResult)
				}

				return errors.Errorf("received unknown event: %v", waitResult)
			}
		}
	}

	for {
		c, err := d.params.Journal.Next()
		if err != nil {
			return errors.Trace(err)
		}

		if c == 0 {
			if err := waitForChange(ctx); err != nil {
				return errors.Trace(err)
			}

			continue
		}

		entry, err := d.params.Journal.GetEntry()
		if err != nil {
			return errors.Trace(err)
		}

		message := Message{
			Timestamp: time.UnixMicro(int64(entry.RealtimeTimestamp)).UTC(),
			Cursor:    entry.Cursor,
			Fields:    entry.Fields,
			WatcherID: d.params.WatcherID,
		}

		if err := params.Send(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}
}
