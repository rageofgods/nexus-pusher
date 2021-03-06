package core

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"nexus-pusher/internal/config"
	"nexus-pusher/pkg/http_clients"
	"nexus-pusher/pkg/utils"
	"time"
)

func (s *NexusServer) uploadComponent(format config.ComponentType,
	component *NexusExportComponent, repoName string) error {
	switch format.Lower() {
	case config.MAVEN2:
		maven2 := NewMaven2(component.ArtifactsSource, component)

		if len(maven2.Component.Assets) == 0 {
			return &utils.ContextError{
				Context: "uploadComponent",
				Err:     fmt.Errorf("zero valid maven artifacts was found after assets filter"),
			}
		}

		// Start to download data and convert it to multipart stream
		contentType, uploadBody, responses, err := prepareToUploadComponent(maven2)
		if err != nil {
			return fmt.Errorf("uploadComponent: %w", err)
		}

		// Close all responses body
		defer func() {
			for _, resp := range responses {
				resp.Body.Close()
			}
		}()

		// Upload component to target nexus server
		if err := s.uploadComponentWithType(repoName, component.FullName(), contentType, uploadBody); err != nil {
			return fmt.Errorf("uploadComponent: %w", err)
		}
	}

	return nil
}

func (s *NexusServer) uploadAsset(format config.ComponentType, asset *NexusExportComponentAsset,
	repoName string, artifactsSource string) error {
	switch format.Lower() {
	case config.NPM:
		npm := NewNpm(artifactsSource, asset.Path, asset.FileName)

		// Start to download data and convert it to multipart stream
		contentType, uploadBody, resp, err := prepareToUploadAsset(npm)
		if err != nil {
			return fmt.Errorf("uploadAsset: %w", err)
		}
		defer resp.Body.Close()

		// Upload component to target nexus server
		if err := s.uploadComponentWithType(repoName, asset.FullName(), contentType, uploadBody); err != nil {
			return fmt.Errorf("uploadAsset: %w", err)
		}

	case config.PYPI:
		pypi := NewPypi(artifactsSource, asset.Path, asset.FileName, asset.Name, asset.Version)

		// Start to download data and convert it to multipart stream
		contentType, uploadBody, resp, err := prepareToUploadAsset(pypi)
		if err != nil {
			return fmt.Errorf("uploadAsset: %w", err)
		}
		defer resp.Body.Close()

		// Upload component to target nexus server
		if err := s.uploadComponentWithType(repoName, asset.FullName(), contentType, uploadBody); err != nil {
			return fmt.Errorf("uploadAsset: %w", err)
		}

	case config.NUGET:
		nuget := NewNuget(artifactsSource, asset.FileName, asset.Name, asset.Version)

		// Start to download data and convert it to multipart stream
		contentType, uploadBody, resp, err := prepareToUploadAsset(nuget)
		if err != nil {
			return fmt.Errorf("uploadAsset: %w", err)
		}
		defer resp.Body.Close()

		// Upload component to target nexus server
		if err := s.uploadComponentWithType(repoName, asset.FullName(), contentType, uploadBody); err != nil {
			return fmt.Errorf("uploadAsset: %w", err)
		}
	}

	return nil
}

// Download component with all assets following provided interface type
func prepareToUploadComponent(c config.Componenter) (string, io.Reader, []*http.Response, error) {
	// Start downloading component from remote repo
	responses, err := c.DownloadComponent()
	if err != nil {
		return "", nil, nil, fmt.Errorf("prepareToUploadComponent: %w", err)
	}

	for _, resp := range responses {
		if resp.StatusCode != http.StatusOK {
			return "", nil, nil, &utils.ContextError{
				Context: "prepareToUploadComponent",
				Err: fmt.Errorf("unable to download asset. sending '%s' request: status code %d %v",
					resp.Request.Method,
					resp.StatusCode,
					resp.Request.URL),
			}
		}
	}

	// Convert to multipart component specific type on the fly
	// and return converted body with correct content type
	contentType, uploadBody := c.PrepareComponentToUpload(responses)
	return contentType, uploadBody, responses, nil
}

// Download asset following provided interface type
func prepareToUploadAsset(a config.Asseter) (string, io.Reader, *http.Response, error) {
	// Start downloading asset from remote repo
	resp, err := a.DownloadAsset()
	if err != nil {
		return "", nil, nil, fmt.Errorf("prepareToUploadAsset: %w", err)
	}

	// Check http response ok status
	if resp.StatusCode != http.StatusOK {
		return "", nil, nil, &utils.ContextError{
			Context: "prepareToUploadAsset",
			Err: fmt.Errorf("unable to download asset. sending '%s' request: status code %d %v",
				resp.Request.Method,
				resp.StatusCode,
				resp.Request.URL),
		}
	}

	// Convert to multipart component specific type on the fly
	// and return converted body with correct content type
	contentType, uploadBody := a.PrepareAssetToUpload(resp.Body)
	return contentType, uploadBody, resp, nil
}

func (s *NexusServer) uploadComponentWithType(repoName string, cPath string, contentType string, body io.Reader) error {
	// Upload component to nexus repo
	srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
		s.BaseUrl,
		s.ApiComponentsUrl,
		repoName)
	req, err := http.NewRequest("POST", srvUrl, body)
	if err != nil {
		return fmt.Errorf("uploadComponentWithType: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(s.Username, s.Password)

	// Start uploading component to remote nexus
	// Set 15 min timeout to handle large files
	// We can't use retryable client here because
	// of direct stream data incompatibility so
	// let's implement simple retry behaviour
	var resp *http.Response
	for i := 1; i <= 4; {
		resp, err = http_clients.HttpClient(900).Do(req)
		if err != nil {
			if i == 4 {
				// if it's last iteration, return error
				return fmt.Errorf("uploadComponentWithType: %w", err)
			} else {
				// if we got error, log it and continue attempts
				log.WithFields(log.Fields{"retry": i}).Errorf("uploadComponentWithType: %v", err)
				// Sleep a little to relax error
				time.Sleep(100 * time.Millisecond)
			}
			i++
		} else {
			break
		}
	}

	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusNoContent {
		// Read response body with error
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("uploadComponentWithType: %w", err)
		}

		// Create formatted message
		const msg = "unable to upload component %s to repository '%s' at server %s. Reason: %s. Response: %s"

		// Return error
		return fmt.Errorf(msg, cPath, repoName, s.Host, resp.Status, string(body))
	} else {
		log.Printf("Component %s successfully uploaded to repository '%s' at server %s",
			cPath,
			repoName,
			s.Host)
	}

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return fmt.Errorf("uploadComponentWithType: %w", err)
	}

	return nil
}
