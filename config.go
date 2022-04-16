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

type Config struct {
	Watchers  []WatcherConfig `yaml:"watchers"`
	Persister PersisterConfig `yaml:"persister"`
}

func NewConfigFromFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer f.Close()

	config, err := NewConfigFromReader(f)

	return config, errors.Trace(err)
}

func NewConfigFromReader(reader io.Reader) (*Config, error) {
	var config Config

	err := yaml.NewDecoder(reader).Decode(&config)

	return &config, errors.Trace(err)
}

func NewConfigFromEnv(env string) (*Config, error) {
	config, err := NewConfigFromString(os.Getenv(env))

	return config, errors.Trace(err)
}

func NewConfigFromString(str string) (*Config, error) {
	var config Config

	err := yaml.Unmarshal([]byte(str), &config)

	return &config, errors.Trace(err)
}

type WatcherConfig struct {
	Name       WatcherName       `yaml:"name"`
	Processors []ProcessorConfig `yaml:"processors"`
}

type ProcessorConfig struct {
	Name string `yaml:"name"`
}

type PersisterConfig struct {
	Name string              `yaml:"name"`
	File PersisterFileConfig `yaml:"file"`
}

type PersisterFileConfig struct {
	Dir string `yaml:"dir"`
}

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

func NewProcessorFromConfig(config ProcessorConfig) (Processor, error) {
	switch config.Name {
	case "log":
		return NewProcessorLog(), nil
	default:
		return nil, errors.Errorf("unknown processor: %q", config.Name)
	}
}

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

func NewWatcherFromConfig(
	logger log.Logger,
	persister Persister,
	config WatcherConfig,
) (Watcher, error) {
	watcherParams := WatcherParams{
		WatcherID: WatcherID(config.Name),
		Logger:    logger,
	}

	switch config.Name {
	case WatcherNameJournald:
		journal, err := sdjournal.NewJournal()
		if err != nil {
			return nil, errors.Trace(err)
		}

		params := JournaldParams{
			WatcherParams: watcherParams,
			Journal:       journal,
		}

		return NewJournald(params), nil

	case WatcherNameDocker:
		cl, err := client.NewClientWithOpts(
			client.FromEnv,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}

		params := DockerParams{
			WatcherParams: watcherParams,
			Client:        cl,
			Persister:     persister,
		}

		return NewDocker(params), nil
	default:
		return nil, errors.Errorf("unfamiliar watcher name: %q", config.Name)
	}
}

type WatcherName string

const (
	WatcherNameDocker   WatcherName = "docker"
	WatcherNameJournald WatcherName = "journald"
)
