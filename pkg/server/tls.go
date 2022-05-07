package server

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"net/http"
	"nexus-pusher/pkg/config"
	"strings"
	"time"
)

// RunAutoCertServer run TLS server with Let's Encrypt auto cert manager
func RunAutoCertServer(cfg *config.Server) {
	// Setup cache directory to store certificate
	c := autocert.DirCache("certs")
	// Generating autocert manager to handle let's encrypt api calls
	m := autocert.Manager{
		Cache:      c,
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.TLS.DomainName),
		// Let's encrypt staging environment
		//Client: &acme.Client{
		//	DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		//},
	}

	s := &http.Server{
		Addr:      fmt.Sprintf("%s:%s", cfg.BindAddress, cfg.Port),
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		Handler:   NewRouter(cfg),
	}

	// Run TLS server is dedicated goroutine to allow running both http/https
	go func() {
		log.Fatal(s.ListenAndServeTLS("", ""))
	}()

	// Setup and run http server on port :80 to allow let's encrypt api callback
	// Also, redirect any requests to TLS server
	httpSrv := makeHTTPToHTTPSRedirectServer(cfg.Port)
	httpSrv.Handler = m.HTTPHandler(httpSrv.Handler)
	httpSrv.Addr = ":http"
	log.Fatal(httpSrv.ListenAndServe())
}

// RunStaticCertServer run TLS server with static key/cert provided as a files
func RunStaticCertServer(cfg *config.Server) {
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf("%s:%s",
		cfg.BindAddress,
		cfg.Port),
		cfg.TLS.CertPath,
		cfg.TLS.KeyPath,
		NewRouter(cfg)))
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

func makeHTTPToHTTPSRedirectServer(port string) *http.Server {
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		newURI := fmt.Sprintf("https://%s:%s%s", strings.Split(r.Host, ":")[0], port, r.URL.String())
		http.Redirect(w, r, newURI, http.StatusFound)
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", handleRedirect)
	return makeServerFromMux(mux)
}
