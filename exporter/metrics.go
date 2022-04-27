package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
)

// AddMetrics - Add's all of the metrics to a map of strings, returns the map.
func AddMetrics() map[string]*prometheus.Desc {
	APIMetrics := make(map[string]*prometheus.Desc)
	APIMetrics["Issues"] = prometheus.NewDesc(
		prometheus.BuildFQName("jira", "project", "issue"),
		"A reference to a ticket on a JIRA metric",
		[]string{"project", "epic", "issue_owner", "issue_type", "assigned", "status", "priority", "votes", "creator"}, nil,
	)
	return APIMetrics
}

// processMetrics - processes the response data and sets the metrics using it as a source
func (e *Exporter) processMetrics(data []*Datum, rates *RateLimits, ch chan<- prometheus.Metric) error {
	// APIMetrics - range through the data slice
	for _, x := range data {
		for _, issue := range x.Issues {
			ch <- prometheus.MustNewConstMetric(e.APIMetrics["Issues"],
				prometheus.Counter,
				issue.Project,
				issue.Epic,
				issue.Owner,
				issue.IssueType,
				issue.Assigned,
				issue.Status,
				issue.Priority,
				issue.Votes,
				issue.Creator,
			)
		}
	}
	return nil
}
