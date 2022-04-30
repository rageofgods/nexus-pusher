package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"nexus-pusher/pkg/config"
)

func NewRouter(cfg *config.Server) *mux.Router {
	us := newUploadService(cfg, make(map[uuid.UUID]*Message))
	var r = Routes{Routes: []Route{
		{"index", "GET", "/", index},
		{"post-components", "POST", "/service/rest/v1/components", us.components},
		{Name: "get-answer", Method: "GET", Pattern: "/service/rest/v1/components", HandlerFunc: us.answerMessage},
	}}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range r.Routes {
		var handler http.Handler
		handler = logger(route.HandlerFunc, route.Name)
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}

	return router
}
