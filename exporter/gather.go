package exporter

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/benri-io/jira-exporter/logger"
	log "github.com/benri-io/jira-exporter/logger"
)

// gatherData - jCollects the data from the API and stores into struct
func (e *Exporter) gatherData() ([]*Datum, error) {

	log.GetDefaultLogger().Info("Gathering Data with %d targets", len(e.TargetURLs()))
	defer log.GetDefaultLogger().Infof("Done gathering data")

	data := []*Datum{}

	responses, err := asyncHTTPGets(e.TargetURLs(), e.APIToken())

	if err != nil {
		return data, err
	}

	for _, response := range responses {

		// Jira can at times present an array, or an object for the same data set.
		// This code checks handles this variation.
		if isArray(response.body) {
			ds := []*Datum{}
			json.Unmarshal(response.body, &ds)
			data = append(data, ds...)
		} else {
			d := new(Datum)

			// Get releases
			if strings.Contains(response.url, "/search") {
				getIssues(e, response.url, &d.Issues)
			}

			json.Unmarshal(response.body, &d)
			data = append(data, d)
		}

		log.GetDefaultLogger().Infof("API data fetched for repository: %s", response.url)
	}

	//return data, rates, err
	return data, nil

}

// getRates obtains the rate limit data for requests against the github API.
// Especially useful when operating without oauth and the subsequent lower cap.
func (e *Exporter) getRates() (*RateLimits, error) {

	log.GetDefaultLogger().Infof("Getting rates")
	defer log.GetDefaultLogger().Infof("Done getting rates")

	u := *e.APIURL()
	u.Path = path.Join(u.Path, "rate_limit")

	resp, err := getHTTPResponse(u.String(), e.APIToken())
	if err != nil {
		return &RateLimits{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &RateLimits{}, fmt.Errorf("Rate Limiting not enabled in JIRA API")
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

type JQLRequest struct {
	JQL          string   `json:"jql"`
	MaxResults   int      `json:"maxResults"`
	FieldsByKeys bool     `json:"fieldsByKeys"`
	Fields       []string `json:"fields"`
	StartAt      int      `json:"startAt"`
}

func getIssues(e *Exporter, url string, data *[]IssueMetric) {

	log.GetDefaultLogger().Infof("Getting issues: %s", url)
	defer log.GetDefaultLogger().Infof("Done getting issues")

	i := strings.Index(url, "?")
	if i > -1 {
		url = url[:i]
	}
	issuesURL := url //+ "/search"

	req := []PostRequest{PostRequest{
		target: issuesURL,
		data: JQLRequest{
			JQL:          fmt.Sprintf("(updated >= -%dh and status CHANGED) or (created >= -%dh", 24, 24),
			MaxResults:   -1,
			FieldsByKeys: false,
			Fields:       []string{"*all"},
			StartAt:      0,
		},
	},
	}

	issueResponse, err := asyncHTTPPosts(req, e.User(), e.APIToken())
	if err != nil {
		logger.GetDefaultLogger().Errorf("Unable to obtain issues from API, Error: %s", err)
	}
	dat, _ := json.Marshal(issueResponse)
	if dat != nil {
		logger.GetDefaultLogger().Infof("Got response: %v", string(dat))
	}
	var response SearchResponse
	err = json.Unmarshal(issueResponse[0].body, &response)
	if err != nil {
		log.GetDefaultLogger().Errorf("Error marshalling response: %s", err)
	}
	//json.Unmarshal(issueResponse[0].body, &data)

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
