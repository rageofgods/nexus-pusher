package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"nexus-pusher/internal/comps"
	"nexus-pusher/internal/config"
)

func NewRouter(cfg *config.Server, v *comps.Version) *mux.Router {
	us := newWebService(cfg, make(map[uuid.UUID]*Message), genRandomJWTKey(32), v)
	var r = Routes{Routes: []Route{
		{"login", "GET", config.URIBase + config.URILogin, stub},
		{"refresh", "GET", config.URIBase + config.URIRefresh, stub},
		{Name: "status", Method: "GET", Pattern: config.URIBase + config.URIStatus, HandlerFunc: status},
		{Name: "version", Method: "GET", Pattern: config.URIBase + config.URIVersion, HandlerFunc: us.version},
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
		case "status":
			// Skip authentication for 'status' requests
			handler = route.HandlerFunc
		case "version":
			// Skip authentication for 'status' requests
			handler = route.HandlerFunc
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
