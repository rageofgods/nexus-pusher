package config

import (
	"fmt"
)

// ValidateConfig is used to validate config file for correct parameters
func (c *NexusConfig) ValidateConfig() error {
	// Check server required parameters and setup defaults if they are missing
	if c.Server.Enabled {
		if c.Server.Port == "" {
			c.Server.Port = serverPort
		} else if c.Server.BindAddress == "" {
			c.Server.BindAddress = serverBindAddress
		} else if len(c.Server.Credentials) == 0 {
			return fmt.Errorf("error: server required 'credentials' variable is missing in %s", c.string)
		} else if c.Server.Concurrency == 0 {
			c.Server.Concurrency = clientConcurrency
		}
		return nil
	}
	// Check client required parameters
	if c.Client.SyncConfigs == nil {
		return fmt.Errorf("error: client required 'syncConfigs' variable is missing in %s", c.string)
	} else if c.Client.ServerAuth.User == "" {
		return fmt.Errorf("error: client required 'serverAuth.user' variable is missing in %s", c.string)
	} else if c.Client.ServerAuth.Pass == "" {
		return fmt.Errorf("error: client required 'serverAuth.pass' variable is missing in %s", c.string)
	} else if c.Client.Server == "" {
		return fmt.Errorf("error: client required 'server' variable is missing in %s", c.string)
	} else if c.Client.Daemon.SyncEveryMinutes == 0 {
		c.Client.Daemon.SyncEveryMinutes = clientDaemonSyncEveryMinutes
	}
	// Check SynConfigs parameters
	for _, v := range c.Client.SyncConfigs {
		if v.Format == "" {
			return fmt.Errorf("error: syncconfig required 'format' variable is missing in %s", v)
		}
	}

	return nil
}
