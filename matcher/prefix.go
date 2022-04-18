package matcher

import (
	"strings"

	"github.com/jeremija/taily/types"
)

type Prefix string

var _ types.Matcher = Prefix("")

func (m Prefix) MatchMessage(message types.Message) bool {
	return strings.HasPrefix(message.Text(), string(m))
}
