package config

type NexusConfig struct {
	Server Server `yaml:"server"`
	Client Client `yaml:"client"`
}

func NewNexusConfig() *NexusConfig {
	return &NexusConfig{}
}

type Server struct {
	Enabled     bool   `yaml:"enabled"`
	BindAddress string `yaml:"bindAddress"`
	Port        string `yaml:"port"`
	Concurrency int    `yaml:"concurrency"`
}

type Client struct {
	SyncConfigs []*SyncConfig `yaml:"syncConfigs"`
}

type SyncConfig struct {
	PushTo          string          `yaml:"pushTo,omitempty"`
	SrcServerConfig SrcServerConfig `yaml:"srcServerConfig"`
	DstServerConfig DstServerConfig `yaml:"dstServerConfig"`
}

type SrcServerConfig struct {
	Server   string `yaml:"server"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	RepoName string `yaml:"repoName"`
}

type DstServerConfig struct {
	Server   string `yaml:"server"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	RepoName string `yaml:"repoName"`
}

const (
	configName string = "config.yaml"
)
