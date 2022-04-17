package reader

import (
	"context"
	"runtime"
	"time"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

//Journald is a Reader that can read events from systemd's journal.
type Journald struct {
	params JournaldParams
}

// Assert that Journald implements Reader.
var _ types.Reader = &Journald{}

// NewJournald creates a new instance of Journald.
func NewJournald(params JournaldParams) *Journald {
	params.Logger = params.Logger.WithNamespaceAppended("journald")

	params.Logger = types.LoggerWithReaderID(params.Logger, params.ReaderID)

	if params.NewJournal == nil {
		params.NewJournal = sdjournal.NewJournal
	}

	return &Journald{
		params: params,
	}
}

// JournaldParams contains parameters for NewJournald.
type JournaldParams struct {
	types.ReaderParams                                    // ReaderParams contains common reader params.
	NewJournal         func() (*sdjournal.Journal, error) // NewJournal creates a Journal to read from.
}

// ReaderID implements Reader.
func (d *Journald) ReaderID() types.ReaderID {
	return d.params.ReaderID
}

// ReadLogs implements Reader.
func (d *Journald) ReadLogs(ctx context.Context, params types.ReadLogsParams) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// We use a factory so all calls can be locked to the OS thread.
	journal, err := d.params.NewJournal()
	if err != nil {
		return errors.Trace(err)
	}

	state := params.State

	if cursor := state.Cursor; cursor != "" {
		cursor := state.Cursor

		if err := journal.SeekCursor(string(cursor)); err != nil {
			d.params.Logger.Error("Failed to seek cursor", err, log.Ctx{
				"cursor": cursor,
			})
		}
	} else if ts := state.Timestamp; !ts.IsZero() {
		usec := uint64(ts.UnixMicro())

		if err := journal.SeekRealtimeUsec(usec); err != nil {
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

			waitResult := journal.Wait(time.Second)

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
		c, err := journal.Next()
		if err != nil {
			return errors.Trace(err)
		}

		if c == 0 {
			if err := waitForChange(ctx); err != nil {
				return errors.Trace(err)
			}

			continue
		}

		entry, err := journal.GetEntry()
		if err != nil {
			return errors.Trace(err)
		}

		message := types.Message{
			Timestamp: time.UnixMicro(int64(entry.RealtimeTimestamp)).UTC(),
			Cursor:    entry.Cursor,
			Fields:    entry.Fields,
			ReaderID:  d.params.ReaderID,
		}

		if err := params.Send(ctx, message); err != nil {
			return errors.Trace(err)
		}
	}
}
