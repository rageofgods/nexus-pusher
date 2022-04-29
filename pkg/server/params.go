package server

import (
	"net/http"
	"nexus-pusher/pkg/config"
)

type Routes struct {
	Routes []Route
}

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type handlerConfig struct {
	cfg *config.Server
}

func newRouteConfig(cfg *config.Server) *handlerConfig {
	return &handlerConfig{cfg: cfg}
}

const (
	// Limit uploaded json to 30mb
	maxBodySize int64 = 31457280
)
