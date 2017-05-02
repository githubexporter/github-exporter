package config

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	"os"

	cfg "github.com/infinityworksltd/go-common/config"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	*cfg.BaseConfig
	APIURL        string
	Repositories  string
	Organisations string
	APIToken      string
	APITokenFile  string
	TargetURLs    []string
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	ac := cfg.Init()

	scraped, err := getScrapeURLs()

	if err != nil {
		log.Errorf("Error initialising Configuration, Error: %v", err)
	}

	appConfig := Config{
		&ac,
		cfg.GetEnv("API_URL", "https://api.github.com"),
		os.Getenv("REPOS"),
		os.Getenv("ORGS"),
		os.Getenv("GITHUB_TOKEN"),
		os.Getenv("GITHUB_TOKEN_FILE"),
		scraped,
	}

	return appConfig
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func getScrapeURLs() ([]string, error) {

	urls := []string{}
	apiURL := cfg.GetEnv("API_URL", "https://api.github.com")
	repos := os.Getenv("REPOS")
	orgs := os.Getenv("ORGS")
	opts := "?&per_page=100" // Used to set the Github API to return 100 results per page (max)

	// User input validation, check that either repositories or organisations have been passed in
	if len(repos) == 0 && len(orgs) == 0 {
		return urls, fmt.Errorf("No organisations or repositories specified")
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

	return urls, nil
}
