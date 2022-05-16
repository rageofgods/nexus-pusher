package comps

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *NexusServer) uploadComponent(format ComponentType, c *http.Client, asset *NexusExportComponentAsset, repoName string) error {
	switch format {
	case NPM:
		// Download NPM component from official repo and return structured data
		npm := NewNpm(npmSrv, asset.Path, asset.FileName)
		npmData, err := prepareToUpload(npm)
		if err != nil {
			return err
		}

		// Check returned interface type
		if nd, ok := npmData.(*Npm); ok {
			// Upload NPM component to Nexus repo
			if err := s.uploadComponentWithType(nd, repoName, c, asset); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("error: wrong data interface type provided. want: 'npm', get: %T", npmData)
		}
	case PYPI:
		// Download PYPI component from official repo and return structured data
		pypi := NewPypi(pypiSrv, asset.Path, asset.FileName, asset.Name, asset.Version)
		pypiData, err := prepareToUpload(pypi)
		if err != nil {
			return err
		}

		// Check returned interface type
		if nd, ok := pypiData.(*Pypi); ok {
			// Upload PYPI component to Nexus repo
			if err := s.uploadComponentWithType(nd, repoName, c, asset); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("error: wrong data interface type provided. want: 'pypi', get: %T", pypiData)
		}
	}
	return nil
}

// Download component following provided interface type
func prepareToUpload(t Typer) (interface{}, error) {
	b, err := t.DownloadComponent()
	if err != nil {
		return nil, err
	}

	data, err := t.PrepareDataToUpload(b)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *NexusServer) uploadComponentWithType(data interface{}, repoName string, c *http.Client, asset *NexusExportComponentAsset) error {
	var contentData *bytes.Buffer
	var contentType string
	switch t := data.(type) {
	case *Npm:
		contentData = t.Content.Data
		contentType = t.Content.Type
	case *Pypi:
		contentData = t.Content.Data
		contentType = t.Content.Type
	default:
		return fmt.Errorf("error: unknown component type provided %T", data)
	}

	// Upload component to nexus repo
	srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
		s.BaseUrl,
		s.ApiComponentsUrl,
		repoName)
	req, err := http.NewRequest("POST", srvUrl, contentData)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(s.Username, s.Password)

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%v", err)
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
		return fmt.Errorf("%v", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
