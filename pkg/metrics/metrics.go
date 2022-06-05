package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Registry struct {
	registry *prometheus.Registry
	uri      string
	port     string
}

func NewRegister(uri string, port string) *Registry {
	return &Registry{registry: prometheus.NewRegistry(), uri: uri, port: port}
}

func (r Registry) StartServing() {
	handler := promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
	http.Handle(r.uri, handler)
	go func() {
		log.Fatal(http.ListenAndServe(":"+r.port, nil))
	}()
	log.WithFields(log.Fields{
		"uri":  r.uri,
		"port": r.port,
	}).Info("Running prometheus metrics exporter")
}

func (r Registry) Registry() *prometheus.Registry {
	return r.registry
}
