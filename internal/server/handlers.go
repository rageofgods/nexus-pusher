package server

import (
	"fmt"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"nexus-pusher/internal/comps"
	"strings"
)

func stub(w http.ResponseWriter, r *http.Request) {
	_ = r // ignore request here
	if _, err := fmt.Fprintln(w, "Welcome to nexus-pusher."); err != nil {
		log.Errorf("%v", err)
	}
}

func status(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "Status OK"); err != nil {
		log.Errorf("%v", err)
	}
}

func (u *webService) version(w http.ResponseWriter, _ *http.Request) {
	if err := json.NewEncoder(w).Encode(u.ver); err != nil {
		responseError(w, err, "error message encode")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

// Upload components to remote nexus
func (u *webService) components(w http.ResponseWriter, r *http.Request) {
	nec := &comps.NexusExportComponents{}
	// Get repository parameter from URL
	urlParam := r.URL.Query().Get("repository")

	var repo string
	// Check for valid repository name in user request following nexus supported pattern
	if !isValidNexusRepoName(urlParam) {
		responseError(w, fmt.Errorf("only letters, digits, underscores(_),"+
			" hyphens(-), and dots(.) are allowed in repository name. but got: '%s'", urlParam), "error")
		return
	}
	// Sanitize user input for repo name
	repo = strings.Replace(urlParam, "\n", "", -1)
	repo = strings.Replace(repo, "\r", "", -1)

	// Read request body
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		log.Errorf("%v", err)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Errorf("%v", err)
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
		results := s.UploadComponents(nec, repo, u.cfg)

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
			log.Warnf("Upload request complete with %d errors:", errorsCounter)
			for _, v := range errorsText {
				log.Warnln(v)
			}
			// Set complete flag to current client request
			if err := u.completeById(msg.ID, errorsText...); err != nil {
				log.Errorf("%v", err)
			}
		} else {
			log.Printf("Upload request successfully complete.")
			// Set complete flag to current client request
			if err := u.completeById(msg.ID, fmt.Sprintf("All assets successfully uploaded for repository '%s'",
				repo)); err != nil {
				log.Errorf("%v", err)
			}
		}
	}()
}

func (u *webService) answerMessage(w http.ResponseWriter, r *http.Request) {
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
