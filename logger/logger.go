package logger

import (
	"sync"

	"github.com/benri-io/jira-exporter/config"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var loggerMap = MakeLoggerMap()

type LoggerMap struct {
	mux     *sync.Mutex
	loggers map[string]*log.Logger
}

func MakeLoggerMap() (l *LoggerMap) {
	return &LoggerMap{
		&sync.Mutex{},
		make(map[string]*log.Logger),
	}
}

func Start(config config.AppConfig) (logger *log.Logger) {
	logger = log.New()
	//	logger.Formatter = &log.JSONFormatter{
	//DisableColors: false,
	//FullTimestamp: false,
	//}
	setLogLevel(config.LogLevel(), logger)
	return logger
}

func SetLogger(key string, logger *log.Logger) {
	loggerMap.mux.Lock()
	defer loggerMap.mux.Unlock()
	loggerMap.loggers[key] = logger
}

func SetDefaultLogger(logger *log.Logger) {
	loggerMap.mux.Lock()
	defer loggerMap.mux.Unlock()
	loggerMap.loggers["default"] = logger
}

func GetDefaultLogger() *log.Logger {
	loggerMap.mux.Lock()
	defer loggerMap.mux.Unlock()
	if v, ok := loggerMap.loggers["default"]; ok {
		return v
	}
	return logrus.New()
}

// setLogLevel - Sets the log level based on the passed argument.
func setLogLevel(level string, l *log.Logger) {
	switch level {
	case "debug":
		l.Level = log.DebugLevel
		break
	case "info":
		l.Level = log.InfoLevel
		break
	case "warn":
		l.Level = log.WarnLevel
		break
	case "fatal":
		l.Level = log.FatalLevel
		break
	case "panic":
		l.Level = log.PanicLevel
		break
	}
}
