package server

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func loggerMiddle(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.WithFields(log.Fields{
			"client_ip":  strings.Split(r.RemoteAddr, ":")[0],
			"route_name": name,
		}).Infof(">>> Serving request... %s: '%s'", r.Method, r.RequestURI)
		// Serve original request
		inner.ServeHTTP(w, r)

		log.WithFields(log.Fields{"took": time.Since(start)}).Info("<<< Request complete.")
	})
}
