package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func LogLevel() log.Level {
	switch viper.GetString("LOG_LEVEL") {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "fatal":
		return log.FatalLevel
	case "panic":
		return log.PanicLevel
	default:
		return log.DebugLevel
	}
}
