package config

type NexusConfig struct {
	string
	Server Server `yaml:"server"`
	Client Client `yaml:"client"`
}

func NewNexusConfig() *NexusConfig {
	return &NexusConfig{}
}

type Server struct {
	Enabled     bool              `yaml:"enabled"`
	BindAddress string            `yaml:"bindAddress"`
	Port        string            `yaml:"port"`
	Concurrency int               `yaml:"concurrency"`
	Credentials map[string]string `json:"credentials"`
}

type Client struct {
	Server      string        `yaml:"server"`
	ServerAuth  ServerAuth    `yaml:"serverAuth"`
	SyncConfigs []*SyncConfig `yaml:"syncConfigs"`
}

type ServerAuth struct {
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
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

const (
	// Set default client concurrency
	clientConcurrency int = 30
	// Set default server port
	serverPort string = "8181"
	// Set default server bind address
	serverBindAddress string = "0.0.0.0"
	// Set default config file name
	configName string = "config.yaml"
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
)

const (
	// JWTTokenTTL Set JWT token TTL in minutes
	JWTTokenTTL = 5
	// JWTCookieName Set JWT token Cookie name
	JWTCookieName = "token"
	// JWTTokenRefreshWindow Set JWT token refresh window in seconds
	JWTTokenRefreshWindow = 30
)
