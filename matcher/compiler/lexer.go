package compiler

import (
	"fmt"

	"github.com/juju/errors"
)

type TokenKind int

const (
	Unknown TokenKind = iota
	OpenParenthesis
	CloseParenthesis
	Function
	Text
)

func (s TokenKind) String() string {
	switch s {
	case Unknown:
		return "unknown"
	case OpenParenthesis:
		return "("
	case CloseParenthesis:
		return ")"
	case Function:
		return "function"
	case Text:
		return "str"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

type Token struct {
	Kind     TokenKind
	Value    string
	StartPos int
	EndPos   int
}

func (t Token) String() string {
	return fmt.Sprintf("[%d:%d] %s: %s", t.StartPos, t.EndPos, t.Kind, t.Value)
}

type TokenErr struct {
	Token
	message string
}

func NewErr(t Token, message string) *TokenErr {
	return &TokenErr{
		Token:   t,
		message: message,
	}
}

func (t *TokenErr) Error() string {
	return fmt.Sprintf("%s: %s", t.Token, t.message)
}

type Lexer struct {
	str   string
	pos   int
	err   error
	token Token
}

func NewLexer(str string) *Lexer {
	return &Lexer{
		str: str,
	}
}

func (l *Lexer) Err() error {
	return errors.Trace(l.err)
}

func (l *Lexer) Token() Token {
	return l.token
}

func (l *Lexer) Scan() bool {
	if l.err != nil {
		return false
	}

	if l.pos >= len(l.str) {
		return false
	}

	var (
		i int
		r rune
		t Token
	)

loop:
	for i, r = range l.str[l.pos:] {
		pos := l.pos + i

		switch r {
		case ' ', '\t', '\n':
			continue
		case '(':
			t = Token{
				Kind:     OpenParenthesis,
				Value:    string(r),
				StartPos: pos,
				EndPos:   pos + 1,
			}

			break loop

		case ')':
			t = Token{
				Kind:     CloseParenthesis,
				Value:    string(r),
				StartPos: pos,
				EndPos:   pos + 1,
			}

			break loop

		case '"':
			ret, offset := l.readUntil(l.str[pos+1:], '"')

			offset++ // Include the closing quote

			t = Token{
				Kind:     Text,
				Value:    ret,
				StartPos: pos,
				EndPos:   pos + offset,
			}

			l.pos += offset

			if l.pos >= len(l.str) {
				l.err = errors.Trace(NewErr(t, "unclosed quote"))
			}

			break loop

		default:
			ret, offset := l.readUntil(l.str[pos:], ' ', '\t', '\n', ')')
			t = Token{
				Kind:     Function,
				Value:    ret,
				StartPos: pos,
				EndPos:   pos + offset,
			}

			l.pos += offset - 1

			break loop
		}
	}

	if l.err != nil {
		return false
	}

	l.pos += i + 1
	l.token = t

	return t.Kind > Unknown
}

func (l *Lexer) readUntil(
	str string,
	runes ...rune,
) (
	ret string,
	offset int,
) {
	var prev rune

	for i, r := range str {
		if prev == '\\' {
			prev = 0
			ret += string(r)

			continue
		}

		if r == '\\' {
			prev = r

			continue
		}

		for _, match := range runes {
			if r == match {
				return ret, i
			}
		}

		ret += string(r)
	}

	return ret, len(str)
}
