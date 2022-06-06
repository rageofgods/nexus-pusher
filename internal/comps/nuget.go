package comps

import (
	"fmt"
	"github.com/goccy/go-json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"nexus-pusher/pkg/utils"
	"path"
	"strings"
)

type Nuget struct {
	Server   string
	FileName string
	Name     string
	Version  string
}

func NewNuget(server string, fileName string, name string, version string) *Nuget {
	return &Nuget{
		Server:   server,
		FileName: fileName,
		Name:     name,
		Version:  version,
	}
}

func (n Nuget) DownloadAsset() (*http.Response, error) {
	// Get NUGET component
	assetURL, err := n.assetDownloadURL()
	if err != nil {
		return nil, fmt.Errorf("DownloadAsset: %w", err)
	}

	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("DownloadAsset: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	return utils.HttpRetryClient(180).Do(req) // Set 3 min timeout to handle files
}

func (n Nuget) PrepareAssetToUpload(fileReader io.Reader) (string, io.Reader) {
	// Create multipart asset
	boundary := genRandomBoundary(32)
	fileName := n.FileName
	const fileHeader = "Content-type: application/octet-stream"
	const fileType = "nuget.asset"
	const fileFormat = "--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n%s\r\n\r\n"
	bodyTop := fmt.Sprintf(fileFormat, boundary, fileType, fileName, fileHeader)
	bodyBottom := fmt.Sprintf("\r\n--%s--\r\n", boundary)

	body := io.MultiReader(strings.NewReader(bodyTop), fileReader, strings.NewReader(bodyBottom))
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	return contentType, body
}

func (n Nuget) assetDownloadURL() (string, error) {
	// Remove '+...' version part from download url
	version := strings.Split(n.Version, "+")[0]

	// Check base server download url version
	baseDownloadUrl, err := n.checkVersion()
	if err != nil {
		return "", fmt.Errorf("assetDownloadURL: %w", err)
	}

	// If original server url is the same as returned one
	// format returned asset path following V2 standard
	if baseDownloadUrl == n.Server {
		return fmt.Sprintf("%s/package/%s/%s",
			removeLastSlash(baseDownloadUrl),
			strings.ToLower(n.Name),
			strings.ToLower(version),
		), nil
	}

	// Return V3 formatted path
	return fmt.Sprintf("%s/%s/%s/%s.%s.nupkg",
		removeLastSlash(baseDownloadUrl),
		strings.ToLower(n.Name),
		strings.ToLower(version),
		strings.ToLower(n.Name),
		strings.ToLower(version),
	), nil
}

//
func (n Nuget) checkVersion() (string, error) {
	parsedUrl, err := url.Parse(n.Server)
	if err != nil {
		return "", fmt.Errorf("checkVersion: %w", err)
	}

	// V3 check
	if path.Base(parsedUrl.Path) == "index.json" {
		baseDownloadUrl, err := n.baseUrlV3()
		if err != nil {
			return "", fmt.Errorf("checkVersion: %w", err)
		}
		// If we found valid v3 base download url, return it
		return baseDownloadUrl, nil
	}
	// if we can't find v3 url with index.json let's assume we
	// have v2 version and will try to process it later
	return n.Server, nil
}

func (n Nuget) baseUrlV3() (string, error) {
	req, err := http.NewRequest("GET", n.Server, nil)
	if err != nil {
		return "", fmt.Errorf("baseUrlV3: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	// Send request
	resp, err := utils.HttpRetryClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("baseUrlV3: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &utils.ContextError{
			Context: "baseUrlV3",
			Err: fmt.Errorf("sending '%s' request: status code %d %v",
				resp.Request.Method,
				resp.StatusCode,
				resp.Request.URL),
		}
	}

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("baseUrlV3: %w", err)
	}

	// Parse response
	respJson := &struct {
		Resources []struct {
			ID   string `json:"@id"`
			Type string `json:"@type"`
		} `json:"resources"`
	}{}
	if err := json.Unmarshal(body, &respJson); err != nil {
		return "", fmt.Errorf("baseUrlV3: %w", err)
	}

	// Try to find base download url in response
	for _, v := range respJson.Resources {
		if v.Type == "PackageBaseAddress/3.0.0" {
			return v.ID, nil
		}
	}

	return "", &utils.ContextError{
		Context: "baseUrlV3",
		Err: fmt.Errorf("unable to find base download path for provided url '%s' in nuget index.json response",
			n.Server),
	}
}

// removeLastSlash remove slash from last symbol if exists
func removeLastSlash(path string) string {
	if string(path[len(path)-1]) == "/" {
		return path[:len(path)-1]
	}
	return path
}
