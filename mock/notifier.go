package mock

import (
	"context"

	"github.com/juju/errors"
	"github.com/nikoksr/notify"
)

type Notification struct {
	Title string
	Body  string
}

type Notifier struct {
	messages chan Notification
}

func NewNotifier() *Notifier {
	return &Notifier{
		messages: make(chan Notification),
	}
}

var _ notify.Notifier = &Notifier{}

// Send implements notify.Notifier
func (n *Notifier) Send(ctx context.Context, title, body string) error {
	notification := Notification{
		Title: title,
		Body:  body,
	}

	select {
	case n.messages <- notification:
		return nil
	case <-ctx.Done():
		return errors.Trace(ctx.Err())
	}
}

func (n *Notifier) Receive(ctx context.Context) (Notification, error) {
	select {
	case notification := <-n.messages:
		return notification, nil
	case <-ctx.Done():
		return Notification{}, errors.Trace(ctx.Err())
	}
}
