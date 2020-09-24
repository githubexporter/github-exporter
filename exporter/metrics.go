package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AddMetrics - Add's all of the metrics to a map of strings, returns the map.
func AddMetrics() map[string]*prometheus.Desc {

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
	APIMetrics["PullRequests"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "pull_request_count"),
		"Total number of pull requests for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	APIMetrics["Commits"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "commit_count"),
		"Total number of commits for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
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
		[]string{"repo", "user", "release", "name", "created_at"}, nil,
	)
	APIMetrics["Releases"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "releases"),
		"Number of releases for a repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
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

func (e *Exporter) derefString(s *string) string {

	if s != nil {
		return *s
	}

	return ""
}

func (e *Exporter) derefBool(b *bool) bool {

	if b != nil {
		return *b
	}

	e.Log.Info("Bool nil, defaulting to false")

	return false
}

func (e *Exporter) derefInt(i *int) int {

	if i != nil {
		return *i
	}

	e.Log.Info("Int nil, defaulting to 0")

	return 0

}

// processMetrics - processes the response data and sets the metrics using it as a source
func (e *Exporter) processMetrics(ch chan<- prometheus.Metric) error {

	// Range through Repository metrics
	for _, x := range e.Repositories {
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Stars"], prometheus.GaugeValue, x.StargazersCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Forks"], prometheus.GaugeValue, x.ForksCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Watchers"], prometheus.GaugeValue, x.WatchersCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Size"], prometheus.GaugeValue, x.Size, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["PullRequests"], prometheus.GaugeValue, x.PullsCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["OpenIssues"], prometheus.GaugeValue, x.OpenIssuesCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Releases"], prometheus.GaugeValue, x.Releases, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Commits"], prometheus.GaugeValue, x.CommitsCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)

	}

	// Set Rate limit stats
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Limit"], prometheus.GaugeValue, e.RateLimits.Limit)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Remaining"], prometheus.GaugeValue, e.RateLimits.Remaining)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Reset"], prometheus.GaugeValue, e.RateLimits.Reset)

	return nil
}
