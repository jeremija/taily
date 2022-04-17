package matcher_test

import (
	"testing"

	"github.com/jeremija/taily/matcher"
	"github.com/stretchr/testify/assert"
)

func TestRegexp(t *testing.T) {
	m, err := matcher.NewRegexp("\\")
	assert.Nil(t, m)
	assert.EqualError(t, err, "error parsing regexp: trailing backslash at end of expression: ``")

	m, err = matcher.NewRegexp(".*test.*")
	assert.NoError(t, err)
	assert.NotNil(t, m)

	assert.True(t, m.MatchMessage(msg("this is another test.")))
	assert.False(t, m.MatchMessage(msg("this is something else again.")))
}
