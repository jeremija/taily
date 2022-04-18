package factory

import (
	"os"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/docker/docker/client"
	"github.com/jeremija/taily/action"
	"github.com/jeremija/taily/config"
	"github.com/jeremija/taily/formatter"
	"github.com/jeremija/taily/matcher"
	"github.com/jeremija/taily/persister"
	"github.com/jeremija/taily/processor"
	"github.com/jeremija/taily/reader"
	"github.com/jeremija/taily/types"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
)

func NewActionsMap(cfgs map[string]config.Action) (map[string]types.Action, error) {
	ret := make(map[string]types.Action, len(cfgs))

	for name, actionConfig := range cfgs {
		action, err := NewActionFromConfig(actionConfig)
		if err != nil {
			return nil, errors.Trace(err)
		}

		ret[name] = action
	}

	return ret, nil
}

func NewActionFromConfig(cfg config.Action) (types.Action, error) {
	switch cfg.Type {
	case "log":
		action, err := NewActionLogFromConfig(cfg.Log)

		return action, errors.Trace(err)
	default:
		return nil, errors.Errorf("unknown action: %q", cfg.Type)
	}
}

func NewActionLogFromConfig(config config.ActionLog) (*action.Log, error) {
	var f types.Formatter

	switch config.Format {
	case "plain":
		f = formatter.NewPlain()
	case "json":
		f = formatter.NewJSON()
	default:
		return nil, errors.Errorf("unknown format: %q", config.Format)
	}

	return action.NewLog(f, os.Stdout), nil
}

func NewProcessorsMap(
	configs map[string]config.Processor,
	actionsMap map[string]types.Action,
) (map[string]processor.Factory, error) {
	ret := make(map[string]processor.Factory, len(configs))

	for procName, procConfig := range configs {
		ret[procName] = func() (types.Processor, error) {
			processor, err := NewProcessorFromConfig(procConfig, actionsMap)

			return processor, errors.Trace(err)
		}
	}

	return ret, nil
}

// NewProcessorsFromMap reads configs and creates Processors.
func NewProcessorsFromMap(processorsMap map[string]processor.Factory, names []string) (processor.Factory, error) {
	factories := make([]processor.Factory, len(processorsMap))

	for i, name := range names {
		proc, ok := processorsMap[name]
		if !ok {
			return nil, errors.Errorf("processor configuration not found: %q", name)
		}

		factories[i] = proc
	}

	newProcessor := func() (types.Processor, error) {
		ret := make(processor.Serial, len(names))

		for i, newProcessor := range factories {
			var err error

			ret[i], err = newProcessor()
			if err != nil {
				return nil, errors.Trace(err)
			}
		}

		return ret, nil
	}

	return newProcessor, nil
}

// NewProcessorFromConfig reads config and creates a Processor.
func NewProcessorFromConfig(cfg config.Processor, actionsMap map[string]types.Action) (types.Processor, error) {
	action, ok := actionsMap[cfg.Action]
	if !ok {
		return nil, errors.Errorf("undefined action: %q", cfg.Action)
	}

	switch cfg.Type {
	case "any":
		return processor.NewAny(action), nil
	case "matcher":
		startLine, err := NewMatcherFromConfig(cfg.Matcher.Start)
		if err != nil {
			return nil, errors.Trace(err)
		}

		var endLine types.Matcher

		if cfg.Matcher.End != nil {
			endLine, err = NewMatcherFromConfig(cfg.Matcher.End)
			if err != nil {
				return nil, errors.Trace(err)
			}
		}

		return processor.NewMatcher(processor.MatcherParams{
			StartLine:  startLine,
			EndLine:    endLine,
			IncludeEnd: cfg.Matcher.IncludeEnd,
			GroupBy:    cfg.Matcher.GroupBy,
			MaxLines:   cfg.Matcher.MaxLines,
			Action:     action,
		}), nil
	default:
		return nil, errors.Errorf("unknown processor: %q", cfg.Type)
	}
}

func NewMatchersFromConfig(configs []*config.Matcher) ([]types.Matcher, error) {
	ret := make([]types.Matcher, len(configs))

	for i, config := range configs {
		m, err := NewMatcherFromConfig(config)
		if err != nil {
			return nil, errors.Trace(err)
		}

		ret[i] = m
	}

	return ret, nil
}

// NewMatcherFromConfig creates a new Mather from config. The field is just
// used for debugging.
func NewMatcherFromConfig(cfg *config.Matcher) (types.Matcher, error) {
	switch cfg.Type {
	case "string":
		return matcher.String(cfg.String), nil
	case "substring":
		return matcher.Substring(cfg.Substring), nil
	case "prefix":
		return matcher.Prefix(cfg.Prefix), nil
	case "suffix":
		return matcher.Suffix(cfg.Suffix), nil
	case "regexp":
		m, err := matcher.NewRegexp(cfg.Regexp)
		return m, errors.Trace(err)
	case "and":
		m, err := NewMatchersFromConfig(cfg.And)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return matcher.And(m), nil
	case "or":
		m, err := NewMatchersFromConfig(cfg.Or)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return matcher.Or(m), nil
	case "not":
		m, err := NewMatcherFromConfig(cfg.Not)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return matcher.Not(m), nil
	case "field":
		m, err := matcher.Field(cfg.Field.Name, cfg.Field.Regexp)

		return m, errors.Trace(err)
	case "any":
		return matcher.Any, nil
	default:
		return nil, errors.Errorf("unknown matcher: %q", cfg.Type)
	}
}

// NewPersisterFromConfig cretaes a new Persister from config.
func NewPersisterFromConfig(cfg config.Persister) (types.Persister, error) {
	switch cfg.Type {
	case "noop":
		return persister.NewNoop(), nil
	case "file":
		return persister.NewFile(cfg.File.Dir), nil
	default:
		return nil, errors.Errorf("unknown persister: %q", cfg.Type)
	}
}

// NewReaderFromConfig creates a new Reader from config.
func NewReaderFromConfig(
	logger log.Logger,
	persister types.Persister,
	newProcessor processor.Factory,
	cfg config.Reader,
) (types.Reader, error) {
	watcherParams := types.ReaderParams{
		ReaderID: cfg.ReaderID(),
		Logger:   logger,
	}

	switch cfg.Type {
	case "journald":
		params := reader.JournaldParams{
			ReaderParams: watcherParams,
			NewJournal:   sdjournal.NewJournal,
		}

		return reader.NewJournald(params), nil

	case "docker":
		cl, err := client.NewClientWithOpts(
			client.FromEnv,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}

		params := reader.DockerParams{
			ReaderParams: watcherParams,
			Client:       cl,
			Persister:    persister,
			NewProcessor: newProcessor,
		}

		return reader.NewDocker(params), nil
	default:
		return nil, errors.Errorf("unfamiliar watcher name: %q", cfg.Type)
	}
}
