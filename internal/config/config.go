package config

// NexusConfig is a root of configuration
type NexusConfig struct {
	string
	Server *Server `yaml:"server"`
	Client *Client `yaml:"client"`
}

// NewNexusConfig returns empty NexusConfig
func NewNexusConfig() *NexusConfig {
	return &NexusConfig{}
}
