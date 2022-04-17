package taily

import (
	"regexp"

	"github.com/juju/errors"
)

type MatcherRegexp struct {
	regexp *regexp.Regexp
}

func NewMatcherRegexp(pattern string) (*MatcherRegexp, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &MatcherRegexp{regexp: r}, nil
}

func (m *MatcherRegexp) MatchMessage(message Message) bool {
	return m.regexp.MatchString(message.Text())
}
