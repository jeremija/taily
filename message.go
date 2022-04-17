package guardlog

import "time"

type Message struct {
	Timestamp time.Time
	Cursor    string
	Fields    Fields
	Source    Source
	ReaderID  ReaderID
}

type Fields map[string]string

func NewMessage(
	timestamp time.Time,
	readerID ReaderID,
	text string,
	extra Fields,
) Message {
	fields := make(Fields, len(extra)+1)

	for k, v := range extra {
		fields[k] = v
	}

	fields["MESSAGE"] = text

	return Message{
		Timestamp: timestamp,
		ReaderID:  readerID,
		Fields:    fields,
	}
}

type Source int

const (
	SourceUndefined Source = iota
	SourceStdout
	SourceStderr
)
