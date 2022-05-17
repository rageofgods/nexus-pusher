package comps

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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

func (n Npm) DownloadComponent(ctx context.Context, innerPipeWriter *io.PipeWriter) error {
	// Get NPM component
	req, err := http.NewRequest("GET", n.assetDownloadURL(), nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "application/octet-stream")

	// Send request
	resp, err := HttpClient(120).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check response for error
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: unable to download npm asset. sending '%s' request: status code %d %v",
			resp.Request.Method,
			resp.StatusCode,
			resp.Request.URL)
	}

	// Read response body
	_, err = io.Copy(innerPipeWriter, resp.Body)
	if err != nil {
		return err
	}
	defer innerPipeWriter.Close()

	return nil
}

func (n *Npm) PrepareDataToUpload(innerPipeReader *io.PipeReader,
	outerPipeWriter *io.PipeWriter, multipartWriter *multipart.Writer) error {
	// Close writers at the end of call
	defer outerPipeWriter.Close()
	defer multipartWriter.Close()

	// Create multipart asset
	part, err := multipartWriter.CreateFormFile("npm.asset", fmt.Sprintf("@%s", n.FileName))
	if err != nil {
		return err
	}

	// Convert downloaded data to multipart
	if _, err := io.Copy(part, innerPipeReader); err != nil {
		return err
	}

	return nil
}

func (n Npm) assetDownloadURL() string {
	return fmt.Sprintf("%s%s", n.Server, n.Path)
}
