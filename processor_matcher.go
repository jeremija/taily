package taily

import (
	"context"
	"time"

	"github.com/juju/errors"
)

type ProcessorMatcher struct {
	params   ProcessorMatcherParams
	state    MatcherState
	messages []Message
}

func NewProcessorMatcher(params ProcessorMatcherParams) *ProcessorMatcher {
	return &ProcessorMatcher{
		params: params,
	}
}

type ProcessorMatcherParams struct {
	Start  Matcher // Start is required.
	End    Matcher // End is optional for multiline matching.
	Action Action  // Action to perform upon a match is found.
}

// Assert that ProcessorLog implements Processor.
var _ Processor = &ProcessorMatcher{}

type MatcherState int

const (
	MatcherClosed MatcherState = iota
	MatcherOpen
)

func (p *ProcessorMatcher) performAction(ctx context.Context) error {
	messages := p.messages

	p.messages = nil
	p.state = MatcherClosed

	// TODO do not call PerformAction synchronously as it will block the main
	// event loop.
	err := p.params.Action.PerformAction(ctx, messages)

	return errors.Trace(err)
}

// ProcessMessage implements Processor.
func (p *ProcessorMatcher) ProcessMessage(ctx context.Context, message Message) error {
	switch p.state {
	case MatcherClosed:
		if p.params.Start.MatchMessage(message) {
			p.messages = append(p.messages, message)
			p.state = MatcherOpen

			// When End Matcher is not defined, perform the action immediately.
			if p.params.End == nil {
				p.performAction(ctx)
			}
		}
	case MatcherOpen:
		p.messages = append(p.messages, message)

		if p.params.End.MatchMessage(message) {
			p.performAction(ctx)
		}
	}

	return nil
}

// Tick implements Processor.
func (p *ProcessorMatcher) Tick(ctx context.Context, now time.Time) error {
	var err error

	if p.state == MatcherOpen {
		// TODO allow some timeout before cleaning up.
		err = p.performAction(ctx)
	}

	return errors.Trace(err)
}
