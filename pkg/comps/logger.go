package comps

import (
	"fmt"
	"log"
	"net/url"
)

type CustomRetryLogger struct {
	*log.Logger
}

// Error Return format error message
func (c *CustomRetryLogger) Error(msg string, keysAndValues ...interface{}) {
	var logMessage string
	for _, v := range keysAndValues {
		switch t := v.(type) {
		case *url.Error:
			logMessage = fmt.Sprintf("%s - %s", t.Err.Error(), t.URL)
		case *url.URL:
		}
	}
	c.Println(msg, logMessage)
}

// Info mock
func (c *CustomRetryLogger) Info(msg string, keysAndValues ...interface{}) {
	return
}

// Debug mock
func (c *CustomRetryLogger) Debug(msg string, keysAndValues ...interface{}) {
	return
}

// Warn Mock
func (c *CustomRetryLogger) Warn(msg string, keysAndValues ...interface{}) {
	return
}
