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
	"path/filepath"
	"strings"
)

func (s *NexusServer) GetComponents(ctx context.Context, c *http.Client, repoName string) ([]*NexusComponent, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var ncs []*NexusComponent
	var continuationToken string
Outer:
	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("GetComponents: canceling processing repo '%s' because of upstream error",
				repoName)
		default:
			var srvUrl string
			if i == 0 {
				// First iteration is always without continuationToken
				srvUrl = fmt.Sprintf("%s%s%s?repository=%s",
					s.Host,
					s.BaseUrl,
					s.ApiComponentsUrl,
					repoName)
			} else {
				srvUrl = fmt.Sprintf("%s%s%s?repository=%s&continuationToken=%s",
					s.Host,
					s.BaseUrl,
					s.ApiComponentsUrl,
					repoName,
					continuationToken)
			}

			body, err := s.SendRequest(srvUrl, "GET", c, nil)
			if err != nil {
				return nil, err
			}

			var nc NexusComponents
			if err := json.Unmarshal(body, &nc); err != nil {
				return nil, err
			}

			// Filter assets for hash artifacts
			filterHashAssets(&nc)

			continuationToken = nc.ContinuationToken
			ncs = append(ncs, nc.Items...)

			// Send log message every 500 new components
			if len(ncs) <= 10 || len(ncs)%500 == 0 {
				log.Debugf("Analyzing repo '%s' at server '%s', please wait... Processed %d assets.",
					repoName,
					s.Host,
					len(ncs))
			}

			// if len(ncs) > 100 {
			//	return ncs, nil
			// }

			if continuationToken == "" {
				break Outer
			}
		}
	}
	return ncs, nil
}

// filterHashAssets filter out assets for hash type artifacts
func filterHashAssets(nc *NexusComponents) {
	for compInd := 0; compInd < len(nc.Items); compInd++ {
		for assetInd := 0; assetInd < len(nc.Items[compInd].Assets); assetInd++ {
			switch nc.Items[compInd].Assets[assetInd].Format {
			case config.MAVEN2.String():
				mavenFilteredExtensions := map[string]struct{}{
					"sha1":   {},
					"md5":    {},
					"sha256": {},
					"sha512": {},
				}
				// Get file extension from path and
				// remove leading dot from it
				fileExtension := filepath.Ext(nc.Items[compInd].Assets[assetInd].Path)[1:]
				// Check file extension to match filter list and
				// remove slice asset following current index
				if _, ok := mavenFilteredExtensions[fileExtension]; ok {
					log.Debugf("Maven2: filtering '%s' asset from comparison "+
						"by extension '%s' list", nc.Items[compInd].Assets[assetInd].Path, fileExtension)

					nc.Items[compInd].Assets = append(nc.Items[compInd].Assets[:assetInd],
						nc.Items[compInd].Assets[assetInd+1:]...)
					// If an asset was filtered - get back to one index position
					assetInd--
				}
			case config.NUGET.String():
				// Remove '+...' postfix from component version and asset path
				// to be able to compare assets with the same names\versions
				// Because nexus will add '+sha.' postfix for some assets
				// Following its own API upload rules
				splitAssetVersion := strings.Split(nc.Items[compInd].Version, "+")
				if len(splitAssetVersion) == 2 {
					log.Debugf("Nuget: removing '%s' suffix from asset version: '%s'",
						splitAssetVersion[1], nc.Items[compInd].Version)

					nc.Items[compInd].Version = splitAssetVersion[0]
				}

				splitAssetPath := strings.Split(filepath.Base(nc.Items[compInd].Assets[assetInd].Path), "+")
				if len(splitAssetPath) == 2 {
					log.Debugf("Nuget: removing '%s' suffix from asset path: '%s'",
						splitAssetPath[1], nc.Items[compInd].Assets[assetInd].Path)

					nc.Items[compInd].Assets[assetInd].Path = fmt.Sprintf("%s/%s",
						filepath.Dir(nc.Items[compInd].Assets[assetInd].Path), splitAssetPath[0],
					)
				}
			}
		}
	}
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
