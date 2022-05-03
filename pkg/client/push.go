package client

import (
	"bytes"
	"fmt"
	"github.com/goccy/go-json"
	"io/ioutil"
	"log"
	"net/http"
	"nexus-pusher/pkg/comps"
	"nexus-pusher/pkg/config"
	"nexus-pusher/pkg/server"
	"time"
)

// pushClient is used to push diff to server-side
type pushClient struct {
	serverAddress string
	serverUser    string
	serverPass    string
	cookie        *http.Cookie
}

func newPushClient(serverAddress string, serverUser string, serverPass string) *pushClient {
	return &pushClient{serverAddress: serverAddress, serverUser: serverUser, serverPass: serverPass}
}

// authorize the client with server using plain type credentials from configuration file
func (p *pushClient) authorize() error {
	requestUrl := fmt.Sprintf("%s%s%s", p.serverAddress, config.URIBase, config.URILogin)
	client := comps.HttpClient()
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.SetBasicAuth(p.serverUser, p.serverPass)
	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: %s responded with status: %d", p.serverAddress, resp.StatusCode)
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
	return fmt.Errorf("error: unable to find JWT Cookie in ")
}

// refreshAuth will refresh JWT token for client through server request
func (p *pushClient) refreshAuth() error {
	// If JWT will be expired soon, try to refresh it
	if time.Until(p.cookie.Expires) < config.JWTTokenRefreshWindow*time.Second {
		requestUrl := fmt.Sprintf("%s%s%s", p.serverAddress, config.URIBase, config.URIRefresh)
		// Setup http client
		client := comps.HttpClient()
		// Make new request
		req, err := http.NewRequest("GET", requestUrl, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		// Append JWT auth Cookie
		req.AddCookie(p.cookie)

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error: unable to refresh JWT auth token. Server responded with status: %s",
				resp.Status)
		}

		// Finding JWT Cookie in response
		for _, v := range resp.Cookies() {
			if v.Name == config.JWTCookieName {
				// If Cookie was found when set pointer to it
				p.cookie = v
				log.Printf("Succesfully refreshed JWT auth token. Continue polling...")
				return nil
			}
		}
	}
	return nil
}

// sendComparedRequest sends diff data to server
func (p *pushClient) sendComparedRequest(data *comps.NexusExportComponents, repoName string) ([]byte, error) {
	var buf bytes.Buffer
	requestUrl := fmt.Sprintf("%s%s%s?repository=%s",
		p.serverAddress,
		config.URIBase,
		config.URIComponents,
		repoName)
	// Setup http client
	client := comps.HttpClient()
	// Encode data to buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	req, err := http.NewRequest("POST", requestUrl, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	// Append JWT auth Cookie
	req.AddCookie(p.cookie)

	// Send request
	log.Printf("Sending components diff to %s server...", p.serverAddress)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: %s responded with status: %s", p.serverAddress, resp.Status)
	}
	log.Printf("Sending components diff to %s succesfully complete.", p.serverAddress)

	// Read all body data
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// pollComparedResults long-http polling function to get upload results from server
func (p *pushClient) pollComparedResults(body []byte) error {
	// Convert body to Message type
	msg := &server.Message{}
	if err := json.Unmarshal(body, msg); err != nil {
		return err
	}

	log.Printf("Starting server polling for message id %s to get upload results...", msg.ID)
	// Queue http polling
	requestUrl := fmt.Sprintf("%s%s%s?uuid=%s",
		p.serverAddress,
		config.URIBase,
		config.URIComponents,
		msg.ID)

	// Setup http client
	client := comps.HttpClient()

	// Poll maximum for 1800 seconds (30 min)
	limitTime := 1800
	for x := 1; x < limitTime; x++ {
		// Setup new Request
		req, err := http.NewRequest("GET", requestUrl, nil)
		if err != nil {
			return err
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		// Append JWT auth Cookie
		req.AddCookie(p.cookie)

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		// Check server response
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("error: %s responded with status: %s",
				p.serverAddress,
				resp.Status)
		}

		// Read all body data
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		// Close response body
		if err := resp.Body.Close(); err != nil {
			return err
		}

		// Convert body to Message type
		if err := json.Unmarshal(body, msg); err != nil {
			return err
		}

		// If server respond with 'complete' message stop polling
		if msg.Complete {
			log.Printf("Server polling for message id %s is complete with response from server:\n>>>\n%v\n<<<",
				msg.ID,
				msg.Response)
			return nil
		}
		// Report server polling status every 30 seconds
		if x%30 == 0 {
			log.Printf("Server polling for message id %s in progress... %d seconds passed",
				msg.ID,
				x)
		}
		// Try to refresh auth token
		if err := p.refreshAuth(); err != nil {
			return err
		}
		// Limit server requests to 1 RPS
		time.Sleep(1 * time.Second)
	}
	// Show error if we don't get results in time
	return fmt.Errorf("error: unable to get results from for message id %s in %d seconds",
		msg.ID,
		limitTime)
}
