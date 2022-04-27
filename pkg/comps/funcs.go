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
	"net/url"
	"strings"
	"time"
)

func (s *NexusServer) GetComponents(client *http.Client,
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
	v := url.Values{}
	//pass the values to the request's body
	req, err := http.NewRequest("GET", srvUrl, strings.NewReader(v.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.Username, s.Password)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Couldn't parse response body. %+v", err)
	}

	var nc NexusComponents
	if err := json.Unmarshal(body, &nc); err != nil {
		log.Fatal(err)
	}

	if err := resp.Body.Close(); err != nil {
		log.Fatal(err)
	}
	ncs = append(ncs, nc.Items...)

	// Send log message every 100 new comps
	if len(ncs) <= 10 || len(ncs)%100 == 0 {
		log.Printf("Analyzing repo '%s', please wait... Processed %d assets.\n", repoName, len(ncs))
	}

	if len(ncs) > 200 {
		return ncs
	}

	// Iterating over all API pages
	if nc.ContinuationToken != "" {
		ncs = s.GetComponents(client, ncs, repoName, nc.ContinuationToken)
	}

	return ncs
}

func (s *NexusServer) UploadComponents(client *http.Client, nec *NexusExportComponents) error {
	for _, v := range nec.Items {
		for _, vv := range v.Assets {
			switch v.Format {
			case "npm":
				data, conType, err := downloadComponent(npm, vv.Path)
				if err != nil {
					return err
				}
				srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
					s.BaseUrl,
					s.ApiComponentsUrl,
					"npm-test2")
				req, err := http.NewRequest("POST", srvUrl, data)
				if err != nil {
					log.Fatal(err)
				}
				req.Header.Set("Content-Type", conType)
				req.SetBasicAuth(s.Username, s.Password)
				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				// Check server response
				if resp.StatusCode != http.StatusNoContent {
					log.Printf("unable to upload component: %s", vv.Path)
					continue
				}
				if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
					log.Fatal(err)
				}
				if err := resp.Body.Close(); err != nil {
					log.Fatal(err)
				}
			}
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
		Timeout: 10 * time.Second,
	}
	return client
}
