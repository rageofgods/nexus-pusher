package config

import (
	"fmt"
	"nexus-pusher/pkg/utils"
)

// ValidateConfig is used to validate config file for correct parameters
func (c *NexusConfig) validateConfig() error {
	// Validate server config
	if err := c.validateServerConfig(); err != nil {
		return fmt.Errorf("ValidateConfig: %w", err)
	}

	// Validate client config
	if err := c.validateClientConfig(); err != nil {
		return fmt.Errorf("ValidateConfig: %w", err)
	}

	return nil
}

func (c *NexusConfig) validateServerConfig() error {
	// Check server required parameters and setup defaults if they are missing
	if c.Server != nil {
		if c.Server.Port == "" {
			c.Server.Port = serverPort
		}

		if c.Server.BindAddress == "" {
			c.Server.BindAddress = serverBindAddress
		}

		if len(c.Server.Credentials) == 0 {
			return &utils.ContextError{
				Context: "validateServerConfig",
				Err:     fmt.Errorf("server required 'credentials' variable is missing in %s", c.string),
			}
		}

		if c.Server.Concurrency == 0 {
			c.Server.Concurrency = clientConcurrency
		}

		if c.Server.TLS.Enabled && c.Server.TLS.Auto {
			if c.Server.TLS.DomainName == "" {
				return &utils.ContextError{
					Context: "validateServerConfig",
					Err:     fmt.Errorf("server required 'domainName' variable is missing in %s", c.string),
				}
			}
		}

		if c.Server.TLS.Enabled && !c.Server.TLS.Auto {
			if c.Server.TLS.KeyPath == "" || c.Server.TLS.CertPath == "" {
				return &utils.ContextError{
					Context: "validateServerConfig",
					Err:     fmt.Errorf("you must set 'KeyPath' and 'CertPath' variables in %s", c.string),
				}
			}
		}
	}
	return nil
}

func (c *NexusConfig) validateClientConfig() error {
	if c.Client != nil {
		// Check client required parameters
		if c.Client.ServerAuth.User == "" {
			return &utils.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'serverAuth.user' variable is missing in %s", c.string),
			}
		}

		if c.Client.ServerAuth.Pass == "" {
			return &utils.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'serverAuth.pass' variable is missing in %s", c.string),
			}
		}

		if c.Client.Server == "" {
			return &utils.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'server' variable is missing in %s", c.string),
			}
		}

		if c.Client.Daemon.SyncEveryMinutes == 0 {
			c.Client.Daemon.SyncEveryMinutes = clientDaemonSyncEveryMinutes
		}

		if c.Client.Metrics.EndpointURI == "" {
			c.Client.Metrics.EndpointURI = clientMetricsEndpointURI
		}

		if c.Client.Metrics.EndpointPort == "" {
			c.Client.Metrics.EndpointPort = clientMetricsEndpointPort
		}

		if c.Client.SyncConfigs == nil {
			return &utils.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'syncConfigs' variable is missing in %s", c.string),
			}
		}

		if c.Client.SyncConfigs != nil {
			for i, v := range c.Client.SyncConfigs {
				// Set global artifacts source if where is no specific one
				if err := c.validateArtifactsSource(v, i); err != nil {
					return fmt.Errorf("validateClientConfig: %w", err)
				}
				// Set global parameters if where is no specific one
				if err := c.validateTargetServerConfigs(v, i); err != nil {
					return fmt.Errorf("validateClientConfig: %w", err)
				}
			}
		}
	}
	return nil
}

func (c *NexusConfig) validateArtifactsSource(syncConfig *SyncConfig, index int) error {
	switch syncConfig.Format {
	case "":
		return &utils.ContextError{
			Context: "validateArtifactsSource",
			Err:     fmt.Errorf("syncconfig required 'format' variable is missing in %v", syncConfig),
		}
	case MAVEN2.String():
		if syncConfig.ArtifactsSource == "" {
			c.Client.SyncConfigs[index].ArtifactsSource = maven2Srv
		}
	case PYPI.String():
		if syncConfig.ArtifactsSource == "" {
			c.Client.SyncConfigs[index].ArtifactsSource = pypiSrv
		}
	case NPM.String():
		if syncConfig.ArtifactsSource == "" {
			c.Client.SyncConfigs[index].ArtifactsSource = npmSrv
		}
	case NUGET.String():
		if syncConfig.ArtifactsSource == "" {
			c.Client.SyncConfigs[index].ArtifactsSource = nugetSrv
		}
	}
	return nil
}

func (c *NexusConfig) validateTargetServerConfigs(syncConfig *SyncConfig, index int) error {
	// Check source server parameters
	if syncConfig.SrcServerConfig.Server == "" {
		if c.Client.SyncGlobalAuth.SrcServer == "" {
			return &utils.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.srcServer' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].SrcServerConfig.Server = c.Client.SyncGlobalAuth.SrcServer
	}

	if syncConfig.SrcServerConfig.User == "" {
		if c.Client.SyncGlobalAuth.SrcServerUser == "" {
			return &utils.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.srcServerUser' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].SrcServerConfig.User = c.Client.SyncGlobalAuth.SrcServerUser
	}

	if syncConfig.SrcServerConfig.Pass == "" {
		if c.Client.SyncGlobalAuth.SrcServerPass == "" {
			return &utils.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.srcServerPass' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].SrcServerConfig.Pass = c.Client.SyncGlobalAuth.SrcServerPass
	}

	// Check destination server parameters
	if syncConfig.DstServerConfig.Server == "" {
		if c.Client.SyncGlobalAuth.DstServer == "" {
			return &utils.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.dstServer' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].DstServerConfig.Server = c.Client.SyncGlobalAuth.DstServer
	}

	if syncConfig.DstServerConfig.User == "" {
		if c.Client.SyncGlobalAuth.DstServerUser == "" {
			return &utils.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.dstServerUser' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].DstServerConfig.User = c.Client.SyncGlobalAuth.DstServerUser
	}

	if syncConfig.DstServerConfig.Pass == "" {
		if c.Client.SyncGlobalAuth.DstServerPass == "" {
			return &utils.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.dstServerPass' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].DstServerConfig.Pass = c.Client.SyncGlobalAuth.DstServerPass
	}

	return nil
}
