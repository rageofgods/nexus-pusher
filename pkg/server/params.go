package server

import (
	"github.com/google/uuid"
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

type Message struct {
	ID       uuid.UUID `json:"id"`
	Response string    `json:"response"`
	Complete bool      `json:"complete"`
}

type webService struct {
	cfg      *config.Server
	messages map[uuid.UUID]*Message
	jwtKey   []byte
}

func newWebService(cfg *config.Server, messages map[uuid.UUID]*Message, jwtKey []byte) *webService {
	return &webService{cfg: cfg, messages: messages, jwtKey: jwtKey}
}

const (
	// Limit uploaded json to 30mb
	maxBodySize int64 = 31457280
	// Set JWT token TTL in minutes
	jwtTokenTTL = 5
	// Set JWT token Cookie name
	jwtCookieName = "token"
	// Set JWT token refresh window in seconds
	jwtTokenRefreshWindow = 30
)
