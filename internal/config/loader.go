package config

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
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

// ScheduleLoadConfig wrapper around LoadConfig to schedule config reads
func (c *NexusConfig) ScheduleLoadConfig(fileName string, seconds int) error {
	loc, err := time.LoadLocation(TimeZone)
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	j, err := s.Every(seconds).Second().Do(c.LoadConfig, fileName)
	if err != nil {
		return fmt.Errorf("error: can't schedule config read. job: %v: error: %w", j, err)
	}
	s.StartAsync()

	return nil
}
