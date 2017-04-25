package exporter

import (
	"github.com/infinityworksltd/github-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	APIMetrics map[string]*prometheus.Desc
	config.Config
}

// APIResponse is used to store data from all the relevant endpoints in the API
type APIResponse []struct {
	Name  string `json:"name"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	Private    bool    `json:"private"`
	Forks      float64 `json:"forks"`
	Stars      float64 `json:"stargazers_count"`
	OpenIssues float64 `json:"open_issues"`
	Watchers   float64 `json:"subscribers_count"`
	Size       float64 `json:"size"`
}

type RateLimits struct {
	Limit     float64
	Remaining float64
	Reset     float64
}
