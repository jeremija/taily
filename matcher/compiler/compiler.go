package compiler

import (
	"github.com/juju/errors"
)

type Compiler struct {
	lexer *Lexer
}

func New(l *Lexer) *Compiler {
	return &Compiler{
		lexer: l,
	}
}

func (c *Compiler) Compile() (*Root, error) {
	root := &Root{}

	st := state{
		stack: []Node{root},
	}

	scanFunction := func(t Token) error {
		if !c.lexer.Scan() {
			return errors.Trace(c.lexer.Err())
		}

		t = c.lexer.Token()
		switch t.Kind {
		case Function:
			f := &FnExpr{
				Name:  t.Value,
				Token: t,
			}

			if !st.push(f) {
				return errors.Trace(NewErr(t, "invalid argument"))
			}

			return nil
		default:
			return errors.Trace(NewErr(t, "expected a function"))
		}
	}

	for c.lexer.Scan() {
		t := c.lexer.Token()

		switch t.Kind {
		case OpenParenthesis:
			if err := scanFunction(t); err != nil {
				return nil, errors.Trace(err)
			}

		case CloseParenthesis:
			if _, ok := st.pop(); !ok {
				return nil, errors.Trace(NewErr(t, "unexpected )"))
			}
		case Text:
			expr := TextExpr{Token: t}

			if !st.addNode(expr) {
				return nil, errors.Trace(NewErr(t, "unexpected text"))
			}
		default:
			return nil, errors.Trace(NewErr(t, "unexpected token"))
		}
	}

	return root, nil
}

type state struct {
	stack []Node
}

func (s *state) tail() Node {
	return s.stack[len(s.stack)-1]
}

func (s *state) push(node Node) bool {
	tail := s.tail()

	n, ok := tail.(interface {
		AddNode(node Node)
	})
	if !ok {
		return false
	}

	n.AddNode(node)

	s.stack = append(s.stack, node)

	return true
}

func (s *state) pop() (Node, bool) {
	if len(s.stack) == 1 {
		return nil, false
	}

	removedNode := s.tail()

	s.stack = s.stack[:len(s.stack)-1]

	return removedNode, true
}

func (s *state) addNode(node Node) bool {
	n, ok := s.tail().(interface {
		AddNode(node Node)
	})
	if !ok {
		return false
	}

	n.AddNode(node)

	return true
}

type node struct{}

func (n node) internal() {}

type Node interface {
	internal()
}

type Root struct {
	node
	Nodes []Node
}

func (r *Root) AddNode(node Node) {
	r.Nodes = append(r.Nodes, node)
}

type FnExpr struct {
	node
	Name      string
	Token     Token
	Arguments []Node
}

func (f *FnExpr) AddNode(node Node) {
	f.Arguments = append(f.Arguments, node)
}

type TextExpr struct {
	node
	Token
}
