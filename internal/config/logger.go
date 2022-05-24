package config

import (
	log "github.com/sirupsen/logrus"
	"os"
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
			Level: log.InfoLevel,
		},
	}
}

func (l Logger) SetupLogger() {
	log.SetFormatter(l.Logger.Formatter)
	log.SetOutput(l.Logger.Out)
	log.SetLevel(l.Logger.Level)
}
