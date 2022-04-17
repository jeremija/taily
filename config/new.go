package config

import (
	"io"
	"os"

	"github.com/juju/errors"
	"gopkg.in/yaml.v3"
)

// NewFromFile opens the filename and tries to decode the config YAML.
func NewFromFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer f.Close()

	config, err := NewFromReader(f)

	return config, errors.Trace(err)
}

// NewFromReader decodes the YAML config from reader.
func NewFromReader(reader io.Reader) (*Config, error) {
	var config Config

	err := yaml.NewDecoder(reader).Decode(&config)

	return &config, errors.Trace(err)
}

// NewFromEnv reads the YAML config from the environment variable.
func NewFromEnv(env string) (*Config, error) {
	config, err := NewFromString(os.Getenv(env))

	return config, errors.Trace(err)
}

// NewFromString reads the YAML config from a string value.
func NewFromString(str string) (*Config, error) {
	var config Config

	err := yaml.Unmarshal([]byte(str), &config)

	return &config, errors.Trace(err)
}
