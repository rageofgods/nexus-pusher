package types

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
)

type Npm struct {
	Server  string
	Path    string
	Content *NpmData
}

type NpmData struct {
	Type string
	Data *bytes.Buffer
}

func NewNpm(server string, path string) *Npm {
	return &Npm{Server: server, Path: path, Content: &NpmData{Data: &bytes.Buffer{}}}
}

func (n Npm) DownloadComponent() ([]byte, error) {
	// Get NPM component
	componentPath := fmt.Sprintf("%s%s", n.Server, n.Path)
	resp, err := http.Get(componentPath)
	if err != nil {
		return nil, err
	}

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm download bad status: %s", resp.Status)
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
	part, err := writer.CreateFormFile("npm.asset", fmt.Sprintf("@%s", n.componentNameFromPath()))
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

func (n Npm) componentNameFromPath() string {
	cmpPathSplit := strings.Split(n.Path, "/")
	return cmpPathSplit[len(cmpPathSplit)-1]
}
