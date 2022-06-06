package client

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"nexus-pusher/internal/config"
	"nexus-pusher/internal/core"
	"nexus-pusher/internal/server"
	"nexus-pusher/pkg/http_clients"
	"nexus-pusher/pkg/utils"
	"time"
)

// pushClient is used to push diff to server-side
type pushClient struct {
	serverAddress string
	serverUser    string
	serverPass    string
	cookie        *http.Cookie
	metrics       *nexusClientMetrics
}

func newPushClient(
	serverAddress string,
	serverUser string,
	serverPass string,
	metrics *nexusClientMetrics) *pushClient {
	return &pushClient{serverAddress: serverAddress, serverUser: serverUser, serverPass: serverPass, metrics: metrics}
}

// authorize the client with server using plain type credentials from configuration file
func (p *pushClient) authorize() error {
	requestUrl := fmt.Sprintf("%s%s%s", p.serverAddress, config.URIBase, config.URILogin)
	client := http_clients.HttpRetryClient()
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return fmt.Errorf("authorize: %w", err)
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.SetBasicAuth(p.serverUser, p.serverPass)
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("authorize: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &utils.ContextError{
			Context: "authorize",
			Err:     fmt.Errorf("%s responded with status: %d", p.serverAddress, resp.StatusCode),
		}
	}
	// Finding JWT Cookie in response
	for _, v := range resp.Cookies() {
		if v.Name == config.JWTCookieName {
			// If Cookie was found when set pointer to it
			p.cookie = v
			return nil
		}
	}
	// If we can't find Cookie in the server response, return error
	return &utils.ContextError{
		Context: "authorize",
		Err:     fmt.Errorf("unable to find JWT Cookie in "),
	}
}

// refreshAuth will refresh JWT token for client through server request
func (p *pushClient) refreshAuth() error {
	// If JWT will be expired soon, try to refresh it
	if time.Until(p.cookie.Expires) < config.JWTTokenRefreshWindow*time.Second {
		requestUrl := fmt.Sprintf("%s%s%s", p.serverAddress, config.URIBase, config.URIRefresh)
		// Setup http client
		client := http_clients.HttpRetryClient()
		// Make new request
		req, err := http.NewRequest("GET", requestUrl, nil)
		if err != nil {
			return fmt.Errorf("refreshAuth: %w", err)
		}
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		// Append JWT auth Cookie
		req.AddCookie(p.cookie)

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("refreshAuth: %w", err)
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return &utils.ContextError{
				Context: "refreshAuth",
				Err: fmt.Errorf("error: unable to refresh JWT auth token. Server responded with status: %s",
					resp.Status),
			}
		}

		// Finding JWT Cookie in response
		for _, v := range resp.Cookies() {
			if v.Name == config.JWTCookieName {
				// If Cookie was found when set pointer to it
				p.cookie = v
				log.Debugf("Successfully refreshed JWT auth token. Continue polling...")
				return nil
			}
		}
	}
	return nil
}

// sendComparedRequest sends diff data to server
func (p *pushClient) sendComparedRequest(data *core.NexusExportComponents, repoName string) ([]byte, error) {
	var buf bytes.Buffer
	requestUrl := fmt.Sprintf("%s%s%s?repository=%s",
		p.serverAddress,
		config.URIBase,
		config.URIComponents,
		repoName)
	// Setup http client
	client := http_clients.HttpRetryClient()
	// Encode data to buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, fmt.Errorf("sendComparedRequest: %w", err)
	}
	req, err := http.NewRequest("POST", requestUrl, &buf)
	if err != nil {
		return nil, fmt.Errorf("sendComparedRequest: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	// Append JWT auth Cookie
	req.AddCookie(p.cookie)

	// Send request
	log.Debugf("Sending components diff to %s server...", p.serverAddress)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sendComparedRequest: %w", err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, &utils.ContextError{
			Context: "sendComparedRequest",
			Err:     fmt.Errorf("error: %s responded with status: %s", p.serverAddress, resp.Status),
		}
	}
	log.Debugf("Sending components diff to %s successfully complete.", p.serverAddress)

	// Read all body data
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sendComparedRequest: %w", err)
	}

	return body, nil
}

// pollComparedResults long-http polling function to get upload results from server
func (p *pushClient) pollComparedResults(body []byte, dstRepo string, dstServer string) error {
	// Convert body to Message type
	msg := &server.Message{}
	if err := json.Unmarshal(body, msg); err != nil {
		return fmt.Errorf("pollComparedResults: %w", err)
	}

	log.WithFields(
		log.Fields{"id": msg.ID},
	).Infof("Start polling results for destination repo '%s' at server '%s'", dstRepo, dstServer)
	// Queue http polling
	requestUrl := fmt.Sprintf("%s%s%s?uuid=%s",
		p.serverAddress,
		config.URIBase,
		config.URIComponents,
		msg.ID)

	// Setup http client
	client := http_clients.HttpRetryClient()

	// Poll maximum for 3600 seconds (60 min)
	limitTime := 3600
	for x := 1; x < limitTime; x++ {
		// Setup new Request
		req, err := http.NewRequest("GET", requestUrl, nil)
		if err != nil {
			return fmt.Errorf("pollComparedResults: %w", err)
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		// Append JWT auth Cookie
		req.AddCookie(p.cookie)

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("pollComparedResults: %w", err)
		}

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return &utils.ContextError{
				Context: "pollComparedResults",
				Err: fmt.Errorf("error: %s responded with status: %s",
					p.serverAddress,
					resp.Status),
			}
		}

		// Read all body data
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("pollComparedResults: %w", err)
		}
		// Close response body
		if err := resp.Body.Close(); err != nil {
			return fmt.Errorf("pollComparedResults: %w", err)
		}

		// Convert body to Message type
		if err := json.Unmarshal(body, msg); err != nil {
			return fmt.Errorf("pollComparedResults: %w", err)
		}

		// If server respond with 'complete' message stop polling
		if msg.Complete {
			log.WithFields(
				log.Fields{
					"id":     msg.ID,
					"errors": len(msg.Response),
				},
			).Infof("Polling complete for destinantion repo '%s' at server '%s'",
				dstRepo, dstServer)

			// Update metric for errors count if any
			if len(msg.Response) > 0 {
				p.metrics.SyncErrorsCountByLabels(dstServer, dstRepo, msg.ID.String()).Set(float64(len(msg.Response)))
			}

			// log all response errors
			for _, m := range msg.Response {
				log.WithFields(
					log.Fields{"id": msg.ID},
				).Warnf("%s", m)
			}
			return nil
		}
		// Report server polling status every 30 seconds
		if x%30 == 0 {
			log.WithFields(
				log.Fields{"id": msg.ID}).Debugf("Server polling in progress... %d seconds passed", x)
		}
		// Try to refresh auth token
		if err := p.refreshAuth(); err != nil {
			return fmt.Errorf("pollComparedResults: %w", err)
		}
		// Limit server requests to 1 RPS
		time.Sleep(1 * time.Second)
	}
	// Show error if we don't get results in time
	return &utils.ContextError{
		Context: "pollComparedResults",
		Err: fmt.Errorf("unable to get results from for message id %s in %d seconds",
			msg.ID,
			limitTime),
	}
}
