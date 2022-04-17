package taily

type Matcher interface {
	MatchMessage(message Message) bool
}
