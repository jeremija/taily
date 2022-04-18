package factory

import (
	"github.com/jeremija/taily/config"
	"github.com/juju/errors"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/slack"
	"github.com/nikoksr/notify/service/telegram"
)

func NewNotifier(cfgs []config.NotifyService) (notify.Notifier, error) {
	n := notify.New()

	services, err := NewNotifierServices(cfgs)
	if err != nil {
		return nil, errors.Trace(err)
	}

	n.UseServices(services...)

	return n, nil
}

func NewNotifierServices(cfgs []config.NotifyService) ([]notify.Notifier, error) {
	ret := make([]notify.Notifier, len(cfgs))

	for i, cfg := range cfgs {
		notifier, err := NewNotifierService(cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}

		ret[i] = notifier
	}

	return ret, nil
}

func NewNotifierService(cfg config.NotifyService) (notify.Notifier, error) {
	switch cfg.Type {
	case "telegram":
		service, err := telegram.New(cfg.Telegram.Token)
		if err != nil {
			return nil, errors.Trace(err)
		}

		service.AddReceivers(cfg.Telegram.Receivers...)

		return service, nil

	case "slack":
		service := slack.New(cfg.Slack.Token)

		service.AddReceivers(cfg.Slack.Receivers...)

		return service, nil

	default:
		return nil, errors.Errorf("unknown notify service: %q", cfg.Type)
	}
}
