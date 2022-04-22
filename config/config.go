package config

import "github.com/jeremija/taily/types"

// Config describes the main YAML config file.
type Config struct {
	Readers    []Reader             `yaml:"readers"`
	Actions    map[string]Action    `yaml:"actions"`
	Processors map[string]Processor `yaml:"processors"`
	Persister  Persister            `yaml:"persister"`
}

// Reader contains configuration for a specific watcher.
type Reader struct {
	ID           types.ReaderID `yaml:"id"`
	Type         string         `yaml:"type"`
	Processors   []string       `yaml:"processors"`
	InitialState types.State    `yaml:"initial_state"`
}

func (r Reader) ReaderID() types.ReaderID {
	if r.ID != "" {
		return r.ID
	}

	return types.ReaderID(r.Type)
}

// Processor contains configuration for a specific processor.
type Processor struct {
	Type    string           `yaml:"type"`
	Action  string           `yaml:"action"`
	Matcher ProcessorMatcher `yaml:"matcher"`
}

// ProcessorMatcher contains cofiguration for ProcessorMatcher.
type ProcessorMatcher struct {
	Start      *Matcher `yaml:"start_line"`
	End        *Matcher `yaml:"end_line"`
	IncludeEnd bool     `yaml:"include_end"`
	MaxLines   int      `yaml:"max_lines"`
	GroupBy    []string `yaml:"group_by"`
}

// Matcher contains configuration for Matcher.
type Matcher struct {
	Type      string     `yaml:"type"`
	Expr      string     `yaml:"expr"`
	String    string     `yaml:"string"`
	Substring string     `yaml:"substring"`
	Prefix    string     `yaml:"prefix"`
	Suffix    string     `yaml:"suffix"`
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
	Type   string       `yaml:"type"`
	Log    ActionLog    `yaml:"log"`
	Notify ActionNotify `yaml:"notify"`
}

type Format struct {
	Type     string   `yaml:"type"`
	Template Template `yaml:"template"`
}

type Template struct {
	Format     string `yaml:"format"`
	OpenTag    rune   `yaml:"open_tag"`
	CloseTag   rune   `yaml:"close_tag"`
	OpenQuote  rune   `yaml:"open_quote"`
	CloseQuote rune   `yaml:"close_quote"`
}

type ActionLog struct {
	Format Format `yaml:"format"`
}

type ActionNotify struct {
	TitleFormat  Format          `yaml:"title_format"`
	BodyFormat   Format          `yaml:"body_format"`
	Services     []NotifyService `yaml:"services"`
	MaxTitleSize int             `yaml:"max_title_size"`
	MaxBodySize  int             `yaml:"max_body_size"`
}

type NotifyService struct {
	Type     string `yaml:"type"`
	Telegram struct {
		Token     string  `yaml:"token"`
		Receivers []int64 `yaml:"receivers"`
	} `yaml:"telegram"`
	Slack struct {
		Token     string   `yaml:"token"`
		Receivers []string `yaml:"receivers"`
	} `yaml:"slack"`
}
