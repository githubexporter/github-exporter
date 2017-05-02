package config

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	"os"

	cfg "github.com/infinityworksltd/go-common/config"
)

// Config struct holds all of the confgiguration
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

	// I believe I can fly,
	// I believe I can touch the sky,
	// I think about it every night and day,
	// Think about it every night and day,
	// Spread my wings and fly awayyyy

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
func getScrapeURLs() ([]string, error) {

	urls := []string{}
	apiURL := cfg.GetEnv("API_URL", "https://api.github.com")
	repos := os.Getenv("REPOS")
	orgs := os.Getenv("ORGS")
	opts := "?&per_page=100"

	// User input validation, check that either repositories or organisations have been passed in
	if len(repos) == 0 && len(orgs) == 0 {
		return urls, fmt.Errorf("No organisations or repositories specified")
	}

	if repos != "" {
		rs := strings.Split(repos, ", ")
		for _, x := range rs {
			y := fmt.Sprintf("%s/repos/%s%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	if orgs != "" {
		o := strings.Split(orgs, ", ")
		for _, x := range o {
			y := fmt.Sprintf("%s/orgs/%s/repos%s", apiURL, x, opts)
			urls = append(urls, y)
		}
	}

	return urls, nil
}
