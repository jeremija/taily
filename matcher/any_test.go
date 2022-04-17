package matcher_test

import (
	"testing"

	"github.com/jeremija/taily/matcher"
	"github.com/stretchr/testify/assert"
)

func TestAny(t *testing.T) {
	any := matcher.Any

	assert.True(t, any.MatchMessage(msg("a one.")))
	assert.True(t, any.MatchMessage(msg("a two.")))
	assert.True(t, any.MatchMessage(msg("a three.")))
}
