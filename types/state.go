package types

import (
	"fmt"
	"time"
)

// State describes the reader state.
type State struct {
	// Timestamp is the last timestamp read.
	Timestamp time.Time `json:"time"`
	// NumMessages is the number of messages read with the same Timestamp.
	NumMessages int `json:"num_messages"`
	// Cursor is the current cursor, if any.
	Cursor string `json:"cursor,omitempty"`
}

// WithTimestamp returns a new State with timestamp set. If the timestamp has
// advanced, the NumMessages is reset.
func (s State) WithTimestamp(timestamp time.Time) State {
	if timestamp.After(s.Timestamp) {
		s.Timestamp = timestamp
		s.NumMessages = 0
	}

	s.NumMessages++

	return s
}

// WithCursor returns a new State with cursor.
func (s State) WithCursor(cursor string) State {
	s.Cursor = cursor

	return s
}

// String implements fmt.Stringer.
func (s State) String() string {
	return fmt.Sprintf("State{ts=%s same_timestamp=%d cursor=%q}",
		s.Timestamp.Format(time.RFC3339Nano),
		s.NumMessages,
		s.Cursor,
	)
}
