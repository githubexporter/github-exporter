package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AddMetrics - Add's all of the metrics to a map of strings, returns the map.
func AddMetrics() map[string]*prometheus.Desc {

	Metrics := make(map[string]*prometheus.Desc)

	Metrics["Stars"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "stars"),
		"Total number of Stars for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["OpenIssues"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "open_issues"),
		"Total number of open issues for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["PullRequests"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "pull_request_count"),
		"Total number of pull requests for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["Commits"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "commit_count"),
		"Total number of commits for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["Watchers"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "watchers"),
		"Total number of watchers/subscribers for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["Forks"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "forks"),
		"Total number of forks for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["Size"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "size_kb"),
		"Size in KB for given repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["ReleaseDownloads"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "release_downloads"),
		"Download count for a given release",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language", "release", "name", "created_at"}, nil,
	)
	Metrics["Releases"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "releases"),
		"Number of releases for a repository",
		[]string{"repo", "user", "private", "fork", "archived", "license", "language"}, nil,
	)
	Metrics["MembersCount"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "org", "members"),
		"Number of members in an organisation",
		[]string{"organisation"}, nil,
	)
	Metrics["OutsideCollaboratorsCount"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "org", "collaborators"),
		"Number outside collaborators in an organisation",
		[]string{"organisation"}, nil,
	)
	Metrics["PendingOrgInvitationsCount"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "org", "pending_invitations"),
		"Number of pending invitations",
		[]string{"organisation"}, nil,
	)
	Metrics["Limit"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "rate", "limit"),
		"Number of API queries allowed in a 60 minute window",
		[]string{}, nil,
	)
	Metrics["Remaining"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "rate", "remaining"),
		"Number of API queries remaining in the current window",
		[]string{}, nil,
	)
	Metrics["Reset"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "rate", "reset"),
		"The time at which the current rate limit window resets in UTC epoch seconds",
		[]string{}, nil,
	)

	return Metrics
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
func (e *Exporter) processMetrics(ch chan<- prometheus.Metric) {

	// Range through Repository metrics
	for _, x := range e.Repositories {
		ch <- prometheus.MustNewConstMetric(e.Metrics["Stars"], prometheus.GaugeValue, x.StargazersCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["Forks"], prometheus.GaugeValue, x.ForksCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["Watchers"], prometheus.GaugeValue, x.WatchersCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["Size"], prometheus.GaugeValue, x.Size, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["OpenIssues"], prometheus.GaugeValue, x.OpenIssuesCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)

		if e.optionalMetricEnabled("pulls") {
			ch <- prometheus.MustNewConstMetric(e.Metrics["PullRequests"], prometheus.GaugeValue, x.OptionalRepositoryMetrics.PullsCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		}
		if e.optionalMetricEnabled("releases") {
			ch <- prometheus.MustNewConstMetric(e.Metrics["Releases"], prometheus.GaugeValue, x.OptionalRepositoryMetrics.Releases, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)

			for _, y := range x.OptionalRepositoryMetrics.ReleaseDownloads {
				ch <- prometheus.MustNewConstMetric(e.Metrics["ReleaseDownloads"], prometheus.GaugeValue, y.DownloadCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language, y.ReleaseName, y.AssetName, y.CreatedAt)
			}
		}
		if e.optionalMetricEnabled("commits") {
			ch <- prometheus.MustNewConstMetric(e.Metrics["Commits"], prometheus.GaugeValue, x.OptionalRepositoryMetrics.CommitsCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		}
	}

	if len(e.Organisations) > 0 {
		for _, y := range e.Organisations {
			ch <- prometheus.MustNewConstMetric(e.Metrics["MembersCount"], prometheus.GaugeValue, y.MembersCount, y.Name)
			ch <- prometheus.MustNewConstMetric(e.Metrics["OutsideCollaboratorsCount"], prometheus.GaugeValue, y.OutsideCollaboratorsCount, y.Name)
			ch <- prometheus.MustNewConstMetric(e.Metrics["PendingOrgInvitationsCount"], prometheus.GaugeValue, y.PendingOrgInvitationsCount, y.Name)
		}

	}

	// Set Rate limit stats
	ch <- prometheus.MustNewConstMetric(e.Metrics["Limit"], prometheus.GaugeValue, e.RateLimits.Limit)
	ch <- prometheus.MustNewConstMetric(e.Metrics["Remaining"], prometheus.GaugeValue, e.RateLimits.Remaining)
	ch <- prometheus.MustNewConstMetric(e.Metrics["Reset"], prometheus.GaugeValue, e.RateLimits.Reset)

	// Clear Exporter, avoids multiple captures
	e.Repositories = nil
	e.ProcessedRepos = nil
	e.Organisations = nil
	e.Client = nil

}
