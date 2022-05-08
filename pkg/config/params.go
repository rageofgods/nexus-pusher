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

// Server is defines server-side config part
type Server struct {
	BindAddress string            `yaml:"bindAddress"`
	Port        string            `yaml:"port"`
	Concurrency int               `yaml:"concurrency"`
	Credentials map[string]string `json:"credentials"`
	TLS         struct {
		Enabled    bool   `yaml:"enabled"`
		Auto       bool   `yaml:"auto"`
		DomainName string `yaml:"domainName"`
		KeyPath    string `yaml:"keyPath"`
		CertPath   string `yaml:"certPath"`
	} `yaml:"tls"`
}

// Client is defines client-side config part
type Client struct {
	Daemon struct {
		Enabled          bool `yaml:"enabled"`
		SyncEveryMinutes int  `yaml:"syncEveryMinutes"`
	} `yaml:"daemon"`
	Server      string        `yaml:"server"`
	ServerAuth  ServerAuth    `yaml:"serverAuth"`
	SyncConfigs []*SyncConfig `yaml:"syncConfigs"`
}

// ServerAuth is defines client side server auth
type ServerAuth struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

// SyncConfig is defines set of sync-configs for client
type SyncConfig struct {
	Format          string          `yaml:"format"`
	SrcServerConfig SrcServerConfig `yaml:"srcServerConfig"`
	DstServerConfig DstServerConfig `yaml:"dstServerConfig"`
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

const (
	// Set default client concurrency
	clientConcurrency int = 30
	// Set default server port
	serverPort string = "8181"
	// Set default server bind address
	serverBindAddress string = "0.0.0.0"
	// Set default config file name
	configName string = "config.yaml"
	// TimeZone Set default timezone
	TimeZone string = "Europe/Moscow"
	// Set default client sync time in daemon mode
	clientDaemonSyncEveryMinutes = 30
)

const (
	// URIBase Set base REST URI
	URIBase string = "/service/rest"
	// URILogin Set login REST URI
	URILogin string = "/login"
	// URIRefresh Set JWT refresh REST URI
	URIRefresh string = "/refresh"
	// URIComponents Set components REST URI
	URIComponents string = "/v1/components"
	// URIRepositories Set repositories REST URI
	URIRepositories string = "/v1/repositories"
)

const (
	// JWTTokenTTL Set JWT token TTL in minutes
	JWTTokenTTL = 5
	// JWTCookieName Set JWT token Cookie name
	JWTCookieName = "token"
	// JWTTokenRefreshWindow Set JWT token refresh window in seconds
	JWTTokenRefreshWindow = 30
)
