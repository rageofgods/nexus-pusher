package server

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
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
func (u *uploadService) components(w http.ResponseWriter, r *http.Request) {
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
	// Send response
	msg, err := u.genMessageWithId()
	if err != nil {
		responseError(w, err, "error")
		return
	}
	msg.Response = "Successfully uploaded"
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		responseError(w, err, "error message encode")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Upload components
	go func() {
		s := comps.NewNexusServer(nec.NexusServer.Username, nec.NexusServer.Password,
			nec.NexusServer.Host, nec.NexusServer.BaseUrl, nec.NexusServer.ApiComponentsUrl)
		results := s.UploadComponents(comps.HttpClient(), nec, repo, u.cfg)

		var errorsCounter int
		var errorsText []string
		for _, v := range results {
			if v.Err != nil {
				errorsCounter++
				errorsText = append(errorsText,
					fmt.Sprintf("request error: %s\n for component '%s'", v.Err.Error(), v.ComponentPath))
			}
		}
		if errorsCounter != 0 {
			log.Printf("Upload request complete with %d errors:", errorsCounter)
			for _, v := range errorsText {
				log.Println(v)
			}
			// Set complete flag to current client request
			if err := u.completeById(msg.ID, errorsText...); err != nil {
				log.Printf("%v", err)
			}
		} else {
			log.Printf("Upload request succesfully complete.")
			// Set complete flag to current client request
			if err := u.completeById(msg.ID, fmt.Sprintf("All assets successfully uploaded for repository '%s'",
				repo)); err != nil {
				log.Printf("%v", err)
			}
		}
	}()
}

func (u *uploadService) answerMessage(w http.ResponseWriter, r *http.Request) {
	// Check uuid parameter
	data := r.URL.Query().Get("uuid")
	if data == "" {
		responseError(w, fmt.Errorf("parameter 'uuid' is required"), "error")
		return
	}
	if _, err := io.Copy(ioutil.Discard, r.Body); err != nil {
		responseError(w, err, "error")
		return
	}
	if err := r.Body.Close(); err != nil {
		responseError(w, err, "error")
		return
	}
	// Convert string to uuid
	id, err := uuid.Parse(data)
	if err != nil {
		responseError(w, err, "unable to parse uuid")
		return
	}
	// Search message by uuid
	msg, err := u.searchById(id)
	if err != nil {
		responseError(w, err, "error")
		return
	}
	// Send response
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		responseError(w, err, "error message encode")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// Clear Message data if we complete
	if msg.Complete {
		u.deleteById(id)
	}
}
