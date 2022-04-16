package guardlog

import "time"

type Message struct {
	Timestamp time.Time
	Cursor    string
	Fields    map[string]string
	Source    Source
	WatcherID WatcherID
}

type Source int

const (
	SourceUndefined Source = iota
	SourceStdout
	SourceStderr
)
