package comps

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"io/ioutil"
	"net/http"
	"nexus-pusher/pkg/utils"
	"strings"
)

type Pypi struct {
	Server   string
	Path     string
	FileName string
	Name     string
	Version  string
}

func NewPypi(server string, path string, fileName string, name string, version string) *Pypi {
	return &Pypi{
		Server:   server,
		Path:     path,
		FileName: fileName,
		Name:     name,
		Version:  version,
	}
}

func (p Pypi) DownloadAsset() (*http.Response, error) {
	// Get PYPI component
	assetURL, err := p.assetDownloadURL()
	if err != nil {
		return nil, fmt.Errorf("DownloadAsset: %w", err)
	}

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("DownloadAsset: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	return HttpRetryClient(900).Do(req) // Set 15 min timeout to handle large files
}

func (p *Pypi) PrepareAssetToUpload(fileReader io.Reader) (string, io.Reader) {
	// Create multipart asset
	boundary := genRandomBoundary(32)
	fileName := p.FileName
	fileHeader := "Content-type: application/octet-stream"
	fileType := "pypi.asset"
	fileFormat := "--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n%s\r\n\r\n"
	bodyTop := fmt.Sprintf(fileFormat, boundary, fileType, fileName, fileHeader)
	bodyBottom := fmt.Sprintf("\r\n--%s--\r\n", boundary)

	body := io.MultiReader(strings.NewReader(bodyTop), fileReader, strings.NewReader(bodyBottom))
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	return contentType, body
}

func (p Pypi) assetDownloadURL() (string, error) {
	// Assemble initial request url to get asset version json
	requestURL := fmt.Sprintf("%spypi/%s/%s/json", p.Server, p.Name, p.Version)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return "", fmt.Errorf("assetDownloadURL: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := HttpRetryClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("assetDownloadURL: %w", err)
	}
	defer resp.Body.Close()

	// Check response for error
	if resp.StatusCode != http.StatusOK {
		return "", &utils.ContextError{
			Context: "assetDownloadURL",
			Err: fmt.Errorf("error: unable to get version for pypi asset. sending '%s' request: status code %d %v",
				resp.Request.Method,
				resp.StatusCode,
				resp.Request.URL),
		}
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("assetDownloadURL: %w", err)
	}

	var result map[string]interface{}
	// Try to unmarshal json as unstructured data
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("assetDownloadURL: %w", err)
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
							return "", &utils.ContextError{
								Context: "assetDownloadURL",
								Err:     fmt.Errorf("%s want: 'string', get: %T", errorText, value["url"]),
							}
						}
					}
				} else {
					return "", &utils.ContextError{
						Context: "assetDownloadURL",
						Err:     fmt.Errorf("%s want: 'string', get: %T", errorText, value["filename"]),
					}
				}
			} else {
				return "", &utils.ContextError{
					Context: "assetDownloadURL",
					Err:     fmt.Errorf("%s want: 'map[string]interface{}', get: %T", errorText, v),
				}
			}
		}
	} else {
		return "", &utils.ContextError{
			Context: "assetDownloadURL",
			Err:     fmt.Errorf("%s want: '[]interface{}', get: %T", errorText, result["urls"]),
		}
	}

	return "", &utils.ContextError{
		Context: "assetDownloadURL",
		Err: fmt.Errorf("error: unable to find component: %s version: %s at: %s",
			p.Name, p.Version, p.Server),
	}
}
