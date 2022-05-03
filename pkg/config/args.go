package config

import (
	"github.com/spf13/pflag"
)

type Args struct {
	ConfigPath string
	DaemonMode bool
}

// GetConfigArgs returns config specific args
func (a *Args) GetConfigArgs() *Args {
	var showHelp bool

	pflag.StringVarP(&a.ConfigPath, "config", "c", configName,
		"Config file path")
	pflag.BoolVarP(&a.DaemonMode, "daemon", "d", false,
		"Enable daemon mode")
	pflag.BoolVarP(&showHelp, "help", "h", false,
		"Show help message")

	pflag.Parse()
	if showHelp {
		pflag.Usage()
		return nil
	}

	return a
}
