package comps

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"nexus-pusher/internal/comps/types"
)

func (s *NexusServer) uploadComponent(format ComponentType, c *http.Client, asset *NexusExportComponentAsset, repoName string) error {
	switch format {
	case NPM:
		// Download NPM component from official repo and return structured data
		npm := types.NewNpm(npmSrv, asset.Path)
		npmData, err := prepareToUpload(npm)
		if err != nil {
			return err
		}

		// Upload NPM component to Nexus repo
		if err := s.uploadNpmComponent(npmData.(*types.Npm), repoName, c, asset); err != nil {
			return err
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

func (s *NexusServer) uploadNpmComponent(npmData *types.Npm, repoName string, c *http.Client, asset *NexusExportComponentAsset) error {
	// Upload component to nexus repo
	srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
		s.BaseUrl,
		s.ApiComponentsUrl,
		repoName)
	req, err := http.NewRequest("POST", srvUrl, npmData.Content.Data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", npmData.Content.Type)
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
		log.Printf("Component %s succesfully uploaded to repository '%s' at server %s",
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
