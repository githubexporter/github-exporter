package exporter

import "github.com/prometheus/client_golang/prometheus"
import "strconv"

// AddMetrics - Add's all of the metrics to a map, returns the map.
func AddMetrics() map[string]*prometheus.Desc {

	APIMetrics := make(map[string]*prometheus.Desc)

	APIMetrics["Stars"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "stars"),
		"Total number of Stars for given repository",
		[]string{"repo", "user", "private"}, nil,
	)
	APIMetrics["OpenIssues"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "open_issues"),
		"Total number of open issues for given repository",
		[]string{"repo", "user", "private"}, nil,
	)
	APIMetrics["Watchers"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "watchers"),
		"Total number of watchers/subscribers for given repository",
		[]string{"repo", "user", "private"}, nil,
	)
	APIMetrics["Forks"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "forks"),
		"Total number of forks for given repository",
		[]string{"repo", "user", "private"}, nil,
	)
	APIMetrics["Size"] = prometheus.NewDesc(
		prometheus.BuildFQName("github", "repo", "size_kb"),
		"Size in KB for given repository",
		[]string{"repo", "user", "private"}, nil,
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

// processMetrics - Collects the data from the API, returns data object
func (e *Exporter) processMetrics(data []*APIResponse, rates *RateLimits, ch chan<- prometheus.Metric) error {

	// APIMetrics - range through the data object
	for _, y := range data {
		for _, x := range *y {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Stars"], prometheus.GaugeValue, x.Stars, x.Name, x.Owner.Login, strconv.FormatBool(x.Private))
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Forks"], prometheus.GaugeValue, x.Forks, x.Name, x.Owner.Login, strconv.FormatBool(x.Private))
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["OpenIssues"], prometheus.GaugeValue, x.OpenIssues, x.Name, x.Owner.Login, strconv.FormatBool(x.Private))
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Watchers"], prometheus.GaugeValue, x.Watchers, x.Name, x.Owner.Login, strconv.FormatBool(x.Private))
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Size"], prometheus.GaugeValue, x.Size, x.Name, x.Owner.Login, strconv.FormatBool(x.Private))
		}
	}

	// Rate limit stats
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Limit"], prometheus.GaugeValue, rates.Limit)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Remaining"], prometheus.GaugeValue, rates.Remaining)
	ch <- prometheus.MustNewConstMetric(e.APIMetrics["Reset"], prometheus.GaugeValue, rates.Reset)

	return nil
}
