package taily

import (
	"io"
	"os"

	"github.com/coreos/go-systemd/v22/sdjournal"
	"github.com/docker/docker/client"
	"github.com/juju/errors"
	"github.com/peer-calls/log"
	"gopkg.in/yaml.v3"
)

// Config describes the main YAML config file.
type Config struct {
	Watchers   []WatcherConfig            `yaml:"watchers"`
	Actions    map[string]ActionConfig    `yaml:"actions"`
	Processors map[string]ProcessorConfig `yaml:"processors"`
	Persister  PersisterConfig            `yaml:"persister"`
}

// NewConfigFromFile opens the filename and tries to decode the config YAML.
func NewConfigFromFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer f.Close()

	config, err := NewConfigFromReader(f)

	return config, errors.Trace(err)
}

// NewConfigFromReader decodes the YAML config from reader.
func NewConfigFromReader(reader io.Reader) (*Config, error) {
	var config Config

	err := yaml.NewDecoder(reader).Decode(&config)

	return &config, errors.Trace(err)
}

// NewConfigFromEnv reads the YAML config from the environment variable.
func NewConfigFromEnv(env string) (*Config, error) {
	config, err := NewConfigFromString(os.Getenv(env))

	return config, errors.Trace(err)
}

// NewConfigFromString reads the YAML config from a string value.
func NewConfigFromString(str string) (*Config, error) {
	var config Config

	err := yaml.Unmarshal([]byte(str), &config)

	return &config, errors.Trace(err)
}

// WatcherConfig contains configuration for a specific watcher.
type WatcherConfig struct {
	Type         ReaderType `yaml:"type"`
	Processors   []string   `yaml:"processors"`
	InitialState State      `yaml:"initial_state"`
}

// ProcessorConfig contains configuration for a specific processor.
type ProcessorConfig struct {
	Type    string                 `yaml:"type"`
	Action  string                 `yaml:"action"`
	Matcher ProcessorMatcherConfig `yaml:"log"`
}

// ProcessorMatcherConfig contains cofiguration for ProcessorMatcher.
type ProcessorMatcherConfig struct {
	Start MatcherConfig `yaml:"start"`
	End   MatcherConfig `yaml:"end"`
}

// MatcherConfig contains configuration for Matcher.
type MatcherConfig struct {
	Type    string `yaml:"type"`
	Pattern string `yaml:"pattern"`
}

// ProcessorLogConfig contains configuration for ProcessorLog.
type ProcessorAnyConfig struct {
	Action string `json:"action"`
}

// PersisterConfig contains configuration for Persister.
type PersisterConfig struct {
	Type string              `yaml:"type"`
	File PersisterFileConfig `yaml:"file"`
}

// PersisterFileConfig contains configuration for PersisterFile.
type PersisterFileConfig struct {
	Dir string `yaml:"dir"`
}

func NewActionsMap(configs map[string]ActionConfig) (map[string]Action, error) {
	ret := make(map[string]Action, len(configs))

	for name, actionConfig := range configs {
		action, err := NewActionFromConfig(actionConfig)
		if err != nil {
			return nil, errors.Trace(err)
		}

		ret[name] = action
	}

	return ret, nil
}

func NewActionFromConfig(config ActionConfig) (Action, error) {
	switch config.Type {
	case "log":
		action, err := NewActionLogFromConfig(config.Log)

		return action, errors.Trace(err)
	default:
		return nil, errors.Errorf("unknown action: %q", config.Type)
	}
}

func NewActionLogFromConfig(config ActionLogConfig) (*ActionLog, error) {
	var f Formatter

	switch config.Format {
	case "plain":
		f = NewFormatterPlain()
	case "json":
		f = NewFormatterJSON()
	default:
		return nil, errors.Errorf("unknown format: %q", config.Format)
	}

	return NewActionLog(f, os.Stdout), nil
}

