package comps

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

func (s *NexusServer) GetComponents(c *http.Client,
	ncs []*NexusComponent,
	repoName string,
	contToken ...string) []*NexusComponent {

	var srvUrl string
	if len(contToken) != 0 {
		srvUrl = fmt.Sprintf("%s%s%s?repository=%s&continuationToken=%s", s.Host,
			s.BaseUrl,
			s.ApiComponentsUrl,
			repoName,
			contToken[0])
	} else {
		srvUrl = fmt.Sprintf("%s%s%s?repository=%s", s.Host,
			s.BaseUrl,
			s.ApiComponentsUrl,
			repoName)
	}

	body, err := s.SendRequest(srvUrl, "GET", c, nil)
	if err != nil {
		log.Fatal(err)
	}

	var nc NexusComponents
	if err := json.Unmarshal(body, &nc); err != nil {
		log.Fatal(err)
	}
	ncs = append(ncs, nc.Items...)

	// Send log message every 100 new components
	if len(ncs) <= 10 || len(ncs)%100 == 0 {
		log.Printf("Analyzing repo '%s', please wait... Processed %d assets.\n", repoName, len(ncs))
	}

	//if len(ncs) > 2000 {
	//	return ncs
	//}

	// Iterating over all API pages
	if nc.ContinuationToken != "" {
		ncs = s.GetComponents(c, ncs, repoName, nc.ContinuationToken)
	}

	return ncs
}

// UploadComponents is used to upload nexus artifacts following by 'nec' list
func (s *NexusServer) UploadComponents(c *http.Client, nec *NexusExportComponents, repoName string) []UploadResult {
	maxParallel := 10
	limitChan := make(chan struct{}, maxParallel)
	resultsChan := make(chan *UploadResult)

	defer func() {
		close(limitChan)
		close(resultsChan)
	}()

	var resultsCounter int
	for _, v := range nec.Items {
		for _, vv := range v.Assets {
			resultsCounter++
			go func(format componentType, c *http.Client, asset *NexusExportComponentAsset, repoName string) {
				limitChan <- struct{}{}
				result := &UploadResult{}
				if err := s.uploadComponent(format, c, asset, repoName); err != nil {
					log.Printf("%v", err)
					result = &UploadResult{err}
				}
				resultsChan <- result
				<-limitChan
			}(componentType(v.Format), c, vv, repoName)
		}
	}
	var results []UploadResult
	for {
		result := <-resultsChan
		results = append(results, *result)

		// if we've reached the expected amount of results then stop
		if len(results) == resultsCounter {
			break
		}
	}
	return results
}

func (s *NexusServer) uploadComponent(format componentType,
	c *http.Client,
	asset *NexusExportComponentAsset,
	repoName string) error {
	switch format {
	case npm:
		// Download NPM component from official repo
		data, conType, err := downloadComponent(npm, asset.Path)
		if err != nil {
			return err
		}
		// Upload component to nexus repo
		srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
			s.BaseUrl,
			s.ApiComponentsUrl,
			repoName)
		req, err := http.NewRequest("POST", srvUrl, data)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", conType)
		req.SetBasicAuth(s.Username, s.Password)
		resp, err := c.Do(req)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		// Check server response
		if resp.StatusCode != http.StatusNoContent {
			log.Printf("unable to upload component %s to repository '%s' at server %s",
				asset.Path,
				repoName,
				s.Host)
			return fmt.Errorf("unable to upload component %s to repository '%s' at server %s",
				asset.Path,
				repoName,
				s.Host)
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
	}
	return nil
}

func downloadComponent(cmpType componentType, cmpPath string) (*bytes.Buffer, string, error) {
	var resp *http.Response
	switch cmpType {
	case npm:
		npmSrv := npmSrv
		var err error
		resp, err = http.Get(fmt.Sprintf("%s%s", npmSrv, cmpPath))
		defer resp.Body.Close()
		if err != nil {
			return nil, "", err
		}
		// Check server response
		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("bad status: %s", resp.Status)
		}
	}
	body := &bytes.Buffer{}
	conType, err := createFormMultipart(body, componentNameFromPath(cmpPath), &resp.Body)
	if err != nil {
		return nil, "", err
	}
	return body, conType, nil
}

func createFormMultipart(v *bytes.Buffer, cmpName string, body *io.ReadCloser) (string, error) {
	writer := multipart.NewWriter(v)
	part, _ := writer.CreateFormFile("r.asset", fmt.Sprintf("@%s", cmpName))
	if _, err := io.Copy(part, *body); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return writer.FormDataContentType(), nil
}

func componentNameFromPath(cmpPath string) string {
	cmpPathSplit := strings.Split(cmpPath, "/")
	return cmpPathSplit[len(cmpPathSplit)-1]
}

func HttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			MaxIdleConns:        100,
		},
		Timeout: 10 * time.Minute,
	}
	return client
}

func (s *NexusServer) SendRequest(srvUrl string, method string, c *http.Client, b io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, srvUrl, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.Username, s.Password)
	// Send request
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error processing request: status code %d", resp.StatusCode)
	}
	// Read all body data
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}
