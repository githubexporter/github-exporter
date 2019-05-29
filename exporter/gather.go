package exporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// gatherData - Collects the data from the API and stores into struct
func (e *Exporter) gatherData() ([]*Datum, []*CommitDatum, *RateLimits, error) {

	data := []*Datum{}
	commitData := []*CommitDatum{}

	responses, err := asyncHTTPGets(e.TargetURLs, e.APIToken)

	if err != nil {
		return data, commitData, nil, err
	}

	opts := "?&per_page=100"
	for _, response := range responses {

		// Github can at times present an array, or an object for the same data set.
		// This code checks handles this variation.
		if isArray(response.body) {
			if isCommitData(response.body) {
				cds := []*CommitDatum{}
				json.Unmarshal(response.body, &cds)
				commitData = append(commitData, cds...)
				for len(commitData[len(commitData)-1].Parents) != 0 {
					apiURL := strings.Split(commitData[len(commitData)-1].URL, "/commits/")[0]
					urls := []string{fmt.Sprintf("%s/commits%s&sha=%s", apiURL, opts, commitData[len(commitData)-1].CommitHash)}
					responsesNext, err := asyncHTTPGets(urls, e.APIToken)
					if err != nil {
						break
					}
					for _, r := range responsesNext {
						cds = []*CommitDatum{}
						json.Unmarshal(r.body, &cds)
						commitData = append(commitData, cds...)
					}
				}
			} else {
				ds := []*Datum{}
				json.Unmarshal(response.body, &ds)
				data = append(data, ds...)
			}
		} else {
			if isCommitData(response.body) {
				cd := new(CommitDatum)
				json.Unmarshal(response.body, &cd)
				commitData = append(commitData, cd)
			} else {
				d := new(Datum)
				json.Unmarshal(response.body, &d)
				data = append(data, d)
			}
		}

		log.Infof("API data fetched for repository: %s", response.url)
	}

	// Check the API rate data and store as a metric
	rates, err := getRates(e.APIURL, e.APIToken)

	if err != nil {
		log.Errorf("Unable to obtain rate limit data from API, Error: %s", err)
	}

	//return data, commitData, rates, err
	return data, commitData, rates, nil

}

// getRates obtains the rate limit data for requests against the github API.
// Especially useful when operating without oauth and the subsequent lower cap.
func getRates(baseURL string, token string) (*RateLimits, error) {

	rateEndPoint := ("/rate_limit")
	url := fmt.Sprintf("%s%s", baseURL, rateEndPoint)

	resp, err := getHTTPResponse(url, token)
	if err != nil {
		return &RateLimits{}, err
	}
	defer resp.Body.Close()

	// Triggers if rate-limiting isn't enabled on private Github Enterprise installations
	if resp.StatusCode == 404 {
		return &RateLimits{}, fmt.Errorf("Rate Limiting not enabled in GitHub API")
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

func isCommitData(body []byte) bool {

	isCommitData := false

	data := body[:10]
	if bytes.Contains(data, []byte(`"sha":`)) {
		isCommitData = true
	}

	return isCommitData
}
