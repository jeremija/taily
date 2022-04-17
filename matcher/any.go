package matcher

import "github.com/jeremija/taily/types"

type any struct{}

var Any types.Matcher = any{}

func (m any) MatchMessage(message types.Message) bool {
	return true
}
