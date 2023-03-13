package exporter

import (
	conf "github.com/infinityworks/github-exporter/config"
	"github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

// New returns an initialized Exporter.
func New(c conf.Config, log *logrus.Logger) *Exporter {

	return &Exporter{
		Metrics: addMetrics(),
		Config:  c,
		Log:     log,
		Client:  newClient(c.APIToken),
	}

}

// addMetrics - Adds all the metrics to a map of strings, returns the map.
func addMetrics() map[string]*prometheus.Desc {

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

func derefString(s *string) string {

	if s != nil {
		return *s
	}

	return ""
}

func derefBool(b *bool) bool {

	if b != nil {
		return *b
	}

	return false
}

func derefInt(i *int) int {

	if i != nil {
		return *i
	}

	return 0

}
