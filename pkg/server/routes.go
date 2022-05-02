package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"nexus-pusher/pkg/config"
)

func NewRouter(cfg *config.Server) *mux.Router {
	us := newWebService(cfg, make(map[uuid.UUID]*Message), genRandomJWTKey(32))
	var r = Routes{Routes: []Route{
		{"login", "GET", config.URIBase + config.URILogin, stub},
		{"refresh", "GET", config.URIBase + config.URIRefresh, stub},
		{"post-components", "POST", config.URIBase + config.URIComponents, us.components},
		{Name: "get-answer", Method: "GET", Pattern: config.URIBase + config.URIComponents, HandlerFunc: us.answerMessage},
	}}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range r.Routes {
		var handler http.Handler

		// Apply auth middlewares
		switch route.Name {
		case "login":
			// Setup JWT sign in middleware for index target
			handler = us.signInMiddle(route.HandlerFunc)
		case "refresh":
			// Refresh JWT token if it's still alive for client
			handler = us.refreshMiddle(route.HandlerFunc)
		default:
			// Default to auth the request
			handler = us.authMiddle(route.HandlerFunc)
		}

		// Setup logger middleware for all handler functions
		handler = loggerMiddle(handler, route.Name)
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}
	return router
}
