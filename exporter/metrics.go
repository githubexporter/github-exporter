package exporter

import (
	"fmt"
	"strconv"

	"github.com/githubexporter/github-exporter/config"

	"github.com/prometheus/client_golang/prometheus"
)

func NewExporter(cfg *config.Config) (*Exporter, error) {
	client, err := cfg.GetClient()
	if err != nil {
		return nil, fmt.Errorf("getting client: %w", err)
	}

	return &Exporter{
		APIMetrics: AddMetrics(cfg),
		Config:     *cfg,
		Client:     client,
	}, nil
}

// AddMetrics - Adds all the metrics to a map of strings, returns the map.
func AddMetrics(cfg *config.Config) map[string]*prometheus.Desc {

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

	if cfg.GitHubRateLimitEnabled {
		rateLimitLabels := []string{"resource"}
		APIMetrics["Limit"] = prometheus.NewDesc(
			prometheus.BuildFQName("github", "rate", "limit"),
			"Number of API queries allowed in a 60 minute window",
			rateLimitLabels, nil,
		)
		APIMetrics["Remaining"] = prometheus.NewDesc(
			prometheus.BuildFQName("github", "rate", "remaining"),
			"Number of API queries remaining in the current window",
			rateLimitLabels, nil,
		)
		APIMetrics["Reset"] = prometheus.NewDesc(
			prometheus.BuildFQName("github", "rate", "reset"),
			"The time at which the current rate limit window resets in UTC epoch seconds",
			rateLimitLabels, nil,
		)
	}

	return APIMetrics
}

// processMetrics - processes the response data and sets the metrics using it as a source
func (e *Exporter) processMetrics(data []*Datum, rates *[]RateLimit, ch chan<- prometheus.Metric) error {
	processed := make(map[string]struct{})

	for _, x := range data {
		// Check if we've already processed this repo
		repo := fmt.Sprintf("%s/%s", x.Owner.Login, x.Name)
		if _, ok := processed[repo]; ok {
			continue
		}
		processed[repo] = struct{}{}

		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Stars"], prometheus.GaugeValue, float64(x.Stars), x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Forks"], prometheus.GaugeValue, float64(x.Forks), x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Watchers"], prometheus.GaugeValue, float64(x.Watchers), x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["Size"], prometheus.GaugeValue, float64(x.Size), x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)

		for _, release := range x.Releases {
			for _, asset := range release.Assets {
				ch <- prometheus.MustNewConstMetric(e.APIMetrics["ReleaseDownloads"], prometheus.GaugeValue, float64(asset.Downloads), x.Name, x.Owner.Login, release.Name, asset.Name, release.Tag, asset.CreatedAt)
			}
		}
		prCount := 0
		for range x.Pulls {
			prCount += 1
		}
		// issueCount = x.OpenIssue - prCount
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["OpenIssues"], prometheus.GaugeValue, float64(x.OpenIssues-prCount), x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)

		// prCount
		ch <- prometheus.MustNewConstMetric(e.APIMetrics["PullRequestCount"], prometheus.GaugeValue, float64(prCount), x.Name, x.Owner.Login)
	}

	// Set Rate limit stats
	if e.GitHubRateLimitEnabled && rates != nil {
		for _, r := range *rates {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Limit"], prometheus.GaugeValue, r.Limit, r.Resource)
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Remaining"], prometheus.GaugeValue, r.Remaining, r.Resource)
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Reset"], prometheus.GaugeValue, r.Reset, r.Resource)
		}
	}

	return nil
}
