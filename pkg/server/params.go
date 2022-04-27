package server

import (
	"net/http"
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

const (
	// Limit uploaded json to 30mb
	maxBodySize int64 = 31457280
)
