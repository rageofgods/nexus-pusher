package server

import (
	"github.com/gorilla/mux"
	"net/http"
	"nexus-pusher/pkg/config"
)

func NewRouter(cfg *config.Server) *mux.Router {
	rc := newRouteConfig(cfg)
	var r = Routes{Routes: []Route{
		{"Index", "GET", "/", index},
		{"Components", "POST", "/service/rest/v1/components", rc.components},
	}}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range r.Routes {
		var handler http.Handler
		handler = logger(route.HandlerFunc, route.Name)
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}

	return router
}
