package matcher

import "github.com/jeremija/taily/types"

type not struct {
	matcher types.Matcher
}

var _ types.Matcher = not{}

func Not(matcher types.Matcher) types.Matcher {
	return not{
		matcher: matcher,
	}
}

func (m not) MatchMessage(message types.Message) bool {
	return !m.matcher.MatchMessage(message)
}
