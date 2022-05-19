package main

import (
	"fmt"
	"log"
	"net/http"
	"nexus-pusher/internal/client"
	"nexus-pusher/internal/comps"
	"nexus-pusher/internal/config"
	"nexus-pusher/internal/server"
	"os"
)

// App version
var (
	Version string
	Build   string
)

func main() {
	// Setup version
	version := comps.NewVersion(Version, Build)

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

	log.Printf("Starting application... Version: %s, Build: %s", Version, Build)
	// Run in Server mode
	if cfg.Server != nil {
		if cfg.Server.TLS.Enabled {
			log.Printf("Running in server mode (TLS). Listening on: %s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port)
			// Run Server with Let's encrypt autocert
			if cfg.Server.TLS.Auto {
				server.RunAutoCertServer(cfg.Server, version)
			} else { // Run Server with static cert config
				server.RunStaticCertServer(cfg.Server, version)
			}
		} else { // Run HTTP server (not secure!)
			log.Printf("Running in server mode (HTTP). Listening on: %s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port)
			log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Server.BindAddress, cfg.Server.Port),
				server.NewRouter(cfg.Server, version)))
		}

	} else if cfg.Client != nil { // Run in Client mode
		if cfg.Client.Daemon.Enabled {
			log.Printf("Running client in 'daemon' mode. Scheduling re-sync every %d minutes",
				cfg.Client.Daemon.SyncEveryMinutes)
			if err := client.ScheduleRunNexusPusher(cfg.Client, version); err != nil {
				log.Printf("%v", err)
				os.Exit(1)
			}
		} else {
			log.Println("Running client in 'ad hoc' mode. Will do sync only once.")
			client.RunNexusPusher(cfg.Client, version)
		}
	}
}
