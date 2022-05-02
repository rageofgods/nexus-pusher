package config

import "github.com/spf13/pflag"

// GetConfigPath parse cmd args
func GetConfigPath() string {
	var showHelp bool
	var configPath string

	pflag.StringVarP(&configPath, "config", "c", "",
		"Config file path")
	pflag.BoolVarP(&showHelp, "help", "h", false,
		"Show help message")
	pflag.Parse()
	if showHelp {
		pflag.Usage()
		return ""
	}
	if configPath == "" {
		configPath = configName
	}
	return configPath
}
