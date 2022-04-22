package compiler

import (
	"fmt"
	"testing"

	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
)

func compile(str string) (*Root, error) {
	root, err := New(NewLexer(str)).Compile()

	return root, errors.Trace(err)
}

func TestCompiler(t *testing.T) {
	type testCase struct {
		input   string
		want    *Root
		wantErr string
	}

	testCases := []testCase{
		{
			input:   `a`,
			want:    nil,
			wantErr: "[0:1] function: a: unexpected token",
		},
		{
			input: `(or "a" "b")`,
			want: &Root{
				Nodes: []Node{
					&FnExpr{
						Token: Token{
							Kind:     Function,
							Value:    "or",
							StartPos: 1,
							EndPos:   3,
						},
						Name: "or",
						Arguments: []Node{
							TextExpr{
								Token: Token{
									Kind:     Text,
									Value:    "a",
									StartPos: 4,
									EndPos:   6,
								},
							},
							TextExpr{
								Token: Token{
									Kind:     Text,
									Value:    "b",
									StartPos: 8,
									EndPos:   10,
								},
							},
						},
					},
				},
			},
		},
		{
			input: `(or "a" "b")`,
			want: &Root{
				Nodes: []Node{
					&FnExpr{
						Token: Token{
							Kind:     Function,
							Value:    "or",
							StartPos: 1,
							EndPos:   3,
						},
						Name: "or",
						Arguments: []Node{
							TextExpr{
								Token: Token{
									Kind:     Text,
									Value:    "a",
									StartPos: 4,
									EndPos:   6,
								},
							},
							TextExpr{
								Token: Token{
									Kind:     Text,
									Value:    "b",
									StartPos: 8,
									EndPos:   10,
								},
							},
						},
					},
				},
			},
		},
		{
			input: `(or (and "a" "b") (re "test"))`,
			want: &Root{
				Nodes: []Node{
					&FnExpr{
						Token: Token{
							Kind:     Function,
							Value:    "or",
							StartPos: 1,
							EndPos:   3,
						},
						Name: "or",
						Arguments: []Node{
							&FnExpr{
								Token: Token{
									Kind:     Function,
									Value:    "and",
									StartPos: 5,
									EndPos:   8,
								},
								Name: "and",
								Arguments: []Node{
									TextExpr{
										Token: Token{
											Kind:     Text,
											Value:    "a",
											StartPos: 9,
											EndPos:   11,
										},
									},
									TextExpr{
										Token: Token{
											Kind:     Text,
											Value:    "b",
											StartPos: 13,
											EndPos:   15,
										},
									},
								},
							},
							&FnExpr{
								Token: Token{
									Kind:     Function,
									Value:    "re",
									StartPos: 19,
									EndPos:   21,
								},
								Name: "re",
								Arguments: []Node{
									TextExpr{
										Token: Token{
											Kind:     Text,
											Value:    "test",
											StartPos: 22,
											EndPos:   27,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		descr := fmt.Sprintf("%d. %s", i, tc.input)

		nodes, err := compile(tc.input)
		if tc.wantErr != "" {
			if assert.Error(t, err, descr) {
				assert.Contains(t, err.Error(), tc.wantErr, descr)
			}

			continue
		}

		assert.NoError(t, err, descr)
		assert.Equal(t, tc.want, nodes, descr)
	}
}
