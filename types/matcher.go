package types

type Matcher interface {
	MatchMessage(message Message) bool
}
