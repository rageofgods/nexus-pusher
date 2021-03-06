package server

import (
	"github.com/google/uuid"
	"net/http"
	"nexus-pusher/internal/config"
	"nexus-pusher/internal/core"
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

type Message struct {
	ID       uuid.UUID `json:"id"`
	Response []string  `json:"response"`
	Complete bool      `json:"complete"`
}

type webService struct {
	cfg      *config.Server
	messages map[uuid.UUID]*Message
	jwtKey   []byte
	ver      *core.Version
}

func newWebService(cfg *config.Server, messages map[uuid.UUID]*Message, jwtKey []byte, v *core.Version) *webService {
	return &webService{cfg: cfg, messages: messages, jwtKey: jwtKey, ver: v}
}

const (
	// Limit uploaded json to 30mb
	maxBodySize int64 = 30 << 20
)
