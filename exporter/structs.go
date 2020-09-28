package exporter

import (
	"github.com/infinityworks/github-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	Metrics        map[string]*prometheus.Desc
	Config         config.Config
	Log            *logrus.Logger
	Repositories   []RepositoryMetrics
	ProcessedRepos []ProcessedRepos
	Organisations  []OrganisationMetrics
	RateLimits     RateMetrics
}

// RepositoryMetrics defines our repository metric footprint
// Similar to the standard github library but value based and not pointers
// Also this includes the additional metrics we capture outside the standard return
// from the github API
type RepositoryMetrics struct {
	Name            string
	Owner           string
	Archived        string
	Private         string
	Fork            string
	ForksCount      float64
	WatchersCount   float64
	StargazersCount float64
	OpenIssuesCount float64
	Size            float64
	License         string
	Language        string
	EnhancedMetrics EnhancedMetrics
}

// EnhancedMetrics is used to track metrics not available through
// standard api endpoints in the v3 API. Usage of these endpoints is though
// very expensive, increasing the API calls used significantly
type EnhancedMetrics struct {
	PullsCount   float64
	CommitsCount float64
	Releases     float64
}

// RateMetrics help us monitor performance against the
// GitHub API Rate limits imposed
type RateMetrics struct {
	Limit     float64
	Remaining float64
	Reset     float64
}

// OrganisationMetrics helps us capture metrics specific to our organisations
// We simply miss when focussing in on repositories alone
type OrganisationMetrics struct {
	Name                       string
	MembersCount               float64
	OutsideCollaboratorsCount  float64
	PendingOrgInvitationsCount float64
}

// ProcessedRepos is used to help track and avoid duplicate repository collection of metrics
type ProcessedRepos struct {
	Owner string
	Name  string
}
