package main

import (
	"fmt"
	"log"
	"net/http"
	"nexus-pusher/pkg/client"
	"nexus-pusher/pkg/config"
	"nexus-pusher/pkg/server"
	"os"
)

func main() {
	// Get Config Args
	args := &config.Args{}
	if args = args.GetConfigArgs(); args == nil {
		return
	}

	// Load Nexus-Pusher configuration from file
	cfg := config.NewNexusConfig()
	if err := cfg.LoadConfig(args.ConfigPath); err != nil {
		log.Fatalf("unable to load config: %v", err)
	}

	// Validate config for correct syntax
	if err := cfg.ValidateConfig(); err != nil {
		log.Fatalf("%v", err)
	}

	// Schedule periodic config file re-read
	if err := cfg.ScheduleLoadConfig(args.ConfigPath, 30); err != nil {
		log.Printf("error: %v", err)
	}

	// Start Server or Client version following provided configuration
	if cfg.Server.Enabled {
		if cfg.Server.TLS.Enabled {
			log.Printf("Running in server mode (TLS). Listening on: %s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port)
			log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port),
				cfg.Server.TLS.CertPath,
				cfg.Server.TLS.KeyPath,
				server.NewRouter(cfg.Server)))
		} else {
			log.Printf("Running in server mode (HTTP). Listening on: %s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port)
			log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port), server.NewRouter(cfg.Server)))
		}
	} else {
		if cfg.Client.Daemon.Enabled {
			log.Printf("Running client in 'daemon' mode. Scheduling re-sync every %d minutes",
				cfg.Client.Daemon.SyncEveryMinutes)
			if err := client.ScheduleRunNexusPusher(cfg.Client); err != nil {
				log.Printf("%v", err)
				os.Exit(1)
			}
		} else {
			log.Println("Running client in 'ad hoc' mode. Will do sync only once.")
			client.RunNexusPusher(cfg.Client)
		}
	}
}
