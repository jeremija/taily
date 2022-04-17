package taily

import (
	"bytes"
	"encoding/json"

	"github.com/juju/errors"
)

// FormatterJSON formats a message as JSON.
type FormatterJSON struct{}

// NewFormatterJSON creates a new instance of FormatterJSON.
func NewFormatterJSON() FormatterJSON {
	return FormatterJSON{}
}

// Assert that FormatterJSON implements Formatter.
var _ Formatter = FormatterJSON{}

// Format implements Formatter.
func (f FormatterJSON) Format(buf *bytes.Buffer, message Message) error {
	err := json.NewEncoder(buf).Encode(message)

	return errors.Trace(err)
}
