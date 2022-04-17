package taily

type MatcherNot struct {
	matcher Matcher
}

var _ Matcher = MatcherNot{}

func NewMatcherNot(matcher Matcher) MatcherNot {
	return MatcherNot{
		matcher: matcher,
	}
}

func (m MatcherNot) MatchMessage(message Message) bool {
	return !m.matcher.MatchMessage(message)
}
