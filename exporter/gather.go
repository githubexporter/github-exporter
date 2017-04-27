package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

// gatherData - Collects the data from thw API, invokes functions to transform that data into metrics
func (e *Exporter) gatherData(ch chan<- prometheus.Metric) ([]*APIResponse, *RateLimits, error) {

	apid := []*APIResponse{}

	for _, u := range e.TargetURLs {

		// Create new data slice from Struct for organisation data
		var d = new(APIResponse)
		var da = []*APIResponse{}

		resp, err := e.getHttpResponse(u)

		if err != nil {
			log.Errorf("Error requesting http data from API for repository: %s. Got Error: %s", u, err)
			return nil, nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		ia := isArray(body)

		if err != nil {
			log.Errorf("Failed to determine API response")
			return nil, nil, err
		}

		if ia {
			log.Info("ARRAY!")
			json.Unmarshal(body, &da)
			apid = append(apid, da...)

		} else if !ia {
			log.Info("NOT ARRAY!")
			apid = append(apid, d)
			json.Unmarshal(body, &d)

		}

		if err != nil {
			log.Errorf("Error gathering JSON data for repository: %s. Got Error: %s", u, err)
			return nil, nil, err
		}

		// Close the response body, the underlying Transport should then close the connection.
		resp.Body.Close()

		log.Infof("API data fetched for repository: %s", u)
	}

	rates, err := e.getRates(e.APIURL)

	if err != nil {
		return apid, rates, err
	}

	return apid, rates, err
}

// getAuth returns oauth2 token as string for usage in http.request
func (e *Exporter) getAuth() (string, error) {

	if e.APIToken != "" {
		return e.APIToken, nil
	} else if e.APITokenFile != "" {
		b, err := ioutil.ReadFile(e.APITokenFile)
		if err != nil {
			return "", err
		}
		return string(b), err

	}

	return "", nil
}

func (e *Exporter) getRates(baseURL string) (*RateLimits, error) {

	rateEndPoint := ("/rate_limit")
	url := fmt.Sprintf("%s%s", baseURL, rateEndPoint)

	resp, err := e.getHttpResponse(url)

	if err != nil {
		log.Errorf("Error requesting http data from API for repository: %s. Got Error: %s", url, err)
		return &RateLimits{}, err
	}

	limit, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Limit"), 64)

	if err != nil {
		return &RateLimits{}, err
	}

	rem, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Remaining"), 64)

	if err != nil {
		return &RateLimits{}, err
	}

	reset, err := strconv.ParseFloat(resp.Header.Get("X-RateLimit-Reset"), 64)

	if err != nil {
		return &RateLimits{}, err
	}

	return &RateLimits{
		Limit:     limit,
		Remaining: rem,
		Reset:     reset,
	}, err

}

func (e *Exporter) getHttpResponse(url string) (*http.Response, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	a, err := e.getAuth()

	if err != nil {
		return nil, err
	}

	if a != "" {
		req.Header.Add("Authorization", "token "+a)
	}

	resp, err := client.Do(req)

	if err != nil {
		return resp, err
	}

	return resp, nil
}

func isArray(body []byte) bool {

	isArray := false

	for _, c := range body {
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			continue
		}
		isArray = c == '['
		break
	}

	return isArray

}
