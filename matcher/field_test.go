package matcher_test

import (
	"testing"

	"github.com/jeremija/taily/matcher"
	"github.com/jeremija/taily/types"
	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	m, err := matcher.Field("a", "\\")
	assert.Nil(t, m)
	assert.EqualError(t, err, "error parsing regexp: trailing backslash at end of expression: ``")

	m, err = matcher.Field("a", ".*test.*")
	assert.NoError(t, err)
	assert.NotNil(t, m)

	msgField := func(key, value string) types.Message {
		return types.Message{
			Fields: types.Fields{
				key: value,
			},
		}
	}

	assert.False(t, m.MatchMessage(msgField("a", "something")))
	assert.True(t, m.MatchMessage(msgField("a", "this is a test")))
	assert.False(t, m.MatchMessage(msgField("b", "something")))
}
