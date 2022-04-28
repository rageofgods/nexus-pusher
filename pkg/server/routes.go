package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

func NewRouter() *mux.Router {
	var r = Routes{Routes: []Route{
		{"Index", "GET", "/", index},
		{"Components", "POST", "/service/rest/v1/components", components},
	}}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range r.Routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = logger(handler, route.Name)
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}

	return router
}
