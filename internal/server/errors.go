package server

import (
	"fmt"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func responseError(w http.ResponseWriter, err error, text string) {
	errorText := fmt.Sprintf("%s: %s", text, err.Error())
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusUnprocessableEntity)
	if err := json.NewEncoder(w).Encode(errorText); err != nil {
		log.Errorf("%v", err)
	}
	log.Errorf("%s", errorText)
}
