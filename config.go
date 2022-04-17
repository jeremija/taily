package guardlog

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
	Watchers  []WatcherConfig `yaml:"watchers"`
	Persister PersisterConfig `yaml:"persister"`
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
	Name       ReaderName        `yaml:"name"`
	Processors []ProcessorConfig `yaml:"processors"`
}

// ProcessorConfig contains configuration for a specific processor.
type ProcessorConfig struct {
	Name string             `yaml:"name"`
	Log  ProcessorLogConfig `yaml:"log"`
}

// ProcessorLogConfig contains configuration for ProcessorLog.
type ProcessorLogConfig struct {
	Format string `json:"format"`
}

// PersisterConfig contains configuration for Persister.
type PersisterConfig struct {
	Name string              `yaml:"name"`
	File PersisterFileConfig `yaml:"file"`
}

// PersisterFileConfig contains configuration for PersisterFile.
type PersisterFileConfig struct {
	Dir string `yaml:"dir"`
}

// NewProcessorsFromConfig reads configs and creates Processors.
func NewProcessorsFromConfig(configs []ProcessorConfig) (Processors, error) {
	processors := make(Processors, len(configs))

	for i, procConfig := range configs {
		var err error

		if processors[i], err = NewProcessorFromConfig(procConfig); err != nil {
			return nil, errors.Trace(err)
		}
	}

	return processors, nil
}

// NewProcessorFromConfig reads config and creates a Processor.
func NewProcessorFromConfig(config ProcessorConfig) (Processor, error) {
	switch config.Name {
	case "log":
		p, err := NewProcessorLogFromConfig(config.Log)
		return p, errors.Trace(err)
	default:
		return nil, errors.Errorf("unknown processor: %q", config.Name)
	}
}

// NewProcessorLogFromConfig creates a ProcessorLog from config.
func NewProcessorLogFromConfig(config ProcessorLogConfig) (Processor, error) {
	var f Formatter

	switch config.Format {
	case "plain":
		f = NewFormatterPlain()
	case "json":
		f = NewFormatterJSON()
	default:
		return nil, errors.Errorf("unknown format: %q", config.Format)
	}

	return NewProcessorLog(f, os.Stdout), nil
}

// NewPersisterFromConfig cretaes a new Persister from config.
func NewPersisterFromConfig(config PersisterConfig) (Persister, error) {
	switch config.Name {
	case "noop":
		return NewPersisterNoop(), nil
	case "file":
		return NewPersisterFile(config.File.Dir), nil
	default:
		return nil, errors.Errorf("unknown persister: %q", config.Name)
	}
}

// NewReaderFromConfig creates a new Reader from config.
func NewReaderFromConfig(
	logger log.Logger,
	persister Persister,
	config WatcherConfig,
) (Reader, error) {
	watcherParams := ReaderParams{
		ReaderID: ReaderID(config.Name),
		Logger:   logger,
	}

	switch config.Name {
	case ReaderNameJournald:
		journal, err := sdjournal.NewJournal()
		if err != nil {
			return nil, errors.Trace(err)
		}

		params := JournaldParams{
			ReaderParams: watcherParams,
			Journal:      journal,
		}

		return NewJournald(params), nil

	case ReaderNameDocker:
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
		}

		return NewDocker(params), nil
	default:
		return nil, errors.Errorf("unfamiliar watcher name: %q", config.Name)
	}
}

// ReaderName describes a watcher.
type ReaderName string

const (
	// ReaderNameDocker describes the Docker watcher.
	ReaderNameDocker ReaderName = "docker"
	// ReaderNameJournald describes the Journald watcher.
	ReaderNameJournald ReaderName = "journald"
)
