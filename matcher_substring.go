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

func (m MatcherSubstring) MatchString(str string) bool {
	return strings.Contains(str, m.substring)
}
