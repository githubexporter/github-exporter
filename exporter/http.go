package exporter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	// "strconv"
	"time"
	// "strings"
	log "github.com/sirupsen/logrus"
	"github.com/tomnomnom/linkheader"
)

// RateLimitExceededStatus is the status response from github when the rate limit is exceeded.
const RateLimitExceededStatus = "403 rate limit exceeded"

func asyncHTTPGets(targets []string, token string) ([]*Response, error) {
	// Expand targets by following GitHub pagination links
	targets = paginateTargets(targets, token)
	fmt.Printf("within asyncHTTPGets targets are :  %+v\n",targets)

	// Channels used to enable concurrent requests
	ch := make(chan *Response, len(targets))

	responses := []*Response{}

	for _, url := range targets {
		fmt.Printf("within asyncHTTPGets currently looping over  :  %+v\n",url)

		go func(url string) {
			err := getResponse(url, token, ch)
			fmt.Printf("getResponse error for :  %+v\n",url)
			fmt.Printf("getResponse error :  %+v\n",err)


			if err != nil {
				fmt.Printf("no getResponse error for :  %+v\n",url)
				ch <- &Response{url, nil, []byte{}, err}
			}
		}(url)

	}


	fmt.Printf("Channel is :  %+v\n",ch)

	for {
		select {
		case r := <-ch:
			if r.err != nil {
				log.Errorf("Error scraping API, Error: %v", r.err)
				return nil, r.err
			}
			responses = append(responses, r)
			fmt.Printf(" len(targets) is :  %+v\n", len(targets) )
			fmt.Printf(" len(responses) is :  %+v\n", len(responses) )
			for _, response := range responses{
				responseData := string(response.body[:])
				fmt.Printf(" response body is :  %+v\n", responseData )
			}

			if len(responses) == len(targets) {
				return responses, nil
			}
		}

	}
}

// paginateTargets returns all pages for the provided targets
func paginateTargets(targets []string, token string) []string {

	paginated := targets

	for _, urlTarget := range targets {
		fmt.Printf("targets are :  %+v\n",targets)

		// make a request to the original target to get link header if it exists
		resp, err := getHTTPResponse(urlTarget, token)
		if err != nil {
			log.Errorf("Error retrieving Link headers, Error: %s", err)
			continue
		}

		fmt.Printf("resp.Header[\"Link\"] is:  %+v\n",resp.Header["Link"])
		if resp.Header["Link"] != nil {
			links := linkheader.Parse(resp.Header["Link"][0])
			fmt.Printf("links are :  %+v\n",links)
			for _, link := range links {
				u, err := neturl.Parse(link.URL)
				if err != nil {
						log.Errorf("Unable to parse %v %s", u, err)
				 }

				// q := u.Query()
				// lastPage, err := strconv.Atoi(q.Get("page"))
				// subs := strings.Split(link.URL, "&page=")
				// lastPage, err := strconv.Atoi(subs[len(subs)-1])
				// fmt.Printf("query  is :  %+v\n",q)

				// fmt.Printf("lastPage is :  %+v\n",lastPage)

				if err != nil {
						log.Errorf("Unable to convert page substring to int, Error: %s", err)
				}


				// add all pages to the slice of targets to return
				// for page := 2; page <= lastPage; page++ {
			    for page := 2; page <= 100; page++ {
						pageURL := fmt.Sprintf("%s?page=%v", urlTarget, page)
						fmt.Printf("pageURL in paginateTarget loop is:  %+v\n",pageURL)

						paginated = append(paginated, pageURL)
				}
				break

			}
		}
	}
	return paginated
}

// getResponse collects an individual http.response and returns a *Response
func getResponse(url string, token string, ch chan<- *Response) error {

	log.Infof("Fetching %s \n", url)

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
		return fmt.Errorf("Error: Received 404 status from Github API, ensure the repository URL is correct. If it's a private repository, also check the oauth token is correct")
	}

	ch <- &Response{url, resp, body, err}

	return nil
}

// getHTTPResponse handles the http client creation, token setting and returns the *http.response
func getHTTPResponse(url string, token string) (*http.Response, error) {

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