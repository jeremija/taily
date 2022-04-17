package processor

import (
	"context"
	"time"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
)

type Matcher struct {
	params   MatcherParams
	state    matchState
	messages []types.Message
}

func NewMatcher(params MatcherParams) *Matcher {
	return &Matcher{
		params: params,
	}
}

type MatcherParams struct {
	StartLine types.Matcher // Start is required.
	EndLine   types.Matcher // End is optional for multiline matching.
	Action    types.Action  // Action to perform upon a match is found.
}

// Assert that Matcherimplements types.Processor.
var _ types.Processor = &Matcher{}

type matchState int

const (
	matchClosed matchState = iota
	matcherOpen
)

func (p *Matcher) performAction(ctx context.Context) error {
	messages := p.messages

	p.messages = nil
	p.state = matchClosed

	// TODO do not call PerformAction synchronously as it will block the main
	// event loop.
	err := p.params.Action.PerformAction(ctx, messages)

	return errors.Trace(err)
}

// ProcessMessage implements Processor.
func (p *Matcher) ProcessMessage(ctx context.Context, message types.Message) error {
	switch p.state {
	case matchClosed:
		if p.params.StartLine.MatchMessage(message) {
			p.messages = append(p.messages, message)
			p.state = matcherOpen

			// When End Matcher is not defined, perform the action immediately.
			if p.params.EndLine == nil {
				p.performAction(ctx)
			}
		}
	case matcherOpen:
		p.messages = append(p.messages, message)

		if p.params.EndLine.MatchMessage(message) {
			p.performAction(ctx)
		}
	}

	return nil
}

// Tick implements Processor.
func (p *Matcher) Tick(ctx context.Context, now time.Time) error {
	var err error

	if p.state == matcherOpen {
		// TODO allow some timeout before cleaning up.
		err = p.performAction(ctx)
	}

	return errors.Trace(err)
}
