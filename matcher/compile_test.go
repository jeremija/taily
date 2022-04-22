package matcher_test

import (
	"testing"

	"github.com/jeremija/taily/matcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompile(t *testing.T) {
	m, err := matcher.Compile(`(or (eq "test1") (eq "test2"))`)
	require.NoError(t, err)

	assert.False(t, m.MatchMessage(msg("test")))
	assert.True(t, m.MatchMessage(msg("test1")))
	assert.True(t, m.MatchMessage(msg("test2")))
}
