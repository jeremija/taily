package taily

import "strings"

type MatcherSubstring struct {
	substring string
}

var _ Matcher = MatcherSubstring{}

func NewMatcherSubstring(substring string) MatcherSubstring {
	return MatcherSubstring{
		substring: substring,
	}
}

func (m MatcherSubstring) MatchMessage(message Message) bool {
	return strings.Contains(message.Text(), m.substring)
}
