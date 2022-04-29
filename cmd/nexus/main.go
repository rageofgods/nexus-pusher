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
	// Load Nexus-Pusher configuration from file
	cfg := config.NewNexusConfig()
	if err := cfg.LoadConfig(configName); err != nil {
		log.Fatalf("unable to load config: %v", err)
	}
	// Start Server or Client version following configuration
	if cfg.Server.Enabled {
		log.Printf("Running in server mode. Listening on: %s:%s",
			cfg.Server.BindAddress,
			cfg.Server.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s",
			cfg.Server.BindAddress,
			cfg.Server.Port), server.NewRouter()))
	} else {
		log.Println("Running in client mode.")
		client.RunNexusPusher(cfg)
	}
}

const (
	configName string = "config.yaml"
)
