package taily

import (
	"regexp"

	"github.com/juju/errors"
)

type MatcherField struct {
	field  string
	regexp *regexp.Regexp
}

var _ Matcher = MatcherSubstring{}

func NewMatcherField(field string, pattern string) (*MatcherField, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &MatcherField{
		field:  field,
		regexp: r,
	}, nil
}

func (m MatcherField) MatchMessage(message Message) bool {
	value, ok := message.Fields[m.field]
	if !ok {
		return false
	}

	return m.regexp.MatchString(value)
}
