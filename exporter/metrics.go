package exporter

import (
	"strconv"
	"strings"

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
	APIMetrics["CommitsHistory"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "commit", "count"),
		"Total number of commits for given repository and given branch",
		[]string{"repo", "branch", "author"}, nil,
	)
	APIMetrics["LatestCommit"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "commit", "latest"),
		"Latest Commit for a given repository and given branch",
		[]string{"repo", "branch", "author", "date", "commithash"}, nil,
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

// processMetrics - processes the response data and sets the metrics using it as a source
func (e *Exporter) processMetrics(data []*Datum, commitData []*CommitDatum, rates *RateLimits, ch chan<- prometheus.Metric) error {

	// APIMetrics - range through the data slice
	for _, x := range data {
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Stars"], prometheus.GaugeValue, x.Stars, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Forks"], prometheus.GaugeValue, x.Forks, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["OpenIssues"], prometheus.GaugeValue, x.OpenIssues, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Watchers"], prometheus.GaugeValue, x.Watchers, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Size"], prometheus.GaugeValue, x.Size, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
	}

	branch := e.Config.Branch
	latestCommits := make(map[string]*LatestCommitHistory)
	totalCommits := make(map[string]*CommitHistory)
	for _, x := range commitData {
		shortenedRepo := strings.Replace(x.URL, "https://api.github.com/repos/", "", -1)
		repo := shortenedRepo[:strings.Index(shortenedRepo, "/commits")]
		author := x.Commit.Author.Name
		if _, ok := latestCommits[author+repo]; !ok {
			date := strings.Split(x.Commit.Author.Date, "T")[0]
			hash := x.CommitHash
			latestCommits[author+repo] = &LatestCommitHistory{author, repo, date, hash}
		}
		if _, ok := totalCommits[author+repo]; ok {
			totalCommits[author+repo].Count++
		} else {
			totalCommits[author+repo] = &CommitHistory{author, repo, 1.0}
		}
	}
	for _, val := range totalCommits {
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["CommitsHistory"], prometheus.GaugeValue, val.Count, val.Repo, branch, val.Author)
	}
	for _, val := range latestCommits {
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["LatestCommit"], prometheus.GaugeValue, 1.0, val.Repo, branch, val.Author, val.Date, val.Hash)
	}

	// Set Rate limit stats
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Limit"], prometheus.GaugeValue, rates.Limit)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Remaining"], prometheus.GaugeValue, rates.Remaining)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Reset"], prometheus.GaugeValue, rates.Reset)

	return nil
}
