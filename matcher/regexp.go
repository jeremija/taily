package matcher

import (
	"regexp"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

type Regexp struct {
	regexp *regexp.Regexp
}

var _ types.Matcher = &Regexp{}

func NewRegexp(pattern string) (*Regexp, error) {
	r, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Regexp{regexp: r}, nil
}

func (m *Regexp) MatchMessage(message types.Message) bool {
	return m.regexp.MatchString(message.Text())
}
