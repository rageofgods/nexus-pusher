package comps

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

type Pypi struct {
	Server   string
	Path     string
	FileName string
	Name     string
	Version  string
	Content  *PypiData
}

type PypiData struct {
	Type string
	Data *bytes.Buffer
}

func NewPypi(server string, path string, fileName string, name string, version string) *Pypi {
	return &Pypi{
		Server:   server,
		Path:     path,
		FileName: fileName,
		Name:     name,
		Version:  version,
		Content:  &PypiData{Data: &bytes.Buffer{}}}
}

func (p Pypi) DownloadComponent() ([]byte, error) {
	// Get PYPI component
	assetURL, err := p.assetDownloadURL()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	resp, err := HttpClient(120).Do(req) // Set 120 sec timeout to handle large files
	if err != nil {
		return nil, err
	}

	// Check response for error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: unable to download pypi asset. sending '%s' request: status code %d %v",
			resp.Request.Method,
			resp.StatusCode,
			resp.Request.URL)
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Close body
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}

func (p *Pypi) PrepareDataToUpload(body []byte) (interface{}, error) {
	writer := multipart.NewWriter(p.Content.Data)
	part, err := writer.CreateFormFile("pypi.asset", fmt.Sprintf("@%s", p.FileName))
	if err != nil {
		return nil, err
	}

	// Convert body bytes to Reader interface
	r := bytes.NewReader(body)
	if _, err := io.Copy(part, r); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}
	p.Content.Type = writer.FormDataContentType()
	return p, nil
}

func (p Pypi) assetDownloadURL() (string, error) {
	// Assemble initial request url to get asset version json
	requestURL := fmt.Sprintf("%spypi/%s/%s/json", pypiSrv, p.Name, p.Version)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := HttpClient().Do(req)
	if err != nil {
		return "", err
	}

	// Check response for error
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error: unable to get version for pypi asset. sending '%s' request: status code %d %v",
			resp.Request.Method,
			resp.StatusCode,
			resp.Request.URL)
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Close body
	if err := resp.Body.Close(); err != nil {
		return "", err
	}

	var result map[string]interface{}
	// Try to unmarshal json as unstructured data
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Validate JSON data to correct format
	errorText := "error: unable to validate json data when polling pypi version."
	if urls, ok := result["urls"].([]interface{}); ok { // Get slice of URLs
		for _, v := range urls {
			if value, ok := v.(map[string]interface{}); ok { // Check type assertion
				if filename, ok := value["filename"].(string); ok { // Get filename string
					if filename == strings.Trim(p.FileName, "@") { // If we found wanted file
						if url, ok := value["url"].(string); ok { // Get URL string
							return url, nil // Return found URL
						} else {
							return "", fmt.Errorf("%s want: 'string', get: %T", errorText, value["url"])
						}
					}
				} else {
					return "", fmt.Errorf("%s want: 'string', get: %T", errorText, value["filename"])
				}
			} else {
				return "", fmt.Errorf("%s want: 'map[string]interface{}', get: %T", errorText, v)
			}
		}
	} else {
		return "", fmt.Errorf("%s want: '[]interface{}', get: %T", errorText, result["urls"])
	}

	return "", fmt.Errorf("error: unable to find component: %s version: %s at: %s",
		p.Name, p.Version, pypiSrv)
}
