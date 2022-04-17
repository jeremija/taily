package taily

import (
	"bytes"

	"github.com/juju/errors"
)

// Formatter formats the message.
type Formatter interface {
	// Format formats the message.
	Format(*bytes.Buffer, Message) error
}

// FormatterFunc allows functions to implement Formatter.
type FormatterFunc func(*bytes.Buffer, Message) error

// Format implements Formatter.
func (f FormatterFunc) Format(buf *bytes.Buffer, message Message) error {
	return errors.Trace(f(buf, message))
}
