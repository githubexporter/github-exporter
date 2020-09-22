package exporter

import (
	"github.com/google/go-github/github"
	"github.com/infinityworks/github-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	APIMetrics    map[string]*prometheus.Desc
	Config        config.Config
	Repositories  []*RepositoryMetrics
	Organisations []*OrganisationMetrics
}

type RepositoryMetrics struct {
	Base       github.Repository
	PullsCount float64
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
