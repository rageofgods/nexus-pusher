package logger

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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
	c.Warnln(msg, logMessage)
}

// Info mock
func (c *CustomRetryLogger) Info(_ string, _ ...interface{}) {
	// Do nothing to disable this type of logs
}

// Debug mock
func (c *CustomRetryLogger) Debug(_ string, _ ...interface{}) {
	// Do nothing to disable this type of logs
}

// Warn Mock
func (c *CustomRetryLogger) Warn(_ string, _ ...interface{}) {
	// Do nothing to disable this type of logs
}
