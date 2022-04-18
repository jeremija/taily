package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

type Matcher struct {
	params  MatcherParams
	matches map[string]*match
}

func NewMatcher(params MatcherParams) *Matcher {
	return &Matcher{
		matches: map[string]*match{},
		params:  params,
	}
}

type MatcherParams struct {
	StartLine  types.Matcher // Start is required.
	EndLine    types.Matcher // End is optional for multiline matching.
	IncludeEnd bool
	MaxLines   int          // MaxLines is max nmber of lines lines to match.
	Action     types.Action // Action to perform upon a match is found.
	GroupBy    []string     // Fields to group by.
}

// Assert that Matcherimplements types.Processor.
var _ types.Processor = &Matcher{}

func (p *Matcher) performAction(ctx context.Context, m *match) error {
	// TODO do not call PerformAction synchronously as it will block the main
	// event loop.
	err := p.params.Action.PerformAction(ctx, m.messages)

	return errors.Trace(err)
}

// ProcessMessage implements Processor.
func (p *Matcher) ProcessMessage(ctx context.Context, message types.Message) error {
	key := groupKey(p.params.GroupBy, message)

	m, ok := p.matches[key]
	if !ok {
		if p.params.StartLine.MatchMessage(message) {
			m = &match{
				time:     time.Now(), // TODO mock
				messages: []types.Message{message},
			}

			// When End Matcher is not defined, perform the action immediately.
			if p.params.EndLine == nil {
				p.performAction(ctx, m)
				return nil
			}

			p.matches[key] = m
		}

		return nil
	}

	isEnd := p.params.EndLine.MatchMessage(message)

	if !isEnd || p.params.IncludeEnd {
		m.messages = append(m.messages, message)
	}

	switch {
	case p.params.MaxLines > 0 && len(m.messages) > p.params.MaxLines:
		delete(p.matches, key)
		p.performAction(ctx, m)
	case isEnd:
		delete(p.matches, key)
		p.performAction(ctx, m)
	}

	return nil
}

// Tick implements Processor.
func (p *Matcher) Tick(ctx context.Context, now time.Time) error {
	errs := make([]string, 0, len(p.matches))

	for k, m := range p.matches {
		delete(p.matches, k)

		if err := p.performAction(ctx, m); err != nil {
			errs = append(errs, fmt.Sprintf("%+v", err))
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("tick failed: \n%s", strings.Join(errs, "\n"))
	}

	return nil
}
