package taily

import "time"

// Message is the main log message.
type Message struct {
	Timestamp time.Time `json:"ts"`                  // Timestamp of the event.
	Cursor    string    `json:"cursor,omitempty"`    // Cursor position of the message.
	Fields    Fields    `json:"fields,omitempty"`    // Fields contains message fields.
	Source    Source    `json:"source,omitempty"`    // Source the message was read from.
	ReaderID  ReaderID  `json:"reader_id,omitempty"` // ReaderID that read the message.
}

// Fields contains the message key-value pairs.
type Fields map[string]string

// NewMessage is a helper function for creating a Message. The text will be
// added as the MESSAGE field.
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

// Source describes the message source.
type Source int

const (
	SourceUndefined Source = iota // SourceUndefined is the default.
	SourceStdout                  // SourceStdout means the message was read from stdout.
	SourceStderr                  // SourceStderr means the message was read from stderr.
)
