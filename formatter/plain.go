package formatter

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// Plain just formats the message as plain text.
type Plain struct{}

// NewPlain creates a new instance of Plain.
func NewPlain() Plain {
	return Plain{}
}

// Assert that Plain implements types.Formatter.
var _ types.Formatter = Plain{}

// Format implements Formatter.
func (f Plain) Format(buf *bytes.Buffer, message types.Message) error {
	_, err := fmt.Fprintf(buf, "%s %s %s\n", message.Timestamp.Format(time.RFC3339Nano), message.ReaderID, message.Fields)

	return errors.Trace(err)
}
