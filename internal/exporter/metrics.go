package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// GetMetrics - Adds all the metric definitions to a map of strings, returns the map.
func GetMetrics() map[string]*prometheus.Desc {

	APIMetrics := make(map[string]*prometheus.Desc)

	APIMetrics["Stars"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "stars"),
		"Total number of Stars for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	APIMetrics["OpenIssues"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "open_issues"),
		"Total number of open issues for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	APIMetrics["PullRequestCount"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "pull_request_count"),
		"Total number of pull requests for given repository",
		[]string{"repo", "user"}, nil,
	)
	APIMetrics["Watchers"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "watchers"),
		"Total number of watchers/subscribers for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	APIMetrics["Forks"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "forks"),
		"Total number of forks for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	APIMetrics["Size"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "size_kb"),
		"Size in KB for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	APIMetrics["ReleaseDownloads"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "release_downloads"),
		"Download count for a given release",
		[]string{"repo", "user", "release", "name", "tag", "created_at"}, nil,
	)
	APIMetrics["Limit"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "rate", "limit"),
		"Number of API queries allowed in a 60 minute window",
		[]string{}, nil,
	)
	APIMetrics["Remaining"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "rate", "remaining"),
		"Number of API queries remaining in the current window",
		[]string{}, nil,
	)
	APIMetrics["Reset"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "rate", "reset"),
		"The time at which the current rate limit window resets in UTC epoch seconds",
		[]string{}, nil,
	)

	return APIMetrics
}
