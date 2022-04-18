package formatter

import (
	"bytes"

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
	_, err := buf.WriteString(message.Text() + "\n")

	return errors.Trace(err)
}
