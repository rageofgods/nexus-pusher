package comps

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type Npm struct {
	Server   string
	Path     string
	FileName string
	Content  *NpmData
}

type NpmData struct {
	Type string
	Data *bytes.Buffer
}

func NewNpm(server string, path string, fileName string) *Npm {
	return &Npm{
		Server:   server,
		Path:     path,
		FileName: fileName,
		Content:  &NpmData{Data: &bytes.Buffer{}}}
}

func (n Npm) DownloadComponent() ([]byte, error) {
	// Get NPM component
	req, err := http.NewRequest("GET", n.assetDownloadURL(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	resp, err := HttpClient().Do(req)
	if err != nil {
		return nil, err
	}

	// Check response for error
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: unable to download npm asset. sending '%s' request: status code %d %v",
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

func (n *Npm) PrepareDataToUpload(body []byte) (interface{}, error) {
	writer := multipart.NewWriter(n.Content.Data)
	part, err := writer.CreateFormFile("npm.asset", fmt.Sprintf("@%s", n.FileName))
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
	n.Content.Type = writer.FormDataContentType()
	return n, nil
}

func (n Npm) assetDownloadURL() string {
	return fmt.Sprintf("%s%s", n.Server, n.Path)
}
