package taily

import (
	"regexp"

	"github.com/juju/errors"
)

func NewMatcherRegexp(pattern string) (Matcher, error) {
	m, err := regexp.Compile(pattern)

	return m, errors.Trace(err)
}
