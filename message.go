package guardlog

import "time"

type Message struct {
	Timestamp time.Time `json:"ts"`
	Cursor    string    `json:"cursor,omitempty"`
	Fields    Fields    `json:"fields,omitempty"`
	Source    Source    `json:"source,omitempty"`
	ReaderID  ReaderID  `json:"reader_id,omitempty"`
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
