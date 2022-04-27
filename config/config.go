package config

import (
	"io/ioutil"
	"net/url"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"os"
)

// Config struct holds all of the runtime confgiguration for the application
type Config struct {
	*BaseConfig
	apiUrl     *url.URL
	projects   []string
	apiToken   string
	jiraUser   string
	targetURLs []string
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	listenPort := GetEnv("LISTEN_PORT", "9171")
	os.Setenv("LISTEN_PORT", listenPort)
	ac := InitBaseConfig()

	appConfig := Config{
		&ac,
		nil,
		nil,
		"",
		"",
		nil,
	}

	err := appConfig.SetAPIURL(GetEnv("JIRA_API_URL", "https://benri.atlassian.net/"))
	if err != nil {
		log.Errorf("Error initialising Configuration. Unable to parse API URL. Error: %v", err)
	}
	repos := os.Getenv("PROJECTS")
	if repos != "" {
		appConfig.SetProjects(strings.Split(repos, ", "))
	}

	appConfig.SetJiraUser(GetEnv("JIRA_USER", ""))

	tokenEnv := os.Getenv("JIRA_API_TOKEN")
	tokenFile := os.Getenv("JIRA_TOKEN_FILE")
	if tokenEnv != "" {
		appConfig.SetAPIToken(tokenEnv)
	} else if tokenFile != "" {
		err = appConfig.SetAPITokenFromFile(tokenFile)
		if err != nil {
			log.Errorf("Error initialising Configuration, Error: %v", err)
		}
	}
	appConfig.setScrapeURLs()
	return appConfig
}

// Returns the base APIURL
func (c *Config) APIURL() *url.URL {
	return c.apiUrl
}

// Returns a list of all object URLs to scrape
func (c *Config) TargetURLs() []string {
	return c.targetURLs
}

// Returns the oauth2 token for usage in http.request
func (c *Config) APIToken() string {
	return c.apiToken
}

func (c *Config) User() string {
	return c.jiraUser
}

// Sets the base API URL returning an error if the supplied string is not a valid URL
func (c *Config) SetAPIURL(u string) error {
	ur, err := url.Parse(u)
	c.apiUrl = ur
	return err
}

// Sets the base API URL returning an error if the supplied string is not a valid URL
func (c *Config) SetJiraUser(u string) {
	c.jiraUser = u
}

// Overrides the entire list of projects
func (c *Config) SetProjects(projects []string) {
	c.projects = projects
	c.setScrapeURLs()
}

// SetAPIToken accepts a string oauth2 token for usage in http.request
func (c *Config) SetAPIToken(token string) {
	c.apiToken = token
}

// SetAPITokenFromFile accepts a file containing an oauth2 token for usage in http.request
func (c *Config) SetAPITokenFromFile(tokenFile string) error {
	b, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return err
	}
	c.apiToken = strings.TrimSpace(string(b))
	return nil
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func (c *Config) setScrapeURLs() error {

	urls := []string{}

	opts := map[string]string{} // Used to set the Jira API to return 100 results per page (max)

	y := *c.apiUrl
	y.Path = path.Join(y.Path, "/search")
	q := y.Query()
	for k, v := range opts {
		q.Add(k, v)
	}
	y.RawQuery = q.Encode()
	urls = append(urls, y.String())

	c.targetURLs = urls

	log.Infof("Got %d targets", len(urls))

	return nil
}
