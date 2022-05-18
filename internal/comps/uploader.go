package comps

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *NexusServer) uploadComponent(format ComponentType, c *http.Client, asset *NexusExportComponentAsset,
	repoName string) error {
	// Create multipart writer to connect it to the pipe and modify incoming
	// binary data from remote external repository (i.e PYPY, NPM, etc) on the fly
	//multipartWriter := multipart.NewWriter(outerPipeWriter)

	switch format {
	case NPM:
		// Download NPM component from official repo and return structured data
		npm := NewNpm(npmSrv, asset.Path, asset.FileName)

		// Start to download data and convert it to multipart stream
		contentType, uploadBody, resp, err := prepareToUpload(npm)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Upload component to target nexus server
		if err := s.uploadComponentWithType(repoName, c, asset, contentType, uploadBody); err != nil {
			return err
		}

	case PYPI:
		// Download PYPI component from official repo and return structured data
		pypi := NewPypi(pypiSrv, asset.Path, asset.FileName, asset.Name, asset.Version)

		// Start to download data and convert it to multipart stream
		contentType, uploadBody, resp, err := prepareToUpload(pypi)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Upload component to target nexus server
		if err := s.uploadComponentWithType(repoName, c, asset, contentType, uploadBody); err != nil {
			return err
		}
	}

	return nil
}

// Download component following provided interface type
func prepareToUpload(t Typer) (string, io.Reader, *http.Response, error) {
	// Start downloading component from remote repo
	resp, err := t.DownloadComponent()
	if err != nil {
		return "", nil, nil, err
	}

	// Check http response ok status
	if resp.StatusCode != http.StatusOK {
		return "", nil, nil, fmt.Errorf("error: unable to download npm asset. sending '%s' request: status code %d %v",
			resp.Request.Method,
			resp.StatusCode,
			resp.Request.URL)
	}

	contentType, uploadBody := t.PrepareDataToUpload(resp.Body)
	return contentType, uploadBody, resp, nil
}

func (s *NexusServer) uploadComponentWithType(repoName string, c *http.Client,
	asset *NexusExportComponentAsset, contentType string, body io.Reader) error {
	// Upload component to nexus repo
	srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
		s.BaseUrl,
		s.ApiComponentsUrl,
		repoName)
	req, err := http.NewRequest("POST", srvUrl, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(s.Username, s.Password)

	// Start uploading component to remote nexus
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Check server response
	if resp.StatusCode != http.StatusNoContent {
		log.Printf("error: unable to upload component %s to repository '%s' at server %s. Reason: %s",
			asset.Path,
			repoName,
			s.Host,
			resp.Status)
		return fmt.Errorf("error: unable to upload component %s to repository '%s' at server %s. Reason: %s",
			asset.Path,
			repoName,
			s.Host,
			resp.Status)
	} else {
		log.Printf("Component %s successfully uploaded to repository '%s' at server %s",
			asset.Path,
			repoName,
			s.Host)
	}

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
