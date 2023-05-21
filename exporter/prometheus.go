package exporter

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation"
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

	gitHubApp := os.Getenv("GITHUB_APP")
	if strings.ToLower(gitHubApp) == "true" {
		needReAuth, err := e.isTokenExpired()
		if err != nil {
			log.Errorf("Error checking token expiration status: %v", err)
			return
		}
		if needReAuth{
			e.reAuth()
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

func (e *Exporter) isTokenExpired () (bool,error){
	u := *e.APIURL()
	u.Path = path.Join(u.Path, "rate_limit")

	resp, err := getHTTPResponse(u.String(), e.APIToken())

	if err != nil {
		return false,err
	}
	defer resp.Body.Close()
	// Triggers if rate-limiting isn't enabled on private Github Enterprise installations
	if resp.StatusCode == 404 {
		return false, errors.New("404 Error")
	}

	limit, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Limit"), 64)

	if err != nil {
		return false, err
	}
	
	defaultLimit := os.Getenv("GITHUB_RATE_LIMIT")
	if len(defaultLimit) == 0 {
        defaultLimit = "15000"
    }
	defaultLimitInt,err := strconv.ParseInt(defaultLimit, 10, 64)	
	if err != nil {
		return false,err
	}
	if limit < float64(defaultLimitInt){
		return true, nil
	}
	return false, nil
	
}

func (e *Exporter) reAuth () error{
	gitHubAppKeyPath := os.Getenv("GITHUB_APP_KEY_PATH")
	gitHubAppId,_ := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64) 
	gitHubAppInstalaltionId,_ := strconv.ParseInt(os.Getenv("GITHUB_APP_INSTALLATION_ID"), 10, 64)
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, gitHubAppId, gitHubAppInstalaltionId, gitHubAppKeyPath)
	if err != nil {
		return err
	}	
	strToken,err := itr.Token(context.Background())
	if err != nil{
		return err
	}
	e.Config.SetAPIToken(strToken) 
	return nil
}