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

// gatherData - Collects the data from the API and stores into struct
func (e *Exporter) gatherData(ch chan<- prometheus.Metric) ([]*APIResponse, *RateLimits, error) {

	aResponses := []*APIResponse{}

	// Scrapes are peformed per URL and data is appended to a slice
	for _, u := range e.TargetURLs {

		resp, err := e.getHTTPResponse(u)

		if err != nil {
			log.Errorf("Error requesting http data from API for repository: %s. Got Error: %s", u, err)
			return nil, nil, err
		}

		// Read the body into a string so we can parse it twice (isArray & Unmarshal)
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Errorf("Failed to read response body, error: %s", err)
			return nil, nil, err
		}

		// Github can at times present an array, or an object for the same data set.
		// This code checks handles this variation.
		if isArray(body) {
			dataSlice := []*APIResponse{}
			json.Unmarshal(body, &dataSlice)
			aResponses = append(aResponses, dataSlice...)
		} else {
			data := new(APIResponse)
			json.Unmarshal(body, &data)
			aResponses = append(aResponses, data)
		}

		log.Infof("API data fetched for repository: %s", u)
	}

	// Check the API rate data and store as a metric
	rates, err := e.getRates(e.APIURL)

	if err != nil {
		return aResponses, rates, err
	}

	return aResponses, rates, err
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

// getRates obtains the rate limit data for requests against the github API.
// Especially useful when operating without oauth and the subsequent lower cap.
func (e *Exporter) getRates(baseURL string) (*RateLimits, error) {

	rateEndPoint := ("/rate_limit")
	url := fmt.Sprintf("%s%s", baseURL, rateEndPoint)

	resp, err := e.getHTTPResponse(url)

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

// getHTTPResponse creates a http client, issues a GET and returns the http.Response
func (e *Exporter) getHTTPResponse(url string) (*http.Response, error) {

	client := &http.Client{}

	// (expensive but robus at these volumes)
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// Obtain auth token from file or environment
	a, err := e.getAuth()

	if err != nil {
		return nil, err
	}

	// If a token is present, add it to the http.request
	if a != "" {
		req.Header.Add("Authorization", "token "+a)
	}

	resp, err := client.Do(req)

	if err != nil {
		return resp, err
	}

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == 404 {
		return resp, fmt.Errorf("404 Recieved from GitHub API from URL %s", url)
	}

	return resp, nil
}

// isArray simply looks for key details that determine if the JSON response is an array or not.
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
