package comps

import (
	"fmt"
	"io"
	"net/http"
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

func (n Npm) DownloadComponent() (*http.Response, error) {
	// Get NPM component
	req, err := http.NewRequest("GET", n.assetDownloadURL(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	return HttpClient(120).Do(req)

}

func (n *Npm) PrepareDataToUpload(fileReader io.Reader) (string, io.Reader) {
	// Create multipart asset
	boundary := "MyMultiPartBoundary12345"
	fileName := n.FileName
	fileHeader := "Content-type: application/octet-stream"
	fileFormat := "--%s\r\nContent-Disposition: form-data; name=\"npm.asset\"; filename=\"@%s\"\r\n%s\r\n\r\n"
	filePart := fmt.Sprintf(fileFormat, boundary, fileName, fileHeader)
	bodyTop := fmt.Sprintf("%s", filePart)
	bodyBottom := fmt.Sprintf("\r\n--%s--\r\n", boundary)

	body := io.MultiReader(strings.NewReader(bodyTop), fileReader, strings.NewReader(bodyBottom))
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
	return contentType, body
}

func (n Npm) assetDownloadURL() string {
	return fmt.Sprintf("%s%s", n.Server, n.Path)
}
