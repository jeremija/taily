package matcher

import "github.com/jeremija/taily/types"

type And []types.Matcher

var _ types.Matcher = And{}

func (m And) MatchMessage(message types.Message) bool {
	for _, matcher := range m {
		if !matcher.MatchMessage(message) {
			return false
		}
	}

	return true
}
