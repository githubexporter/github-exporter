package config

import "github.com/spf13/viper"

type QueryConfiguration struct {
	Stars      bool
	OpenIssues bool
	Watchers   bool
	Forks      bool
	Size       bool
}

func init() {
	viper.SetDefault("STARS", true)
	viper.SetDefault("ISSUES", true)
	viper.SetDefault("WATCHERS", true)
	viper.SetDefault("FORKS", true)
	viper.SetDefault("SIZE", true)
}
