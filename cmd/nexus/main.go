package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nexus-pusher/internal/client"
	"nexus-pusher/internal/config"
	"nexus-pusher/internal/core"
	"nexus-pusher/internal/server"
	"nexus-pusher/pkg/logger"
	"nexus-pusher/pkg/metrics"
	"os"
)

// App version
var (
	Version string
	Build   string
)

func main() {
	// Setup logger
	logger.NewLogger().SetupLogger()

	// Setup version
	version := core.NewVersion(Version, Build)

	// Get Config Args
	args := &config.Args{}
	if args = args.GetConfigArgs(); args == nil {
		log.Fatalf("args is nil")
	}

	// Load Nexus-Pusher configuration from file
	cfg := config.NewNexusConfig()
	if err := cfg.LoadConfig(args.ConfigPath); err != nil {
		log.Fatalf("unable to load config: %v", err)
	}

	// Schedule periodic config file re-read
	// if err := cfg.ScheduleLoadConfig(args.ConfigPath, 30); err != nil {
	//	log.Printf("error: %v", err)
	// }

	log.WithFields(log.Fields{"version": Version, "build": Build}).Info("Starting application...")

	// Run in Server mode
	if cfg.Server != nil {
		if cfg.Server.TLS.Enabled {
			log.WithFields(log.Fields{
				"proto":        "TLS",
				"bind_address": cfg.Server.BindAddress,
				"port":         cfg.Server.Port},
			).Info("Running in server mode.")

			// Run Server with Let's encrypt autocert
			if cfg.Server.TLS.Auto {
				server.RunAutoCertServer(cfg.Server, version)
			} else { // Run Server with static cert config
				server.RunStaticCertServer(cfg.Server, version)
			}
		} else { // Run HTTP server (not secure!)
			log.WithFields(log.Fields{
				"proto":        "HTTP",
				"bind_address": cfg.Server.BindAddress,
				"port":         cfg.Server.Port,
			}).Info("Running in server mode.")

			log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", cfg.Server.BindAddress, cfg.Server.Port),
				server.NewRouter(cfg.Server, version)))
		}

	} else if cfg.Client != nil { // Run in Client mode
		// Create new prometheus registry
		r := metrics.NewRegister(cfg.Client.Metrics.EndpointURI, cfg.Client.Metrics.EndpointPort)

		// Start serving metrics endpoint
		if cfg.Client.Metrics.Enabled {
			r.StartServing()
		}

		// Create client metrics with prometheus exporter
		clientMetrics := client.NewMetrics(r.Registry())

		// Export client version and build info
		clientMetrics.ClientInfo().WithLabelValues(version.Version, version.Build).Set(1)

		// Create new nexus-pusher client
		c := client.NewClient(version, cfg.Client, clientMetrics)

		if cfg.Client.Daemon.Enabled {
			syncMinutes := cfg.Client.Daemon.SyncEveryMinutes
			log.WithFields(log.Fields{
				"sync_minutes": syncMinutes,
			}).Info("Running client in 'daemon' mode.")

			// Run client in daemon mode (schedule)
			if err := c.ScheduleRunNexusPusher(syncMinutes); err != nil {
				log.Printf("%v", err)
				os.Exit(1)
			}
		} else {
			log.WithFields(log.Fields{
				"sync(minutes)": cfg.Client.Daemon.SyncEveryMinutes,
			}).Info("Running client in 'ad hoc' mode. Will do sync only once.")

			// Run client in ad-hoc mode
			c.RunNexusPusher()
		}
	}
}
