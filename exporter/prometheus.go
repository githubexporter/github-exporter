package exporter

import (
	"path"
	"strconv"

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
	data := []*Datum{}
	var err error

	if e.Config.GitHubApp() {
		needReAuth, err := e.isTokenExpired()
		if err != nil {
			log.Errorf("Error checking token expiration status: %v", err)
			return
		}
		if needReAuth {
			err = e.Config.SetAPITokenFromGitHubApp()
			if err != nil {
				log.Errorf("Error authenticating with GitHub app: %v", err)
			}
		}
	}
	// Scrape the Data from Github
	if len(e.TargetURLs()) > 0 {
		data, err = e.gatherData()
		if err != nil {
			log.Errorf("Error gathering Data from remote API: %v", err)
			return
		}
	}

	rates, err := e.getRates()
	if err != nil {
		log.Errorf("Error gathering Rates from remote API: %v", err)
		return
	}

	// Set prometheus gauge metrics using the data gathered
	err = e.processMetrics(data, rates, ch)

	if err != nil {
		log.Error("Error Processing Metrics", err)
		return
	}

	log.Info("All Metrics successfully collected")

}

func (e *Exporter) isTokenExpired() (bool, error) {
	u := *e.APIURL()
	u.Path = path.Join(u.Path, "rate_limit")

	resp, err := getHTTPResponse(u.String(), e.APIToken())

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	// Triggers if rate-limiting isn't enabled on private Github Enterprise installations
	if resp.StatusCode == 404 {
		return false, nil
	}

	limit, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Limit"), 64)

	if err != nil {
		return false, err
	}

	defaultRateLimit := e.Config.GitHubRateLimit()
	if limit < defaultRateLimit {
		return true, nil
	}
	return false, nil

}
