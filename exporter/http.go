package exporter

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"strconv"
	"time"

	log "github.com/benri-io/jira-exporter/logger"
	"github.com/tomnomnom/linkheader"
)

// RateLimitExceededStatus is the status response from github when the rate limit is exceeded.
const RateLimitExceededStatus = "403 rate limit exceeded"

type PostRequest struct {
	target string
	data   interface{}
}

func asyncHTTPPosts(targets []PostRequest, token, user string) ([]*Response, error) {

	log.GetDefaultLogger().Infof("Searching through %d targets", len(targets))
	defer log.GetDefaultLogger().Infof("Done Searching through %d targets", len(targets))

	// Expand targets by following GitHub pagination links
	//targets = paginateTargets(targets, token)

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(targets))

	responses := []*Response{}

	for _, url := range targets {

		go func(url PostRequest) {
			err := postResponse(url.target, url.data, user, token, ch)
			if err != nil {
				ch <- &Response{url.target, nil, []byte{}, err}
			}
		}(url)

	}

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				log.GetDefaultLogger().Errorf("Error scraping API, Error: %v", r.err)
				break
			}
			responses = append(responses, r)

			if len(responses) == len(targets) {
				return responses, nil
			}
		}

	}
}

func asyncHTTPGets(targets []string, token string) ([]*Response, error) {

	log.GetDefaultLogger().Infof("Searching through %d targets", len(targets))
	defer log.GetDefaultLogger().Infof("Done Searching through %d targets", len(targets))

	// Expand targets by following GitHub pagination links
	targets = paginateTargets(targets, token)

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(targets))

	responses := []*Response{}

	for _, url := range targets {

		log.GetDefaultLogger().Infof("Getting url: %s", url)
		go func(url string) {
			err := getResponse(url, token, ch)
			if err != nil {
				ch <- &Response{url, nil, []byte{}, err}
			}
		}(url)

	}

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				log.GetDefaultLogger().Errorf("Error scraping API, Error: %v", r.err)
				break
			}
			responses = append(responses, r)

			if len(responses) == len(targets) {
				return responses, nil
			}
		}

	}
}

// paginateTargets returns all pages for the provided targets
func paginateTargets(targets []string, token string) []string {

	log.GetDefaultLogger().Infof("Paginating %d targets", len(targets))
	defer log.GetDefaultLogger().Infof("Done paginating through %d targets", len(targets))

	paginated := targets

	for _, url := range targets {

		// make a request to the original target to get link header if it exists
		resp, err := getHTTPResponse(url, token)
		if err != nil {
			log.GetDefaultLogger().Errorf("Error retrieving Link headers, Error: %s", err)
			continue
		}

		if resp.Header["Link"] != nil {
			links := linkheader.Parse(resp.Header["Link"][0])

			for _, link := range links {
				if link.Rel == "last" {

					u, err := neturl.Parse(link.URL)
					if err != nil {
						log.GetDefaultLogger().Errorf("Unable to parse page URL, Error: %s", err)
					}

					q := u.Query()

					lastPage, err := strconv.Atoi(q.Get("page"))
					if err != nil {
						log.GetDefaultLogger().Errorf("Unable to convert page substring to int, Error: %s", err)
					}

					// add all pages to the slice of targets to return
					for page := 2; page <= lastPage; page++ {
						q.Set("page", strconv.Itoa(page))
						u.RawQuery = q.Encode()
						paginated = append(paginated, u.String())
					}

					break
				}
			}
		}
	}
	return paginated
}

// getResponse collects an individual http.response and returns a *Response
func postResponse(url string, data interface{}, token, user string, ch chan<- *Response) error {

	log.GetDefaultLogger().Infof("Fetching %s \n", url)
	defer log.GetDefaultLogger().Infof("Done Fetching %s \n", url)

	resp, err := postHTTPResponse(url, data, token, user) // do this earlier
	if err != nil {
		return fmt.Errorf("Error during post: %v", err)
	}
	defer resp.Body.Close()

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == 404 {
		return fmt.Errorf("Error: Received 404 status from Jira API, ensure the repsository URL is correct")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

// getResponse collects an individual http.response and returns a *Response
func getResponse(url string, token string, ch chan<- *Response) error {

	log.GetDefaultLogger().Infof("Fetching %s \n", url)
	defer log.GetDefaultLogger().Infof("Done Fetching %s \n", url)

	resp, err := getHTTPResponse(url, token) // do this earlier
	if err != nil {
		return fmt.Errorf("Error fetching http response: %v", err)
	}
	defer resp.Body.Close()

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == 404 {
		return fmt.Errorf("Error: Received 404 status from Jira API, ensure the repsository URL is correct")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	log.GetDefaultLogger().Infof("Auth HTTP Response To %s", auth)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func postHTTPResponse(url string, data interface{}, token, user string) (*http.Response, error) {

	log.GetDefaultLogger().Infof("Getting HTTP Response To %s", url)

	dat, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	log.GetDefaultLogger().Infof("Posting to url: %s with payload: %v. User %s:%s ", url, string(dat), user, token)
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}")))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(user, token)

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	log.GetDefaultLogger().Infof("Got response: %v", resp.Status)

	// check rate limit exceeded.
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("%s", resp.Status)
	}

	// check rate limit exceeded.
	if resp.Status == RateLimitExceededStatus {
		resp.Body.Close()
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp, err
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string) (*http.Response, error) {
	log.GetDefaultLogger().Infof("Getting HTTP Response To %s", url)
	log.GetDefaultLogger().Infof("Getting HTTP Response To %s", url)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// If a token is present, add it to the http.request
	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	// check rate limit exceeded.
	if resp.Status == RateLimitExceededStatus {
		resp.Body.Close()
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return resp, err
}
