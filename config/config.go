package config

import "github.com/jeremija/taily/types"

// Config describes the main YAML config file.
type Config struct {
	Watchers   []Watcher            `yaml:"watchers"`
	Actions    map[string]Action    `yaml:"actions"`
	Processors map[string]Processor `yaml:"processors"`
	Persister  Persister            `yaml:"persister"`
}

// Watcher contains configuration for a specific watcher.
type Watcher struct {
	Type         string      `yaml:"type"`
	Processors   []string    `yaml:"processors"`
	InitialState types.State `yaml:"initial_state"`
}

// Processor contains configuration for a specific processor.
type Processor struct {
	Type    string           `yaml:"type"`
	Action  string           `yaml:"action"`
	Matcher ProcessorMatcher `yaml:"matcher"`
}

// ProcessorMatcher contains cofiguration for ProcessorMatcher.
type ProcessorMatcher struct {
	Start *Matcher `yaml:"start_line"`
	End   *Matcher `yaml:"end_line"`
}

// Matcher contains configuration for Matcher.
type Matcher struct {
	Type      string     `yaml:"type"`
	Substring string     `yaml:"substring"`
	Regexp    string     `yaml:"regexp"`
	And       []*Matcher `yaml:"and"`
	Or        []*Matcher `yaml:"or"`
	Not       *Matcher   `yaml:"not"`
	Field     struct {
		Name   string `yaml:"name"`
		Regexp string `yaml:"pattern"`
	} `yaml:"field"`
}

// ProcessorLogConfig contains configuration for ProcessorLog.
type ProcessorAny struct {
	Action string `json:"action"`
}

// Persister contains configuration for Persister.
type Persister struct {
	Type string        `yaml:"type"`
	File PersisterFile `yaml:"file"`
}

// PersisterFile contains configuration for PersisterFile.
type PersisterFile struct {
	Dir string `yaml:"dir"`
}

type Action struct {
	Type string
	Log  ActionLog
}

type ActionLog struct {
	Format string
}
