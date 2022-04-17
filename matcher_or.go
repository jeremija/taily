package taily

type MatcherOr struct {
	matchers []Matcher
}

var _ Matcher = MatcherOr{}

func NewMatcherOr(matchers []Matcher) MatcherOr {
	return MatcherOr{
		matchers: matchers,
	}
}

func (m MatcherOr) MatchMessage(message Message) bool {
	for _, matcher := range m.matchers {
		if matcher.MatchMessage(message) {
			return true
		}
	}

	return false
}
