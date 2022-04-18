package pipeline_test

import (
	"context"
	"testing"
	"time"

	"github.com/jeremija/taily/action"
	"github.com/jeremija/taily/config"
	"github.com/jeremija/taily/factory"
	"github.com/jeremija/taily/formatter"
	"github.com/jeremija/taily/mock"
	"github.com/jeremija/taily/persister"
	"github.com/jeremija/taily/pipeline"
	"github.com/jeremija/taily/processor"
	"github.com/jeremija/taily/types"
	"github.com/jeremija/taily/watcher"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
	"github.com/stretchr/testify/assert"
)

func TestPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logger := log.NewFromEnv("TAILY_LOG")

	reader := mock.NewReader("test")
	notifier := mock.NewNotifier()

	w := watcher.New(watcher.Params{
		Logger:    logger,
		Persister: persister.NewNoop(),
		Reader:    reader,
	})

	newProcessor := func() (types.Processor, error) {
		p := processor.NewMatcher(processor.MatcherParams{
			StartLine: MustMatcher(&config.Matcher{
				Type:      "substring",
				Substring: "test",
			}),
			Action: action.NewNotify(action.NotifyParams{
				Logger:       logger,
				Formatter:    formatter.NewPlain(),
				Notifier:     notifier,
				MaxTitleSize: 50,
				MaxBodySize:  100,
			}),
		})

		return p, nil
	}

	pline := pipeline.New(pipeline.Params{
		Logger:       log.New(),
		Watcher:      w,
		NewProcessor: newProcessor,
		BufferSize:   0,
	})

	errCh := make(chan error, 1)

	go func() {
		errCh <- errors.Trace(pline.ProcessPipeline(ctx))
	}()

	readCtx, err := reader.Accept(ctx)
	assert.NoError(t, err)

	defer readCtx.Close()

	err = readCtx.MockMessage(ctx, types.NewMessage(
		time.Now(),
		reader.ReaderID(),
		"this is a test message",
		nil,
	))
	assert.NoError(t, err)

	notification, err := notifier.Receive(ctx)
	assert.NoError(t, err)

	assert.Equal(t, mock.Notification{
		Title: "this is a test message",
		Body:  "this is a test message\n",
	}, notification)

	readCtx.Close()

	err = <-errCh
	assert.NoError(t, err)
}

func MustMatcher(cfg *config.Matcher) types.Matcher {
	m, err := factory.NewMatcher(cfg)
	if err != nil {
		panic(err)
	}

	return m
}
