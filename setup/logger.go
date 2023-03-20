package setup

import (
	log "github.com/sirupsen/logrus"
)

func ConfigureLogger(config LogConfig) *log.Logger {
	lvl, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		lvl = log.InfoLevel
	}
	log.SetLevel(lvl)
	return log.StandardLogger()
}
