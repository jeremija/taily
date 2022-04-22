package compiler

import (
	"fmt"
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	type testCase struct {
		input   string
		want    []Token
		wantErr string
	}

	testCases := []testCase{
		{
			input:   `"`,
			wantErr: "unclosed quote",
		},
		{
			input: `" a "`,
			want: []Token{
				{
					Kind:     Text,
					Value:    ` a `,
					StartPos: 0,
					EndPos:   4,
				},
			},
		},
		{
			input:   `"\"`,
			wantErr: "unclosed quote",
		},
		{
			input: `" \"escaped \\ \string "`,
			want: []Token{
				{
					Kind:     Text,
					Value:    ` "escaped \ string `,
					StartPos: 0,
					EndPos:   23,
				},
			},
		},
		{
			input: `  " a "  `,
			want: []Token{
				{
					Kind:     Text,
					Value:    ` a `,
					StartPos: 2,
					EndPos:   6,
				},
			},
		},
		{
			input: `a`,
			want: []Token{
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 0,
					EndPos:   1,
				},
			},
		},
		{
			input: ` a`,
			want: []Token{
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 1,
					EndPos:   2,
				},
			},
		},
		{
			input: `a `,
			want: []Token{
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 0,
					EndPos:   1,
				},
			},
		},
		{
			input: ` a `,
			want: []Token{
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 1,
					EndPos:   2,
				},
			},
		},
		{
			input: `a b`,
			want: []Token{
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 0,
					EndPos:   1,
				},
				{
					Kind:     Function,
					Value:    `b`,
					StartPos: 2,
					EndPos:   3,
				},
			},
		},
		{
			input: `(a b`,
			want: []Token{
				{
					Kind:     OpenParenthesis,
					Value:    `(`,
					StartPos: 0,
					EndPos:   1,
				},
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 1,
					EndPos:   2,
				},
				{
					Kind:     Function,
					Value:    `b`,
					StartPos: 3,
					EndPos:   4,
				},
			},
		},
		{
			input: `(a b)`,
			want: []Token{
				{
					Kind:     OpenParenthesis,
					Value:    `(`,
					StartPos: 0,
					EndPos:   1,
				},
				{
					Kind:     Function,
					Value:    `a`,
					StartPos: 1,
					EndPos:   2,
				},
				{
					Kind:     Function,
					Value:    `b`,
					StartPos: 3,
					EndPos:   4,
				},
				{
					Kind:     CloseParenthesis,
					Value:    `)`,
					StartPos: 4,
					EndPos:   5,
				},
			},
		},
	}

	for i, tc := range testCases {
		descr := fmt.Sprintf("%d. %s", i, tc.input)

		tokens, err := scan(tc.input)
		if tc.wantErr != "" {
			if assert.Error(t, err, descr) {
				assert.Contains(t, err.Error(), tc.wantErr, descr)
			}

			continue
		}

		assert.NoError(t, err, descr)
		assert.Equal(t, tc.want, tokens, descr)
	}
}

func scan(str string) ([]Token, error) {
	l := NewLexer(str)
	tokens, err := scanTokens(l)

	return tokens, errors.Trace(err)
}

func scanTokens(l *Lexer) ([]Token, error) {
	var tokens []Token

	for l.Scan() {
		t := l.Token()
		tokens = append(tokens, t)
	}

	return tokens, errors.Trace(l.Err())
}
