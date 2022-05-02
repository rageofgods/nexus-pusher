package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func (c *NexusConfig) LoadConfig(fileName string) error {
	config, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(config, c); err != nil {
		return err
	}
	// Set config path
	c.string = fileName

	return nil
}
