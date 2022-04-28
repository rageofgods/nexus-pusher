package config

type NexusConfig struct {
	Server Server `yaml:"server"`
	Client Client `yaml:"client"`
}

func NewNexusConfig() *NexusConfig {
	return &NexusConfig{}
}

type Server struct {
	Enabled bool   `yaml:"enabled"`
	Port    string `yaml:"port"`
}

type Client struct {
	PushTo      string        `yaml:"pushTo"`
	SyncConfigs []*SyncConfig `yaml:"syncConfigs"`
}

type SyncConfig struct {
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
