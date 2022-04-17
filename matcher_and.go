package taily

type MatcherAnd struct {
	matchers []Matcher
}

var _ Matcher = MatcherAnd{}

func NewMatcherAnd(matchers []Matcher) MatcherAnd {
	return MatcherAnd{
		matchers: matchers,
	}
}

func (m MatcherAnd) MatchMessage(message Message) bool {
	for _, matcher := range m.matchers {
		if !matcher.MatchMessage(message) {
			return false
		}
	}

	return true
}
