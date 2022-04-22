package matcher

import (
	"fmt"

	"github.com/jeremija/taily/matcher/compiler"
	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

func MatcherFromFn(fn *compiler.FnExpr) (types.Matcher, error) {
	switch fn.Name {
	case "and":
		if len(fn.Arguments) == 0 {
			return nil, compiler.NewErr(fn.Token, "and expects at least one argument")
		}

		and := make(And, 0, len(fn.Arguments))

		for _, arg := range fn.Arguments {
			switch tArg := arg.(type) {
			case *compiler.FnExpr:
				m, err := MatcherFromFn(tArg)
				if err != nil {
					return nil, errors.Trace(err)
				}

				and = append(and, m)
			default:
				return nil, compiler.NewErr(fn.Token, "expected a function expression")
			}
		}

		return and, nil
	case "any":
		if len(fn.Arguments) > 0 {
			return nil, compiler.NewErr(fn.Token, "any expects no arguments")
		}

		return Any, nil
	case "field":
		if len(fn.Arguments) != 2 {
			return nil, compiler.NewErr(fn.Token, "field expectes exactly two arguments")
		}

		arg0, ok := fn.Arguments[0].(compiler.TextExpr)
		if !ok {
			return nil, compiler.NewErr(fn.Token, "argument 0 should be a text expr")
		}

		arg1, ok := fn.Arguments[1].(compiler.TextExpr)
		if !ok {
			return nil, compiler.NewErr(fn.Token, "argument 1 should be a text expr")
		}

		f, err := Field(arg0.Value, arg1.Value)
		return f, errors.Trace(err)

	case "not":
		if len(fn.Arguments) != 1 {
			return nil, compiler.NewErr(fn.Token, "not accepts only a sinle argument")
		}

		arg := fn.Arguments[0]

		switch arg := arg.(type) {
		case *compiler.FnExpr:
			m, err := MatcherFromFn(arg)
			if err != nil {
				return nil, errors.Trace(err)
			}

			return Not(m), nil
		default:
			return nil, compiler.NewErr(fn.Token, "expected a function expression")
		}
	case "or":
		if len(fn.Arguments) == 0 {
			return nil, compiler.NewErr(fn.Token, "or expects at least one argument")
		}

		or := make(Or, 0, len(fn.Arguments))

		for _, arg := range fn.Arguments {

			switch tArg := arg.(type) {
			case *compiler.FnExpr:
				m, err := MatcherFromFn(tArg)
				if err != nil {
					return nil, errors.Trace(err)
				}

				or = append(or, m)
			default:
				return nil, compiler.NewErr(fn.Token, "expected a function expression")
			}
		}

		return or, nil
	case "eq", "string":
		if len(fn.Arguments) != 1 {
			return nil, compiler.NewErr(fn.Token, "string accepts only a single argument")
		}

		arg := fn.Arguments[0]

		switch arg := arg.(type) {
		case compiler.TextExpr:
			return String(arg.Value), nil
		default:
			return nil, compiler.NewErr(fn.Token, "expected a text expression")
		}
	case "substring":
		if len(fn.Arguments) != 1 {
			return nil, compiler.NewErr(fn.Token, "substring accepts only a single argument")
		}

		arg := fn.Arguments[0]

		switch arg := arg.(type) {
		case compiler.TextExpr:
			return Substring(arg.Value), nil
		default:
			return nil, compiler.NewErr(fn.Token, "expected a text expression")
		}
	case "pre", "prefix":
		if len(fn.Arguments) != 1 {
			return nil, compiler.NewErr(fn.Token, "string accepts only a single argument")
		}

		arg := fn.Arguments[0]

		switch arg := arg.(type) {
		case compiler.TextExpr:
			return Prefix(arg.Value), nil
		default:
			return nil, compiler.NewErr(fn.Token, "expected a text expression")
		}
	case "suf", "suffix":
		if len(fn.Arguments) != 1 {
			return nil, compiler.NewErr(fn.Token, "string accepts only a single argument")
		}

		arg := fn.Arguments[0]

		switch arg := arg.(type) {
		case compiler.TextExpr:
			return Suffix(arg.Value), nil
		default:
			return nil, compiler.NewErr(fn.Token, "expected a text expression")
		}
	case "re", "regexp":
		if len(fn.Arguments) != 1 {
			return nil, compiler.NewErr(fn.Token, "string accepts only a single argument")
		}

		arg := fn.Arguments[0]

		switch arg := arg.(type) {
		case compiler.TextExpr:
			m, err := NewRegexp(arg.Value)

			return m, errors.Trace(err)
		default:
			return nil, compiler.NewErr(fn.Token, "expected a text expression")
		}
	default:
		return nil, compiler.NewErr(fn.Token, fmt.Sprintf("unfamiliar function: %s", fn.Name))
	}
}

func BuildMatcher(root *compiler.Root) (types.Matcher, error) {
	if len(root.Nodes) == 0 {
		return nil, errors.Errorf("no rules defined")
	}

	ret := make(And, len(root.Nodes))

	for i, node := range root.Nodes {
		switch t := node.(type) {
		case *compiler.FnExpr:
			m, err := MatcherFromFn(t)
			if err != nil {
				return nil, errors.Trace(err)
			}

			ret[i] = m
		default:
			return nil, errors.Errorf("expected a function expression, node: %d", i)
		}
	}

	if len(ret) == 1 {
		return ret[0], nil
	}

	return ret, nil
}

func Compile(str string) (types.Matcher, error) {
	l := compiler.NewLexer(str)
	root, err := compiler.New(l).Compile()
	if err != nil {
		return nil, errors.Trace(err)
	}

	m, err := BuildMatcher(root)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return m, nil
}
