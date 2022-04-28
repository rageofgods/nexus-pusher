package server

import (
	"log"
	"net/http"
	"time"
)

func logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf(">>> Serving request... Method: %s, RequestURI: %s, Route name: %s",
			r.Method,
			r.RequestURI,
			name)
		inner.ServeHTTP(w, r)
		log.Printf(
			"<<< Request complete. Took %s",
			time.Since(start),
		)
	})
}
