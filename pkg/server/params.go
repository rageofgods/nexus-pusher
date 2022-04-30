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

type uploadService struct {
	cfg      *config.Server
	messages map[uuid.UUID]*Message
}

func newUploadService(cfg *config.Server, messages map[uuid.UUID]*Message) *uploadService {
	return &uploadService{cfg: cfg, messages: messages}
}

const (
	// Limit uploaded json to 30mb
	maxBodySize int64 = 31457280
)
