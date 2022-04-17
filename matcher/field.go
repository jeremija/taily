package matcher

import (
	"regexp"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

type field struct {
	field  string
	regexp *regexp.Regexp
}

var _ types.Matcher = field{}

func Field(fieldName string, pattern string) (*field, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &field{
		field:  fieldName,
		regexp: r,
	}, nil
}

func (m field) MatchMessage(message types.Message) bool {
	value, ok := message.Fields[m.field]
	if !ok {
		return false
	}

	return m.regexp.MatchString(value)
}
