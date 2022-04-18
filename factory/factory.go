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

func NewActionsMap(logger log.Logger, cfgs map[string]config.Action) (map[string]types.Action, error) {
	ret := make(map[string]types.Action, len(cfgs))

	for name, actionConfig := range cfgs {
		action, err := NewAction(logger, actionConfig)
		if err != nil {
			return nil, errors.Trace(err)
		}

		ret[name] = action
	}

	return ret, nil
}

func NewAction(logger log.Logger, cfg config.Action) (types.Action, error) {
	switch cfg.Type {
	case "log":
		action, err := NewActionLog(cfg.Log)

		return action, errors.Trace(err)
	case "notify":
		action, err := NewActionNotify(logger, cfg.Notify)

		return action, errors.Trace(err)
	default:
		return nil, errors.Errorf("unknown action: %q", cfg.Type)
	}
}

func NewFormatter(format string) (types.Formatter, error) {
	switch format {
	case "plain":
		return formatter.NewPlain(), nil
	case "json":
		return formatter.NewJSON(), nil
	default:
		return nil, errors.Errorf("unknown format: %q", format)
	}
}

func NewActionLog(cfg config.ActionLog) (*action.Log, error) {
	f, err := NewFormatter(cfg.Format)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return action.NewLog(f, os.Stdout), nil
}

func NewActionNotify(logger log.Logger, cfg config.ActionNotify) (*action.Notify, error) {
	f, err := NewFormatter(cfg.Format)
	if err != nil {
		return nil, errors.Trace(err)
	}

	notifier, err := NewNotifier(cfg.Services)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return action.NewNotify(action.NotifyParams{
		Logger:       logger,
		Formatter:    f,
		Notifier:     notifier,
		MaxTitleSize: cfg.MaxTitleSize,
		MaxBodySize:  cfg.MaxBodySize,
	}), nil
}

func NewProcessorsMap(
	cfgs map[string]config.Processor,
	actionsMap map[string]types.Action,
) (map[string]processor.Factory, error) {
	ret := make(map[string]processor.Factory, len(cfgs))

	for procName, procConfig := range cfgs {
		ret[procName] = func() (types.Processor, error) {
			processor, err := NewProcessor(procConfig, actionsMap)

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

// NewProcessor reads config and creates a Processor.
func NewProcessor(cfg config.Processor, actionsMap map[string]types.Action) (types.Processor, error) {
	action, ok := actionsMap[cfg.Action]
	if !ok {
		return nil, errors.Errorf("undefined action: %q", cfg.Action)
	}

	switch cfg.Type {
	case "any":
		return processor.NewAny(action), nil
	case "matcher":
		startLine, err := NewMatcher(cfg.Matcher.Start)
		if err != nil {
			return nil, errors.Trace(err)
		}

		var endLine types.Matcher

		if cfg.Matcher.End != nil {
			endLine, err = NewMatcher(cfg.Matcher.End)
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

func NewMatchers(cfgs []*config.Matcher) ([]types.Matcher, error) {
	ret := make([]types.Matcher, len(cfgs))

	for i, cfg := range cfgs {
		m, err := NewMatcher(cfg)
		if err != nil {
			return nil, errors.Trace(err)
		}

		ret[i] = m
	}

	return ret, nil
}

// NewMatcher creates a new Mather from config. The field is just
// used for debugging.
func NewMatcher(cfg *config.Matcher) (types.Matcher, error) {
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
		m, err := NewMatchers(cfg.And)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return matcher.And(m), nil
	case "or":
		m, err := NewMatchers(cfg.Or)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return matcher.Or(m), nil
	case "not":
		m, err := NewMatcher(cfg.Not)
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

// NewPersister cretaes a new Persister from config.
func NewPersister(cfg config.Persister) (types.Persister, error) {
	switch cfg.Type {
	case "noop":
		return persister.NewNoop(), nil
	case "file":
		return persister.NewFile(cfg.File.Dir), nil
	default:
		return nil, errors.Errorf("unknown persister: %q", cfg.Type)
	}
}

// NewReader creates a new Reader from config.
func NewReader(
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
