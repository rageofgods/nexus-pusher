package config

// Client is defines client-side config part
type Client struct {
	Daemon struct {
		Enabled          bool `yaml:"enabled"`
		SyncEveryMinutes int  `yaml:"syncEveryMinutes"`
	} `yaml:"daemon"`
	Metrics struct {
		Enabled      bool   `yaml:"enabled"`
		EndpointURI  string `yaml:"endpointUri"`
		EndpointPort string `yaml:"endpointPort"`
	} `yaml:"metrics"`
	Server         string         `yaml:"server"`
	ServerAuth     ServerAuth     `yaml:"serverAuth"`
	SyncGlobalAuth SyncGlobalAuth `yaml:"syncGlobalAuth"`
	SyncConfigs    []*SyncConfig  `yaml:"syncConfigs"`
}

type SyncGlobalAuth struct {
	SrcServer     string `yaml:"srcServer"`
	SrcServerUser string `yaml:"srcServerUser"`
	SrcServerPass string `yaml:"srcServerPass"`
	DstServer     string `yaml:"dstServer"`
	DstServerUser string `yaml:"dstServerUser"`
	DstServerPass string `yaml:"dstServerPass"`
}

// ServerAuth is defines client side server auth
type ServerAuth struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

// SyncConfig is defines set of sync-configs for client
type SyncConfig struct {
	Format          string          `yaml:"format"`
	ArtifactsSource string          `yaml:"artifactsSource"`
	SrcServerConfig SrcServerConfig `yaml:"srcServerConfig"`
	DstServerConfig DstServerConfig `yaml:"dstServerConfig"`
	IsProcessing    bool
}

func (sc *SyncConfig) Lock() {
	sc.IsProcessing = true
}

func (sc *SyncConfig) UnLock() {
	sc.IsProcessing = false
}

func (sc *SyncConfig) IsLock() bool {
	return sc.IsProcessing
}

// SrcServerConfig is defines source server which will be compared to destination
type SrcServerConfig struct {
	Server   string `yaml:"server"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	RepoName string `yaml:"repoName"`
}

// DstServerConfig is defines destination server config (target)
type DstServerConfig struct {
	Server   string `yaml:"server"`
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	RepoName string `yaml:"repoName"`
}
