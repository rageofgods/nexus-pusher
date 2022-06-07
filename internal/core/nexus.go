package core

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"nexus-pusher/internal/config"
	"nexus-pusher/pkg/utils"
)

func (s *NexusServer) GetComponents(ctx context.Context, c *http.Client, ncs []*NexusComponent,
	repoName string) ([]*NexusComponent, error) {
	srvUrl := fmt.Sprintf("%s%s%s?repository=%s", s.Host,
		s.BaseUrl,
		s.ApiComponentsUrl,
		repoName)

	body, err := s.SendRequest(srvUrl, "GET", c, nil)
	if err != nil {
		return nil, err
	}

	var nc NexusComponents
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(body, &nc); err != nil {
		return nil, err
	}
	ncs = append(ncs, nc.Items...)

	if nc.ContinuationToken != "" {
		srvUrl = fmt.Sprintf("%s%s%s?repository=%s", s.Host,
			s.BaseUrl,
			s.ApiComponentsUrl,
			repoName)

		for {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("GetComponents: canceling processing repo '%s' because of upstream error",
					repoName)
			default:
				body, err := s.SendRequest(srvUrl, "GET", c, nil)
				if err != nil {
					return nil, err
				}

				if err := json.Unmarshal(body, &nc); err != nil {
					return nil, err
				}
				ncs = append(ncs, nc.Items...)

				// Send log message every 500 new components
				if len(ncs) <= 10 || len(ncs)%500 == 0 {
					log.Debugf("Analyzing repo '%s' at server '%s', please wait... Processed %d assets.",
						repoName,
						s.Host,
						len(ncs))
				}

				if nc.ContinuationToken == "" {
					break
				}
			}
		}
	}
	return ncs, nil
}

// UploadComponents is used to upload nexus artifacts following by 'nec' list
func (s *NexusServer) UploadComponents(nec *NexusExportComponents, repoName string, cs *config.Server) []UploadResult {

	limitChan := make(chan struct{}, cs.Concurrency)
	resultsChan := make(chan *UploadResult)

	defer func() {
		close(limitChan)
		close(resultsChan)
	}()

	var resultsCounter int
	for _, v := range nec.Items {
		if config.ComponentType(v.Format).Bundled() {
			// Process assets as a bundle
			resultsCounter++
			go func(format config.ComponentType, component *NexusExportComponent, repoName string) {
				limitChan <- struct{}{}
				result := &UploadResult{}
				if err := s.uploadComponent(format, component, repoName); err != nil {
					log.Errorf("%v", err)
					result = &UploadResult{Err: err, ComponentPath: component.FullName()}
				}
				resultsChan <- result
				<-limitChan
			}(config.ComponentType(v.Format), v, repoName)
		} else {
			// Process assets individually
			for _, vv := range v.Assets {
				resultsCounter++
				go func(format config.ComponentType, asset *NexusExportComponentAsset, repoName string, src string) {
					limitChan <- struct{}{}
					result := &UploadResult{}
					if err := s.uploadAsset(format, asset, repoName, src); err != nil {
						log.Errorf("%v", err)
						result = &UploadResult{Err: err, ComponentPath: asset.Path}
					}
					resultsChan <- result
					<-limitChan
				}(config.ComponentType(v.Format), vv, repoName, v.ArtifactsSource)
			}
		}
	}
	var results []UploadResult
	for {
		result := <-resultsChan
		results = append(results, *result)

		// if we've reached the expected amount of results then stop
		if len(results) == resultsCounter {
			break
		}
	}
	return results
}

func (s *NexusServer) SendRequest(srvUrl string, method string, c *http.Client, b io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, srvUrl, b)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.Username, s.Password)
	// Send request
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &utils.ContextError{
			Context: "SendRequest",
			Err: fmt.Errorf("error: sending '%s' request: status code %d %v",
				resp.Request.Method,
				resp.StatusCode,
				resp.Request.URL),
		}
	}
	// Read all body data
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("SendRequest: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("SendRequest: %w", err)
	}
	return body, nil
}
