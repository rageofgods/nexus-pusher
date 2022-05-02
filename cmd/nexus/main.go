package main

import (
	"fmt"
	"log"
	"net/http"
	"nexus-pusher/pkg/client"
	"nexus-pusher/pkg/config"
	"nexus-pusher/pkg/server"
)

func main() {
	// Get Config Path
	var configPath string
	if configPath = config.GetConfigPath(); configPath == "" {
		return
	}

	// Load Nexus-Pusher configuration from file
	cfg := config.NewNexusConfig()
	if err := cfg.LoadConfig(configPath); err != nil {
		log.Fatalf("unable to load config: %v", err)
	}
	// Validate config for correct syntax
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("%v", err)
	}
	// Start Server or Client version following configuration
	if cfg.Server.Enabled {
		log.Printf("Running in server mode. Listening on: %s:%s",
			cfg.Server.BindAddress,
			cfg.Server.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s",
			cfg.Server.BindAddress,
			cfg.Server.Port), server.NewRouter(&cfg.Server)))
	} else {
		log.Println("Running in client mode.")
		client.RunNexusPusher(cfg)
	}
}
