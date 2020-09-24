package exporter

import (
	"github.com/infinityworks/github-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	APIMetrics   map[string]*prometheus.Desc
	Config       config.Config
	Repositories []RepositoryMetrics
	RateLimits   RateMetrics
}

type RepositoryMetrics struct {
	Name            string
	Owner           string
	Archived        string
	Private         string
	Fork            string
	ForksCount      float64
	WatchersCount   float64
	StargazersCount float64
	PullsCount      float64
	CommitsCount    float64
	OpenIssuesCount float64
	Size            float64
	Releases        float64
	License         string
	Language        string
}

type RateMetrics struct {
	Limit     float64
	Remaining float64
	Reset     float64
}

type OrganisationMetrics struct {
	TotalMemberCount   float64
	ActiveMemberCount  float64
	PendingMemberCount float64
}
