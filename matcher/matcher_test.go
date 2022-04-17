package matcher_test

import (
	"testing"

	"github.com/jeremija/taily/matcher"
	"github.com/jeremija/taily/types"
	"github.com/stretchr/testify/assert"
)

var onetwo = []types.Matcher{
	matcher.Substring("one"),
	matcher.Substring("two"),
}

func TestOr(t *testing.T) {
	or := matcher.Or(onetwo)

	assert.True(t, or.MatchMessage(msg("a one.")))
	assert.True(t, or.MatchMessage(msg("a two.")))
	assert.False(t, or.MatchMessage(msg("a three.")))
}

func TestAnd(t *testing.T) {
	or := matcher.And(onetwo)

	assert.False(t, or.MatchMessage(msg("a one.")))
	assert.False(t, or.MatchMessage(msg("a two.")))
	assert.False(t, or.MatchMessage(msg("a three.")))
	assert.True(t, or.MatchMessage(msg("a one and a two.")))
}

func TestNot(t *testing.T) {
	not := matcher.Not(matcher.And(onetwo))

	assert.True(t, not.MatchMessage(msg("a one.")))
	assert.True(t, not.MatchMessage(msg("a two.")))
	assert.True(t, not.MatchMessage(msg("a three.")))
	assert.False(t, not.MatchMessage(msg("a one and a two.")))
}
