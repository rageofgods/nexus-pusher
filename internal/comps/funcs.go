package comps

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/hashicorp/go-retryablehttp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"nexus-pusher/internal/config"
	"os"
	"runtime/debug"
	"time"
)

func (s *NexusServer) GetComponents(
	ctx context.Context,
	c *http.Client,
	ncs []*NexusComponent,
	repoName string,
	contToken ...string) ([]*NexusComponent, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("error: canceling processing repo '%s' because of upstream error",
			repoName)
	default:
		// Do nothing (continue execution)
	}

	var srvUrl string
	if len(contToken) != 0 {
		srvUrl = fmt.Sprintf("%s%s%s?repository=%s&continuationToken=%s", s.Host,
			s.BaseUrl,
			s.ApiComponentsUrl,
			repoName,
			contToken[0])
	} else {
		srvUrl = fmt.Sprintf("%s%s%s?repository=%s", s.Host,
			s.BaseUrl,
			s.ApiComponentsUrl,
			repoName)
	}

	body, err := s.SendRequest(srvUrl, "GET", c, nil)
	if err != nil {
		return nil, err
	}

	var nc NexusComponents
	if err := json.Unmarshal(body, &nc); err != nil {
		return nil, err
	}
	ncs = append(ncs, nc.Items...)

	// Send log message every 500 new components
	if len(ncs) <= 10 || len(ncs)%500 == 0 {
		log.Printf("Analyzing repo '%s' at server '%s', please wait... Processed %d assets.\n",
			repoName,
			s.Host,
			len(ncs))
	}

	//if len(ncs) > 1000 {
	//	return ncs, nil
	//}

	// Iterating over all API pages
	if nc.ContinuationToken != "" {
		ncs, err = s.GetComponents(ctx, c, ncs, repoName, nc.ContinuationToken)
		if err != nil {
			return nil, err
		}
	}

	return ncs, nil
}

// UploadComponents is used to upload nexus artifacts following by 'nec' list
func (s *NexusServer) UploadComponents(c *http.Client,
	nec *NexusExportComponents,
	repoName string,
	cs *config.Server) []UploadResult {

	limitChan := make(chan struct{}, cs.Concurrency)
	resultsChan := make(chan *UploadResult)

	defer func() {
		close(limitChan)
		close(resultsChan)
	}()

	var resultsCounter int
	for _, v := range nec.Items {
		for _, vv := range v.Assets {
			resultsCounter++
			go func(format ComponentType, c *http.Client, asset *NexusExportComponentAsset, repoName string) {
				limitChan <- struct{}{}
				result := &UploadResult{}
				if err := s.uploadComponent(format, c, asset, repoName); err != nil {
					log.Printf("%v", err)
					result = &UploadResult{Err: err, ComponentPath: asset.Path}
				}
				resultsChan <- result
				<-limitChan
				debug.FreeOSMemory()
			}(ComponentType(v.Format), c, vv, repoName)
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

// HttpClient returns http client with optional timeout parameter
// Default timeout value is 10 seconds
func HttpClient(seconds ...int) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		MaxIdleConns:        100,
	}
	customLogger := &CustomRetryLogger{log.New(os.Stdout, "", log.Ldate|log.Ltime)}
	retryClient.Logger = customLogger
	retryClient.RetryMax = 3
	if len(seconds) != 0 {
		retryClient.HTTPClient.Timeout = time.Duration(seconds[0]) * time.Second
	} else {
		retryClient.HTTPClient.Timeout = 10 * time.Second
	}
	client := retryClient.StandardClient()

	return client
}

func (s *NexusServer) SendRequest(srvUrl string, method string, c *http.Client, b io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, srvUrl, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.Username, s.Password)
	// Send request
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: sending '%s' request: status code %d %v",
			resp.Request.Method,
			resp.StatusCode,
			resp.Request.URL)
	}
	// Read all body data
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}
