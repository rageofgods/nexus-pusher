package config

import (
	"fmt"
	"nexus-pusher/pkg/helper"
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
		switch {
		case c.Server.Port == "":
			c.Server.Port = serverPort
			fallthrough
		case c.Server.BindAddress == "":
			c.Server.BindAddress = serverBindAddress
			fallthrough
		case len(c.Server.Credentials) == 0:
			return &helper.ContextError{
				Context: "validateServerConfig",
				Err:     fmt.Errorf("server required 'credentials' variable is missing in %s", c.string),
			}
		case c.Server.Concurrency == 0:
			c.Server.Concurrency = clientConcurrency
			fallthrough
		case c.Server.TLS.Enabled && c.Server.TLS.Auto:
			if c.Server.TLS.DomainName == "" {
				return &helper.ContextError{
					Context: "validateServerConfig",
					Err:     fmt.Errorf("server required 'domainName' variable is missing in %s", c.string),
				}
			}
			fallthrough
		case c.Server.TLS.Enabled && !c.Server.TLS.Auto:
			if c.Server.TLS.KeyPath == "" || c.Server.TLS.CertPath == "" {
				return &helper.ContextError{
					Context: "validateServerConfig",
					Err:     fmt.Errorf("you must set 'KeyPath' and 'CertPath' variables in %s", c.string),
				}
			}
		}
	}
	return nil
}

func (c *NexusConfig) validateClientConfig() error {
	// Check client required parameters
	if c.Client != nil {
		switch {
		case c.Client.ServerAuth.User == "":
			return &helper.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'serverAuth.user' variable is missing in %s", c.string),
			}
		case c.Client.ServerAuth.Pass == "":
			return &helper.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'serverAuth.pass' variable is missing in %s", c.string),
			}
		case c.Client.Server == "":
			return &helper.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'server' variable is missing in %s", c.string),
			}
		case c.Client.Daemon.SyncEveryMinutes == 0:
			c.Client.Daemon.SyncEveryMinutes = clientDaemonSyncEveryMinutes
			fallthrough
		case c.Client.SyncConfigs == nil:
			return &helper.ContextError{
				Context: "validateClientConfig",
				Err:     fmt.Errorf("client required 'syncConfigs' variable is missing in %s", c.string),
			}
		case c.Client.SyncConfigs != nil:
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
		return &helper.ContextError{
			Context: "validateArtifactsSource",
			Err:     fmt.Errorf("syncconfig required 'format' variable is missing in %s", syncConfig),
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
	}
	return nil
}

func (c *NexusConfig) validateTargetServerConfigs(syncConfig *SyncConfig, index int) error {
	// Check source server parameters
	switch src := syncConfig.SrcServerConfig; {
	case src.Server == "":
		if c.Client.SyncGlobalAuth.SrcServer == "" {
			return &helper.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.srcServer' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].SrcServerConfig.Server = c.Client.SyncGlobalAuth.SrcServer
		fallthrough
	case src.User == "":
		if c.Client.SyncGlobalAuth.SrcServerUser == "" {
			return &helper.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.srcServerUser' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].SrcServerConfig.User = c.Client.SyncGlobalAuth.SrcServerUser
		fallthrough
	case src.Pass == "":
		if c.Client.SyncGlobalAuth.SrcServerPass == "" {
			return &helper.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.srcServerPass' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].SrcServerConfig.Pass = c.Client.SyncGlobalAuth.SrcServerPass
	}
	// Check destination server parameters
	switch dst := syncConfig.DstServerConfig; {
	case dst.Server == "":
		if c.Client.SyncGlobalAuth.DstServer == "" {
			return &helper.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.dstServer' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].DstServerConfig.Server = c.Client.SyncGlobalAuth.DstServer
		fallthrough
	case dst.User == "":
		if c.Client.SyncGlobalAuth.DstServerUser == "" {
			return &helper.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.dstServerUser' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].DstServerConfig.User = c.Client.SyncGlobalAuth.DstServerUser
		fallthrough
	case dst.Pass == "":
		if c.Client.SyncGlobalAuth.DstServerPass == "" {
			return &helper.ContextError{
				Context: "validateTargetServerConfigs",
				Err:     fmt.Errorf("no 'client.syncGlobalAuth.dstServerPass' or syncConfig specific defined"),
			}
		}
		c.Client.SyncConfigs[index].DstServerConfig.Pass = c.Client.SyncGlobalAuth.DstServerPass
	}
	return nil
}
