package config

import "github.com/spf13/viper"

type APIConfig struct {
	APIURL string
}

func init() {
	viper.SetDefault("API_URL", "https://api.github.com")
}
