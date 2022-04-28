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

// Upload components to remote nexus
func components(w http.ResponseWriter, r *http.Request) {
	nec := &comps.NexusExportComponents{}
	// Get repository parameter from URL
	repo := r.URL.Query().Get("repository")
	if repo == "" {
		responseError(w, fmt.Errorf("parameter 'repository' is required"), "error")
		return
	}
	// Read request body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		log.Printf("%v", err)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Printf("%v", err)
	}
	// Try to decode body to NexusExportComponents struct
	if err := json.Unmarshal(body, nec); err != nil {
		responseError(w, err, "unable to decode request data")
		return
	}
	// Upload components
	s := comps.NewNexusServer(nec.NexusServer.Username, nec.NexusServer.Password,
		nec.NexusServer.Host, nec.NexusServer.BaseUrl, nec.NexusServer.ApiComponentsUrl)
	results := s.UploadComponents(comps.HttpClient(), nec, repo)
	var errorResults string
	for _, v := range results {
		if v.Err != nil {
			errorResults += fmt.Sprintf("%s\n", v.Err.Error())
		}
	}
	if errorResults != "" {
		responseError(w, fmt.Errorf(errorResults), "error")
		return
	}
	// Generate final response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "Successfully uploaded"); err != nil {
		log.Printf("%v", err)
	}
}
