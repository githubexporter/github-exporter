package exporter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

func (e *Exporter) asyncHTTPGets() ([]*Response, error) {

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(e.TargetURLs))

	responses := []*Response{}

	for _, url := range e.TargetURLs {

		go func(url string) {
			err := e.getResponse(url, ch)
			if err != nil {
				ch <- &Response{url, nil, []byte{}, err}
			}
		}(url)

	}

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				log.Errorf("Error scraping API, Error: %v", r.err)
				break
			}
			responses = append(responses, r)

			if len(responses) == len(e.TargetURLs) {
				return responses, nil
			}
		}

	}
}

// getResponse collects an individual http.response and returns a *Response
func (e *Exporter) getResponse(url string, ch chan<- *Response) error {

	log.Infof("Fetching %s \n", url)

	resp, err := getHTTPResponse(url, e.APIToken, e.APITokenFile)

	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Read the body to a byte array so it can be used elsewhere
	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return fmt.Errorf("Error converting body to byte array: %v", err)
	}

	// Triggers if a user specifies an invalid or not visible repository
	if resp.StatusCode == 404 {
		return fmt.Errorf("Error: Recieved 404 status from Github API, ensure the repsository URL is correct. If it's a privare repository, also check the oauth token is correct")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string, tokenFile string) (*http.Response, error) {

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	// Obtain auth token from file or environment
	a, err := getAuth(token, tokenFile)

	// If a token is present, add it to the http.request
	if a != "" {
		req.Header.Add("Authorization", "token "+a)
	}

	resp, err := client.Do(req)

	return resp, err
}
