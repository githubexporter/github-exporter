package config

import "github.com/spf13/viper"

type ExporterConfig struct {
	MetricsPath     string
	ListenPort      string
	LogLevel        string
	ApplicationName string
}

func init() {
	viper.SetDefault("METRICS_PATH", "/metrics")
	viper.SetDefault("LISTEN_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("APP_NAME", "app")
}
