package matcher

import (
	"github.com/jeremija/taily/types"
)

type String string

var _ types.Matcher = String("")

func (m String) MatchMessage(message types.Message) bool {
	return string(m) == message.Text()
}
