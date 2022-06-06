package comps

import (
	"fmt"
	"io"
	"net/http"
	"nexus-pusher/pkg/http_clients"
	"strings"
)

type Npm struct {
	Server   string
	Path     string
	FileName string
}

func NewNpm(server string, path string, fileName string) *Npm {
	return &Npm{
		Server:   server,
		Path:     path,
		FileName: fileName,
	}
}

func (n Npm) DownloadAsset() (*http.Response, error) {
	// Get NPM component
	req, err := http.NewRequest("GET", n.assetDownloadURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("DownloadAsset: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	return http_clients.HttpRetryClient(180).Do(req) // Set 3 min timeout to handle files

}

func (n *Npm) PrepareAssetToUpload(fileReader io.Reader) (string, io.Reader) {
	// Create multipart asset
	boundary := genRandomBoundary(32)
	fileName := n.FileName
	fileHeader := "Content-type: application/octet-stream"
	fileType := "npm.asset"
	fileFormat := "--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n%s\r\n\r\n"
	bodyTop := fmt.Sprintf(fileFormat, boundary, fileType, fileName, fileHeader)
	bodyBottom := fmt.Sprintf("\r\n--%s--\r\n", boundary)

	body := io.MultiReader(strings.NewReader(bodyTop), fileReader, strings.NewReader(bodyBottom))
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	return contentType, body
}

func (n Npm) assetDownloadURL() string {
	return fmt.Sprintf("%s%s", n.Server, n.Path)
}
