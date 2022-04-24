package config

import (
	"io"
	"os"

	"github.com/juju/errors"
	"gopkg.in/yaml.v3"
)

// FromYAMLFile opens the file and tries to decode the config YAML.
func (c *Config) FromYAMLFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.Trace(err)
	}

	defer f.Close()

	err = c.FromYAMLReader(f)

	return errors.Trace(err)
}

// FromYAMLReader decodes the YAML config from reader.
func (c *Config) FromYAMLReader(reader io.Reader) error {
	err := yaml.NewDecoder(reader).Decode(c)

	return errors.Trace(err)
}

// FromYAMLEnv reads the YAML config from the environment variable.
func (c *Config) FromYAMLEnv(env string) error {
	err := c.FromYAMLString(os.Getenv(env))

	return errors.Trace(err)
}

// FromYAMLString reads the YAML config from a string value.
func (c *Config) FromYAMLString(str string) error {
	err := yaml.Unmarshal([]byte(str), c)

	return errors.Trace(err)
}
