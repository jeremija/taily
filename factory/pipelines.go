package factory

import (
	"github.com/jeremija/taily/config"
	"github.com/jeremija/taily/pipeline"
	"github.com/jeremija/taily/types"
	"github.com/jeremija/taily/watcher"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

func NewPipelines(logger log.Logger, cfg *config.Config) ([]*pipeline.Pipeline, error) {
	actionsMap, err := NewActionsMap(cfg.Actions)
	if err != nil {
		return nil, errors.Trace(err)
	}

	persister, err := NewPersisterFromConfig(cfg.Persister)
	if err != nil {
		return nil, errors.Trace(err)
	}

	processorsMap, err := NewProcessorsMap(cfg.Processors, actionsMap)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ret := make([]*pipeline.Pipeline, len(cfg.Readers))

	// errCh := make(chan error, len(cfg.Watchers))

	readerIDs := make(map[types.ReaderID]struct{}, len(cfg.Readers))

	for i, config := range cfg.Readers {
		readerID := config.ReaderID()

		if _, ok := readerIDs[readerID]; ok {
			return nil, errors.Errorf("duplicate reader ID: %q", readerID)
		}

		readerIDs[readerID] = struct{}{}

		newProcessor, err := NewProcessorsFromMap(processorsMap, config.Processors)
		if err != nil {
			return nil, errors.Trace(err)
		}

		r, err := NewReaderFromConfig(logger, persister, newProcessor, config)
		if err != nil {
			return nil, errors.Trace(err)
		}

		w := watcher.New(watcher.Params{
			Persister:    persister,
			Reader:       r,
			Logger:       logger,
			InitialState: config.InitialState,
		})

		pline := pipeline.New(pipeline.Params{
			Logger:       logger,
			Watcher:      w,
			NewProcessor: newProcessor,
			BufferSize:   0,
		})

		ret[i] = pline
	}

	return ret, nil
}
