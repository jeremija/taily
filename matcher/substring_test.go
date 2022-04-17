package matcher_test

import (
	"testing"
	"time"

	"github.com/jeremija/taily/matcher"
	"github.com/jeremija/taily/types"
	"github.com/stretchr/testify/assert"
)

func TestSubstring(t *testing.T) {
	assert.True(t, matcher.Substring("test").MatchMessage(msg("this is a test.")))
	assert.False(t, matcher.Substring("test").MatchMessage(msg("this is something else.")))
}

func msg(str string) types.Message {
	return types.NewMessage(time.Now(), "", str, nil)
}
