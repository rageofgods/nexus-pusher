package config

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		&log.Logger{
			Out: os.Stdout,
			Formatter: &log.TextFormatter{
				FullTimestamp:   true,
				DisableQuote:    true,
				TimestampFormat: LogTimeFormat,
			},
			Level: logLevelFromEnv(),
		},
	}
}

func (l Logger) SetupLogger() {
	log.SetFormatter(l.Logger.Formatter)
	log.SetOutput(l.Logger.Out)
	log.SetLevel(l.Logger.Level)
}

// logLevelFromEnv read environment for log level
func logLevelFromEnv() log.Level {
	level := os.Getenv(envLogLevel)
	switch strings.ToLower(level) {
	case "debug":
		return log.DebugLevel
	case "warn":
		return log.WarnLevel
	case "info":
		return log.InfoLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

const (
	envLogLevel string = "NEXUS_PUSHER_LOG_LEVEL"
)
