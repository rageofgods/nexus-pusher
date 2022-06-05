package config

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
	// Set default client prometheus metrics endpoint url
	clientMetricsEndpointURI = "/metrics"
	// Set default client prometheus metrics endpoint port
	clientMetricsEndpointPort = "9090"
)

const (
	// URIBase Set base REST URI
	URIBase string = "/service/rest"
	// URILogin Set login REST URI
	URILogin string = "/login"
	// URIRefresh Set JWT refresh REST URI
	URIRefresh string = "/refresh"
	// URIStatus Set status REST URI
	URIStatus string = "/status"
	// URIVersion Set status REST URI
	URIVersion string = "/version"
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

const (
	npmSrv    string = "https://registry.npmjs.org/"
	pypiSrv   string = "https://pypi.org/"
	maven2Srv string = "https://repo1.maven.org/maven2/"
	nugetSrv  string = "https://api.nuget.org/v3/index.json"
)

// LogTimeFormat will format logrus time to specified format
const LogTimeFormat string = "02-01-2006 15:04:05 MST-07"
