package taily

import (
	"bytes"
	"fmt"
	"time"

	"github.com/juju/errors"
)

// FormatterPlain just formats the message as plain text.
type FormatterPlain struct{}

// NewFormatterPlain creates a new instance of FormamtterPlain.
func NewFormatterPlain() FormatterPlain {
	return FormatterPlain{}
}

// Assert that FormatterPlain implements Formatter.
var _ Formatter = FormatterPlain{}

// Format implements Formatter.
func (f FormatterPlain) Format(buf *bytes.Buffer, message Message) error {
	_, err := fmt.Fprintf(buf, "%s %s %s\n", message.Timestamp.Format(time.RFC3339Nano), message.ReaderID, message.Fields)

	return errors.Trace(err)
}
