package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)
import "strconv"

// AddMetrics - Adds all of the metrics to a map of strings, returns the map.
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
func (e *Exporter) processMetrics(data []*Datum, rates *RateLimits, ch chan<- prometheus.Metric) error {

	// APIMetrics - range through the data slice
	for _, x := range data {
		if viper.GetBool("stars") {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Stars"], prometheus.GaugeValue, x.Stars, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		}
		if viper.GetBool("forks") {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Forks"], prometheus.GaugeValue, x.Forks, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		}
		if viper.GetBool("issues") {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["OpenIssues"], prometheus.GaugeValue, x.OpenIssues, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		}
		if viper.GetBool("watchers") {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Watchers"], prometheus.GaugeValue, x.Watchers, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		}
		if viper.GetBool("size") {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Size"], prometheus.GaugeValue, x.Size, x.Name, x.Owner.Login, strconv.FormatBool(x.Private), strconv.FormatBool(x.Fork), strconv.FormatBool(x.Archived), x.License.Key, x.Language)
		}
	}

	// Set Rate limit stats
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Limit"], prometheus.GaugeValue, rates.Limit)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Remaining"], prometheus.GaugeValue, rates.Remaining)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Reset"], prometheus.GaugeValue, rates.Reset)

	return nil
}
