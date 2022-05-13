package server

import (
	"log"
	"net/http"
	"strings"
	"time"
)

func loggerMiddle(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf(">>> Serving request... Client ip: %s, Method: %s, RequestURI: %s, Route name: %s",
			strings.Split(r.RemoteAddr, ":")[0],
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