func NewProcessorsMap(
	configs map[string]ProcessorConfig,
	actionsMap map[string]Action,
) (map[string]ProcessorFactory, error) {
	ret := make(map[string]ProcessorFactory, len(configs))

	for procName, procConfig := range configs {
		ret[procName] = func() (Processor, error) {
			processor, err := NewProcessorFromConfig(procConfig, actionsMap)

			return processor, errors.Trace(err)
		}
	}

	return ret, nil
}

// NewProcessorsFromMap reads configs and creates Processors.
func NewProcessorsFromMap(processorsMap map[string]ProcessorFactory, names []string) (ProcessorFactory, error) {
	factories := make([]ProcessorFactory, len(processorsMap))

	for i, name := range names {
		proc, ok := processorsMap[name]
		if !ok {
			return nil, errors.Errorf("processor configuration not found: %q", name)
		}

		factories[i] = proc
	}

	newProcessor := func() (Processor, error) {
		ret := make(Processors, len(names))

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
func NewProcessorFromConfig(config ProcessorConfig, actionsMap map[string]Action) (Processor, error) {
	action, ok := actionsMap[config.Action]
	if !ok {
		return nil, errors.Errorf("undefined action: %q", config.Action)
	}

	switch config.Type {
	case "any":
		return NewProcessorAny(action), nil
	case "matcher":
		start, err := NewMatcherFromConfig(config.Matcher.Start)
		if err != nil {
			return nil, errors.Trace(err)
		}

		end, err := NewMatcherFromConfig(config.Matcher.End)
		if err != nil {
			return nil, errors.Trace(err)
		}

		return NewProcessorMatcher(ProcessorMatcherParams{
			Start:  start,
			End:    end,
			Action: action,
		}), nil
	default:
		return nil, errors.Errorf("unknown processor: %q", config.Type)
	}
}

func NewMatcherFromConfig(config MatcherConfig) (Matcher, error) {
	switch config.Type {
	case "substring":
		return NewMatcherSubstring(config.Pattern), nil
	case "regexp":
		m, err := NewMatcherRegexp(config.Pattern)
		return m, errors.Trace(err)
	default:
		return nil, errors.Errorf("unknown matcher: %q", config.Type)
	}
}

// NewPersisterFromConfig cretaes a new Persister from config.
func NewPersisterFromConfig(config PersisterConfig) (Persister, error) {
	switch config.Type {
	case "noop":
		return NewPersisterNoop(), nil
	case "file":
		return NewPersisterFile(config.File.Dir), nil
	default:
		return nil, errors.Errorf("unknown persister: %q", config.Type)
	}
}

// NewReaderFromConfig creates a new Reader from config.
func NewReaderFromConfig(
	logger log.Logger,
	persister Persister,
	newProcessor ProcessorFactory,
	config WatcherConfig,
) (Reader, error) {
	watcherParams := ReaderParams{
		ReaderID: ReaderID(config.Type),
		Logger:   logger,
	}

	switch config.Type {
	case ReaderTypeJournald:
		params := JournaldParams{
			ReaderParams: watcherParams,
			NewJournal:   sdjournal.NewJournal,
		}

		return NewJournald(params), nil

	case ReaderTypeDocker:
		cl, err := client.NewClientWithOpts(
			client.FromEnv,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}

		params := DockerParams{
			ReaderParams: watcherParams,
			Client:       cl,
			Persister:    persister,
			NewProcessor: newProcessor,
		}

		return NewDocker(params), nil
	default:
		return nil, errors.Errorf("unfamiliar watcher name: %q", config.Type)
	}
}

type ActionConfig struct {
	Type string
	Log  ActionLogConfig
}

type ActionLogConfig struct {
	Format string
}

// ReaderType describes a watcher.
type ReaderType string

const (
	// ReaderTypeDocker describes the Docker watcher.
	ReaderTypeDocker ReaderType = "docker"
	// ReaderTypeJournald describes the Journald watcher.
	ReaderTypeJournald ReaderType = "journald"
)
