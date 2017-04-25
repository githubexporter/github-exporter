package exporter

import (
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

// Describe describes all the metrics ever exported by the Exporter
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, met := range e.APIMetrics {
		ch <- met
	}

}

// Collect function, called on by Prometheus Client library
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	var data, rates, err = e.gatherData(ch)

	if err != nil {
		log.Errorf("Error gathering Data from remote API: %v", err)
		return
	}

	err = e.processMetrics(data, rates, ch)

	if err != nil {
		log.Error("Error Processing Metrics", err)
		return
	}

	log.Info("All Metrics successfully collected")

}
