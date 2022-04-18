package matcher

import (
	"strings"

	"github.com/jeremija/taily/types"
)

type Suffix string

var _ types.Matcher = Suffix("")

func (m Suffix) MatchMessage(message types.Message) bool {
	return strings.HasSuffix(message.Text(), string(m))
}
