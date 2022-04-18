package action

import (
	"context"

	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/nikoksr/notify"
	"github.com/peer-calls/log"
)

type Notify struct {
	params NotifyParams
	pool   *pool
}

type NotifyParams struct {
	Logger       log.Logger
	Formatter    types.Formatter
	Notifier     notify.Notifier
	MaxTitleSize int
	MaxBodySize  int
}

func NewNotify(params NotifyParams) *Notify {
	params.Logger = params.Logger.WithNamespaceAppended("action_notify")

	return &Notify{
		params: params,
		pool:   newPool(),
	}
}

func (n *Notify) PerformAction(ctx context.Context, messages []types.Message) error {
	if len(messages) == 0 {
		return errors.Errorf("no messages")
	}

	buffer := n.pool.Get()
	defer n.pool.Put(buffer)

	if err := formatMessage(n.params.Formatter, messages, buffer); err != nil {
		return errors.Trace(err)
	}

	title := messages[0].Text()
	body := buffer.String()

	title = limitSize(title, n.params.MaxTitleSize)
	body = limitSize(body, n.params.MaxBodySize)

	n.params.Logger.Info("Sending notification with title", log.Ctx{
		"notification_title": title,
	})

	// TODO maybe use strings.Builder instead of buffer.}
	if err := n.params.Notifier.Send(ctx, title, body); err != nil {
		return errors.Trace(err)
	}

	return nil
}

func limitSize(str string, maxSize int) string {
	if maxSize == 0 {
		return str
	}

	if len(str) > maxSize {
		return str[:maxSize]
	}

	return str
}
