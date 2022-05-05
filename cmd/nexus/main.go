package main

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"nexus-pusher/pkg/client"
	"nexus-pusher/pkg/config"
	"nexus-pusher/pkg/server"
	"os"
	"strings"
	"time"
)

// App version
var (
	Version string
	Build   string
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

	log.Printf("Starting application... Version: %s, Build: %s", Version, Build)
	// Start Server or Client version following provided configuration
	if cfg.Server.Enabled {
		if cfg.Server.TLS.Enabled {
			log.Printf("Running in server mode (TLS). Listening on: %s:%s",
				cfg.Server.BindAddress,
				cfg.Server.Port)
			//log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%s",
			//	cfg.Server.BindAddress,
			//	cfg.Server.Port),
			//	cfg.Server.TLS.CertPath,
			//	cfg.Server.TLS.KeyPath,
			//	server.NewRouter(cfg.Server)))
			runLetsEncrypt(cfg.Server)
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

func runLetsEncrypt(cfg *config.Server) {
	c := autocert.DirCache("certs")
	m := autocert.Manager{
		Cache:      c,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("rageofgods.xyz"),
	}

	s := &http.Server{
		Addr:      ":8443",
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   server.NewRouter(cfg),
	}

	go func() {
		log.Fatal(s.ListenAndServeTLS("", ""))
	}()

	httpSrv := makeHTTPToHTTPSRedirectServer()
	httpSrv.Handler = m.HTTPHandler(httpSrv.Handler)
	httpSrv.Addr = ":8080"
	fmt.Printf("Starting HTTP server on %s\n", httpSrv.Addr)
	err := httpSrv.ListenAndServe()
	if err != nil {
		log.Fatalf("httpSrv.ListenAndServe() failed with %s", err)
	}
}

func makeServerFromMux(mux *http.ServeMux) *http.Server {
	// set timeouts so that a slow or malicious client doesn't
	// hold resources forever
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
}

func makeHTTPToHTTPSRedirectServer() *http.Server {
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		newURI := "https://" + strings.Split(r.Host, ":")[0] + ":8443" + r.URL.String()
		http.Redirect(w, r, newURI, http.StatusFound)
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleRedirect)
	return makeServerFromMux(mux)
}
