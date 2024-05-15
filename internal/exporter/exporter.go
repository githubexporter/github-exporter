package exporter

import (
	"context"
	"github.com/githubexporter/github-exporter/internal/config"
	"github.com/google/go-github/v61/github"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"strconv"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	ctx          context.Context
	logger       *slog.Logger
	githubClient *github.Client
	apiMetrics   map[string]*prometheus.Desc
	Config       *config.Config
}

func NewExporter(ctx context.Context, logger *slog.Logger, config *config.Config, githubClient *github.Client, apiMetrics map[string]*prometheus.Desc) Exporter {
	return Exporter{
		Config:       config,
		logger:       logger,
		ctx:          ctx,
		apiMetrics:   apiMetrics,
		githubClient: githubClient,
	}
}

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.apiMetrics {
		ch <- m
	}
}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is performed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	repos, err := e.getRepos()
	if err != nil {
		e.logger.Error("gathering repos", "error", err.Error())
	}

	rateLimits, err := e.getRateLimits()
	if err != nil {
		e.logger.Error("gathering rate limits", "error", err.Error())
	}

	err = e.processMetrics(ch, repos, rateLimits)
}

func (e *Exporter) processMetrics(ch chan<- prometheus.Metric, repos []*github.Repository, rateLimits *github.RateLimits) error {
	// TODO - unsafe pointer access
	for _, x := range repos {
		ch <- prometheus.MustNewConstMetric(e.apiMetrics["Stars"], prometheus.GaugeValue, float64(*x.StargazersCount), *x.Name, *x.Owner.Login, strconv.FormatBool(*x.Private), strconv.FormatBool(*x.Fork), strconv.FormatBool(*x.Archived), *x.License.Key, *x.Language)
		ch <- prometheus.MustNewConstMetric(e.apiMetrics["Forks"], prometheus.GaugeValue, float64(*x.ForksCount), *x.Name, *x.Owner.Login, strconv.FormatBool(*x.Private), strconv.FormatBool(*x.Fork), strconv.FormatBool(*x.Archived), *x.License.Key, *x.Language)
		ch <- prometheus.MustNewConstMetric(e.apiMetrics["Watchers"], prometheus.GaugeValue, float64(*x.SubscribersCount), *x.Name, *x.Owner.Login, strconv.FormatBool(*x.Private), strconv.FormatBool(*x.Fork), strconv.FormatBool(*x.Archived), *x.License.Key, *x.Language)
		ch <- prometheus.MustNewConstMetric(e.apiMetrics["Size"], prometheus.GaugeValue, float64(*x.Size), *x.Name, *x.Owner.Login, strconv.FormatBool(*x.Private), strconv.FormatBool(*x.Fork), strconv.FormatBool(*x.Archived), *x.License.Key, *x.Language)

		releases, err := e.getReleases(*x.Owner.Login, *x.Name)
		if err != nil {
			e.logger.Error("getting releases", "error", err.Error())
			continue
		}
		for _, release := range releases {
			for _, asset := range release.Assets {
				ch <- prometheus.MustNewConstMetric(e.apiMetrics["ReleaseDownloads"], prometheus.GaugeValue, float64(*asset.DownloadCount), *x.Name, *x.Owner.Login, *release.Name, *asset.Name, *release.TagName, asset.CreatedAt.String())
			}
		}

		prCount, err := e.getPullRequestCount(*x.Owner.Login, *x.Name)
		if err != nil {
			e.logger.Error("getting pull request count", "error", err.Error())
			continue
		}

		issueCount := *x.OpenIssuesCount - prCount
		ch <- prometheus.MustNewConstMetric(e.apiMetrics["OpenIssues"], prometheus.GaugeValue, float64(issueCount), *x.Name, *x.Owner.Login, strconv.FormatBool(*x.Private), strconv.FormatBool(*x.Fork), strconv.FormatBool(*x.Archived), *x.License.Key, *x.Language)
		ch <- prometheus.MustNewConstMetric(e.apiMetrics["PullRequestCount"], prometheus.GaugeValue, float64(prCount), *x.Name, *x.Owner.Login)
	}

	// Set Rate limit stats
	ch <- prometheus.MustNewConstMetric(e.apiMetrics["Limit"], prometheus.GaugeValue, float64(rateLimits.Core.Limit))
	ch <- prometheus.MustNewConstMetric(e.apiMetrics["Remaining"], prometheus.GaugeValue, float64(rateLimits.Core.Remaining))
	ch <- prometheus.MustNewConstMetric(e.apiMetrics["Reset"], prometheus.GaugeValue, float64(rateLimits.Core.Reset.Unix()))

	return nil

}
