package matcher

import "github.com/jeremija/taily/types"

type Or []types.Matcher

var _ types.Matcher = Or{}

func (m Or) MatchMessage(message types.Message) bool {
	for _, matcher := range m {
		if matcher.MatchMessage(message) {
			return true
		}
	}

	return false
}
