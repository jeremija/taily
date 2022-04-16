package guardlog

import (
	"fmt"
	"time"
)

type State struct {
	Timestamp     time.Time `json:"time"`
	SameTimestamp int       `json:"same_timestamp"`
	Cursor        string    `json:"cursor"`
}

func (s State) WithTimestamp(ts time.Time) State {
	if ts.After(s.Timestamp) {
		s.Timestamp = ts
		s.SameTimestamp = 0
	}

	s.SameTimestamp++

	return s
}

func (s State) WithCursor(cursor string) State {
	s.Cursor = cursor

	return s
}

func (s State) String() string {
	return fmt.Sprintf("State{ts=%s same_timestamp=%d cursor=%q}",
		s.Timestamp.Format(time.RFC3339Nano),
		s.SameTimestamp,
		s.Cursor,
	)
}
