package exporter

import (
	log "github.com/benri-io/jira-exporter/logger"
	"github.com/prometheus/client_golang/prometheus"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	log.GetDefaultLogger().Infof("Describing metrics")

	for _, m := range e.APIMetrics {
		ch <- m
	}
	log.GetDefaultLogger().Infof("Done describing metrics")

}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is peformed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	log.GetDefaultLogger().Infof("Collecting metrics with %v targets", len(e.TargetURLs()))
	defer log.GetDefaultLogger().Infof("Done Collecting metrics")

	data := []*Datum{}
	var err error
	// Scrape the Data from Github
	if len(e.TargetURLs()) > 0 {
		data, err = e.gatherData()
		if err != nil {
			log.GetDefaultLogger().Errorf("Error gathering Data from remote API: %v", err)
			return
		}
	}

	// // Set prometheus gauge metrics using the data gathered
	err = e.processMetrics(data, nil, ch)

	if err != nil {
		log.GetDefaultLogger().Error("Error Processing Metrics", err)
		return
	}

	log.GetDefaultLogger().Info("All Metrics successfully collected")

}
