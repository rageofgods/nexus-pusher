package comps

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
)

func (s *NexusServer) uploadComponent(format ComponentType, c *http.Client, asset *NexusExportComponentAsset,
	repoName string) error {
	// Create outer pipe to interconnect with inner pipe, to be able
	// to stream data directly from source component repo to target nexus repo.
	outerPipeReader, outerPipeWriter := io.Pipe()

	// Create multipart writer to connect it to the pipe and modify incoming
	// binary data from remote external repository (i.e PYPY, NPM, etc) on the fly
	multipartWriter := multipart.NewWriter(outerPipeWriter)

	// Creating error group for awaiting result from check repos types
	// errCtx will cancel http request if any errors was found
	errGroup, errCtx := errgroup.WithContext(context.Background())

	switch format {
	case NPM:
		// Download NPM component from official repo and return structured data
		npm := NewNpm(npmSrv, asset.Path, asset.FileName)

		// Start to download data and convert it to multipart stream
		prepareToUpload(errCtx, npm, multipartWriter, outerPipeWriter, errGroup)

		// Upload component to target nexus server
		errGroup.Go(func() error {
			if err := s.uploadComponentWithType(errCtx, repoName, c, asset, outerPipeReader, multipartWriter); err != nil {
				return err
			}
			return nil
		})

	case PYPI:
		// Download PYPI component from official repo and return structured data
		pypi := NewPypi(pypiSrv, asset.Path, asset.FileName, asset.Name, asset.Version)

		// Start to download data and convert it to multipart stream
		prepareToUpload(errCtx, pypi, multipartWriter, outerPipeWriter, errGroup)

		// Upload component to target nexus server
		errGroup.Go(func() error {
			if err := s.uploadComponentWithType(errCtx, repoName, c, asset, outerPipeReader, multipartWriter); err != nil {
				return err
			}
			return nil
		})

	}
	// If we found error, return it
	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}

// Download component following provided interface type
func prepareToUpload(ctx context.Context, t Typer, multipartWriter *multipart.Writer,
	outerPipeWriter *io.PipeWriter, errGroup *errgroup.Group) {
	// Create error errGroup to handle any errors
	innerPipeReader, innerPipeWriter := io.Pipe()

	// Start downloading component from remote repo
	errGroup.Go(func() error {
		if err := t.DownloadComponent(ctx, innerPipeWriter); err != nil {
			return err
		}
		return nil
	})

	// Convert downloaded component to multipart asset on the fly
	errGroup.Go(func() error {
		if err := t.PrepareDataToUpload(innerPipeReader, outerPipeWriter, multipartWriter); err != nil {
			return err
		}
		return nil
	})
}

func (s *NexusServer) uploadComponentWithType(ctx context.Context, repoName string, c *http.Client,
	asset *NexusExportComponentAsset, r *io.PipeReader, mw *multipart.Writer) error {
	// Upload component to nexus repo
	srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
		s.BaseUrl,
		s.ApiComponentsUrl,
		repoName)
	req, err := http.NewRequest("POST", srvUrl, r)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.SetBasicAuth(s.Username, s.Password)
	req = req.WithContext(ctx)

	// Start uploading component to remote nexus
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// Check server response
	if resp.StatusCode != http.StatusNoContent {
		log.Printf("error: unable to upload component %s to repository '%s' at server %s. Reason: %s",
			asset.Path,
			repoName,
			s.Host,
			resp.Status)
		return fmt.Errorf("error: unable to upload component %s to repository '%s' at server %s. Reason: %s",
			asset.Path,
			repoName,
			s.Host,
			resp.Status)
	} else {
		log.Printf("Component %s successfully uploaded to repository '%s' at server %s",
			asset.Path,
			repoName,
			s.Host)
	}

	if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
		return fmt.Errorf("%w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
