package guardlog

import "time"

type State struct {
	Timestamp     time.Time
	SameTimestamp int
	Cursor        string
}

func (s State) WithTimestamp(ts time.Time) State {
	if ts.After(s.Timestamp) {
		s.Timestamp = ts
		s.SameTimestamp = 0
	}

	s.SameTimestamp++

	return s
}
