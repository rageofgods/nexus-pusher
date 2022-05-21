package config

import (
	"fmt"
)

// ValidateConfig is used to validate config file for correct parameters
func (c *NexusConfig) ValidateConfig() error {
	// Check server required parameters and setup defaults if they are missing
	if c.Server != nil {
		switch {
		case c.Server.Port == "":
			c.Server.Port = serverPort
		case c.Server.BindAddress == "":
			c.Server.BindAddress = serverBindAddress
		case len(c.Server.Credentials) == 0:
			return fmt.Errorf("error: server required 'credentials' variable is missing in %s", c.string)
		case c.Server.Concurrency == 0:
			c.Server.Concurrency = clientConcurrency
		case c.Server.TLS.Enabled && c.Server.TLS.Auto:
			if c.Server.TLS.DomainName == "" {
				return fmt.Errorf("error: server required 'domainName' variable is missing in %s", c.string)
			}
		case c.Server.TLS.Enabled && !c.Server.TLS.Auto:
			if c.Server.TLS.KeyPath == "" || c.Server.TLS.CertPath == "" {
				return fmt.Errorf("error: you must set 'KeyPath' and 'CertPath' variables in %s", c.string)
			}
		default:
			return nil
		}
	}

	// Check client required parameters
	if c.Client != nil {
		switch {
		case c.Client.ServerAuth.User == "":
			return fmt.Errorf("error: client required 'serverAuth.user' variable is missing in %s", c.string)
		case c.Client.ServerAuth.Pass == "":
			return fmt.Errorf("error: client required 'serverAuth.pass' variable is missing in %s", c.string)
		case c.Client.Server == "":
			return fmt.Errorf("error: client required 'server' variable is missing in %s", c.string)
		case c.Client.Daemon.SyncEveryMinutes == 0:
			c.Client.Daemon.SyncEveryMinutes = clientDaemonSyncEveryMinutes
		case c.Client.SyncConfigs == nil:
			return fmt.Errorf("error: client required 'syncConfigs' variable is missing in %s", c.string)
		case c.Client.SyncConfigs != nil:
			for i, v := range c.Client.SyncConfigs {
				switch v.Format {
				case "":
					return fmt.Errorf("error: syncconfig required 'format' variable is missing in %s", v)
				case MAVEN2.String():
					if v.ArtifactsSource == "" {
						c.Client.SyncConfigs[i].ArtifactsSource = maven2Srv
					}
				case PYPI.String():
					if v.ArtifactsSource == "" {
						c.Client.SyncConfigs[i].ArtifactsSource = pypiSrv
					}
				case NPM.String():
					if v.ArtifactsSource == "" {
						c.Client.SyncConfigs[i].ArtifactsSource = npmSrv
					}
				}
			}
		default:
			return nil
		}
	}

	return nil
}
