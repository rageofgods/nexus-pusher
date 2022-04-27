package server

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"nexus-pusher/pkg/comps"
)

func index(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprintln(w, "Welcome to nexus-pusher version: "); err != nil {
		log.Printf("%v", err)
	}
}

func components(w http.ResponseWriter, r *http.Request) {
	nec := &comps.NexusExportComponents{}
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		log.Printf("%v", err)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Printf("%v", err)
	}
	if err := json.Unmarshal(body, nec); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Printf("%v", err)
		}
		if _, err := fmt.Fprintf(w, "%v", err); err != nil {
			log.Printf("%v", err)
		}
		return
	}

	s := comps.NewNexusServer(nec.NexusServer.Username, nec.NexusServer.Password,
		nec.NexusServer.Host, nec.NexusServer.BaseUrl, nec.NexusServer.ApiComponentsUrl)
	if err := s.UploadComponents(comps.HttpClient(), nec); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity) // unprocessable entity
		log.Printf("%v", err)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Printf("%v", err)
		}
		if _, err := fmt.Fprintf(w, "%v", err); err != nil {
			log.Printf("%v", err)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "Successfully uploaded"); err != nil {
		log.Printf("%v", err)
	}
}
