package taily

type Matcher interface {
	MatchString(string) bool
}
