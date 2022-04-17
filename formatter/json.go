package formatter

import (
	"bytes"
	"encoding/json"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

// JSON formats a message as JSON.
type JSON struct{}

// NewJSON creates a new instance of FormatterJSON.
func NewJSON() JSON {
	return JSON{}
}

// Assert that FormatterJSON implements Formatter.
var _ types.Formatter = JSON{}

// Format implements Formatter.
func (f JSON) Format(buf *bytes.Buffer, message types.Message) error {
	err := json.NewEncoder(buf).Encode(message)

	return errors.Trace(err)
}
