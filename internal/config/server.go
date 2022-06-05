package config

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
