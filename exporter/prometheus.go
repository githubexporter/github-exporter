package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, m := range e.APIMetrics {
		ch <- m
	}

}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is peformed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	// Scrape the Data and CommitData from Github
	var data, commitData, rates, err = e.gatherData()

	if err != nil {
		log.Errorf("Error gathering Data from remote API: %v", err)
		return
	}

	// Set prometheus gauge metrics using the data gathered
	err = e.processMetrics(data, commitData, rates, ch)

	if err != nil {
		log.Error("Error Processing Metrics", err)
		return
	}

	log.Info("All Metrics successfully collected")

}
