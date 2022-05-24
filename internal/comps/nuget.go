package comps

import (
	"fmt"
	"io"
	"net/http"
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
	req, err := http.NewRequest("GET", n.assetDownloadURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("DownloadAsset: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	return HttpRetryClient(180).Do(req) // Set 3 min timeout to handle files

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

func (n Nuget) assetDownloadURL() string {
	// Remove '+...' version part from download url
	version := strings.Split(n.Version, "+")[0]
	return fmt.Sprintf("%s%s/%s/%s.%s.nupkg",
		n.Server,
		strings.ToLower(n.Name),
		strings.ToLower(version),
		strings.ToLower(n.Name),
		strings.ToLower(version),
	)
}
