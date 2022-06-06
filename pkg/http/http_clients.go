package http

import (
	"github.com/hashicorp/go-retryablehttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"nexus-pusher/pkg/logger"
	"time"
)

// HttpRetryClient returns http client with optional timeout parameter
// Default timeout value is 10 seconds
func HttpRetryClient(seconds ...int) *http.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		MaxIdleConns:        100,
		DisableKeepAlives:   true,
	}

	customLogger := &logger.CustomRetryLogger{log.StandardLogger()}
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

// HttpClient returns http client with optional timeout parameter
// Default timeout value is 10 seconds
func HttpClient(seconds ...int) *http.Client {
	c := &http.Client{}
	c.Transport = &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
		MaxIdleConns:        100,
		DisableKeepAlives:   true,
	}

	if len(seconds) != 0 {
		c.Timeout = time.Duration(seconds[0]) * time.Second
	} else {
		c.Timeout = 10 * time.Second
	}

	return c
}
