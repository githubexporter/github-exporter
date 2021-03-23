package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"os"

	cfg "github.com/infinityworks/go-common/config"
)

var (
	prQueryOptionDefault string = "per_page=100&sort=long-running&direction=descending"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	*cfg.BaseConfig
	APIURL                string
	Repositories          string
	Organisations         string
	Users                 string
	APITokenEnv           string
	APITokenFile          string
	APIToken              string
	TargetURLs            []string
	PRLongRunningTimeDiff float64 //  Now - PR created time in hours; Above this PR considered to be long running
	PRQueryOptions        string  // API query options for /pulls

}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	listenPort := cfg.GetEnv("LISTEN_PORT", "9171")
	os.Setenv("LISTEN_PORT", listenPort)
	ac := cfg.Init()
	url := cfg.GetEnv("API_URL", "https://api.github.com")
	repos := os.Getenv("REPOS")
	orgs := os.Getenv("ORGS")
	users := os.Getenv("USERS")
	tokenEnv := os.Getenv("GITHUB_TOKEN")
	tokenFile := os.Getenv("GITHUB_TOKEN_FILE")
	token, err := getAuth(tokenEnv, tokenFile)
	scraped, err := getScrapeURLs(url, repos, orgs, users)
	// default for long running PRs is 2 weeks
	prLongRunningTimeDiff, err := strconv.ParseFloat(cfg.GetEnv("PR_LONG_RUNNING_TIME_DIFF", "336"), 64)
	prQueryOptions := cfg.GetEnv("PR_QUERY_OPTIONS", prQueryOptionDefault)

	if err != nil {
		log.Errorf("Error initialising Configuration, Error: %v", err)
	}

	appConfig := Config{
		&ac,
		url,
		repos,
		orgs,
		users,
		tokenEnv,
		tokenFile,
		token,
		scraped,
		prLongRunningTimeDiff,
		prQueryOptions,
	}

	return appConfig
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func getScrapeURLs(apiURL, repos, orgs, users string) ([]string, error) {

	urls := []string{}

	opts := "?&per_page=100" // Used to set the Github API to return 100 results per page (max)

	if len(repos) == 0 && len(orgs) == 0 && len(users) == 0 {
		log.Info("No targets specified. Only rate limit endpoint will be scraped")
	}

	// Append repositories to the array
	if repos != "" {
		rs := strings.Split(repos, ", ")
		for _, x := range rs {
			y := fmt.Sprintf("%s/repos/%s%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	// Append github orginisations to the array
	if orgs != "" {
		o := strings.Split(orgs, ", ")
		for _, x := range o {
			y := fmt.Sprintf("%s/orgs/%s/repos%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	if users != "" {
		us := strings.Split(users, ", ")
		for _, x := range us {
			y := fmt.Sprintf("%s/users/%s/repos%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	return urls, nil
}

// getAuth returns oauth2 token as string for usage in http.request
func getAuth(token string, tokenFile string) (string, error) {

	if token != "" {
		return token, nil
	} else if tokenFile != "" {
		b, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), err

	}

	return "", nil
}
