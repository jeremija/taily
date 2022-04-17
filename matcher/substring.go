package matcher

import (
	"strings"

	"github.com/jeremija/taily/types"
)

type Substring string

var _ types.Matcher = Substring("")

func (m Substring) MatchMessage(message types.Message) bool {
	return strings.Contains(message.Text(), string(m))
}
