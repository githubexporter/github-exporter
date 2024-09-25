package config

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	cfg "github.com/infinityworks/go-common/config"
	log "github.com/sirupsen/logrus"
)

// Config struct holds all the runtime configuration for the application
type Config struct {
	*cfg.BaseConfig
	apiUrl                  *url.URL
	repositories            []string
	organisations           []string
	users                   []string
	apiToken                string
	targetURLs              []string
	gitHubApp               bool
	gitHubAppKeyPath        string
	gitHubAppId             int64
	gitHubAppInstallationId int64
	gitHubRateLimit         float64
}

// Init populates the Config struct based on environmental runtime configuration
func Init() Config {

	listenPort := cfg.GetEnv("LISTEN_PORT", "9171")
	os.Setenv("LISTEN_PORT", listenPort)
	ac := cfg.Init()

	appConfig := Config{
		&ac,
		nil,
		nil,
		nil,
		nil,
		"",
		nil,
		false,
		"",
		0,
		0,
		15000,
	}

	err := appConfig.SetAPIURL(cfg.GetEnv("API_URL", "https://api.github.com"))
	if err != nil {
		log.Errorf("Error initialising Configuration. Unable to parse API URL. Error: %v", err)
	}
	repos := os.Getenv("REPOS")
	if repos != "" {
		appConfig.SetRepositories(strings.Split(repos, ", "))
	}
	orgs := os.Getenv("ORGS")
	if orgs != "" {
		appConfig.SetOrganisations(strings.Split(orgs, ", "))
	}
	users := os.Getenv("USERS")
	if users != "" {
		appConfig.SetUsers(strings.Split(users, ", "))
	}

	gitHubApp := strings.ToLower(os.Getenv("GITHUB_APP"))
	if gitHubApp == "true" {
		gitHubAppKeyPath := os.Getenv("GITHUB_APP_KEY_PATH")
		gitHubAppId, _ := strconv.ParseInt(os.Getenv("GITHUB_APP_ID"), 10, 64)
		gitHubAppInstallationId, _ := strconv.ParseInt(os.Getenv("GITHUB_APP_INSTALLATION_ID"), 10, 64)
		gitHubRateLimit, _ := strconv.ParseFloat(cfg.GetEnv("GITHUB_RATE_LIMIT", "15000"), 64)
		appConfig.SetGitHubApp(true)
		appConfig.SetGitHubAppKeyPath(gitHubAppKeyPath)
		appConfig.SetGitHubAppId(gitHubAppId)
		appConfig.SetGitHubAppInstallationId(gitHubAppInstallationId)
		appConfig.SetGitHubRateLimit(gitHubRateLimit)
		err = appConfig.SetAPITokenFromGitHubApp()
		if err != nil {
			log.Errorf("Error initializing Configuration, Error: %v", err)
		}
	}

	tokenEnv := os.Getenv("GITHUB_TOKEN")
	tokenFile := os.Getenv("GITHUB_TOKEN_FILE")
	if tokenEnv != "" {
		appConfig.SetAPIToken(tokenEnv)
	} else if tokenFile != "" {
		err = appConfig.SetAPITokenFromFile(tokenFile)
		if err != nil {
			log.Errorf("Error initialising Configuration, Error: %v", err)
		}
	}
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

// Returns the GitHub App authentication value
func (c *Config) GitHubApp() bool {
	return c.gitHubApp
}

// Returns the GitHub app private key path
func (c *Config) GitHubAppKeyPath() string {
	return c.gitHubAppKeyPath
}

// Returns the GitHub app id
func (c *Config) GitHubAppId() int64 {
	return c.gitHubAppId
}

// Returns the GitHub app installation id
func (c *Config) GitHubAppInstallationId() int64 {
	return c.gitHubAppInstallationId
}

// Returns the GitHub RateLimit
func (c *Config) GitHubRateLimit() float64 {
	return c.gitHubRateLimit
}

// Sets the base API URL returning an error if the supplied string is not a valid URL
func (c *Config) SetAPIURL(u string) error {
	ur, err := url.Parse(u)
	c.apiUrl = ur
	return err
}

// Overrides the entire list of repositories
func (c *Config) SetRepositories(repos []string) {
	c.repositories = repos
	c.setScrapeURLs()
}

// Overrides the entire list of organisations
func (c *Config) SetOrganisations(orgs []string) {
	c.organisations = orgs
	c.setScrapeURLs()
}

// Overrides the entire list of users
func (c *Config) SetUsers(users []string) {
	c.users = users
	c.setScrapeURLs()
}

// SetAPIToken accepts a string oauth2 token for usage in http.request
func (c *Config) SetAPIToken(token string) {
	c.apiToken = token
}

// SetAPITokenFromFile accepts a file containing an oauth2 token for usage in http.request
func (c *Config) SetAPITokenFromFile(tokenFile string) error {
	b, err := os.ReadFile(tokenFile)
	if err != nil {
		return err
	}
	c.apiToken = strings.TrimSpace(string(b))
	return nil
}

// SetGitHubApp accepts a boolean for GitHub app authentication
func (c *Config) SetGitHubApp(githubApp bool) {
	c.gitHubApp = githubApp
}

// SetGitHubAppKeyPath accepts a string for GitHub app private key path
func (c *Config) SetGitHubAppKeyPath(gitHubAppKeyPath string) {
	c.gitHubAppKeyPath = gitHubAppKeyPath
}

// SetGitHubAppId accepts a string for GitHub app id
func (c *Config) SetGitHubAppId(gitHubAppId int64) {
	c.gitHubAppId = gitHubAppId
}

// SetGitHubAppInstallationId accepts a string for GitHub app installation id
func (c *Config) SetGitHubAppInstallationId(gitHubAppInstallationId int64) {
	c.gitHubAppInstallationId = gitHubAppInstallationId
}

// SetGitHubAppRateLimit accepts a string for GitHub RateLimit
func (c *Config) SetGitHubRateLimit(gitHubRateLimit float64) {
	c.gitHubRateLimit = gitHubRateLimit
}

// SetAPITokenFromGitHubApp generating api token from github app configuration.
func (c *Config) SetAPITokenFromGitHubApp() error {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, c.gitHubAppId, c.gitHubAppInstallationId, c.gitHubAppKeyPath)
	if err != nil {
		return err
	}
	strToken, err := itr.Token(context.Background())
	if err != nil {
		return err
	}
	c.SetAPIToken(strToken)
	return nil
}

// Init populates the Config struct based on environmental runtime configuration
// All URL's are added to the TargetURL's string array
func (c *Config) setScrapeURLs() error {

	urls := []string{}

	opts := map[string]string{"per_page": "100"} // Used to set the Github API to return 100 results per page (max)

	if len(c.repositories) == 0 && len(c.organisations) == 0 && len(c.users) == 0 {
		log.Info("No targets specified. Only rate limit endpoint will be scraped")
	}

	// Append repositories to the array
	if len(c.repositories) > 0 {
		for _, x := range c.repositories {
			y := *c.apiUrl
			y.Path = path.Join(y.Path, "repos", x)
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	// Append github orginisations to the array

	if len(c.organisations) > 0 {
		for _, x := range c.organisations {
			y := *c.apiUrl
			y.Path = path.Join(y.Path, "orgs", x, "repos")
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	if len(c.users) > 0 {
		for _, x := range c.users {
			y := *c.apiUrl
			y.Path = path.Join(y.Path, "users", x, "repos")
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	c.targetURLs = urls

	return nil
}
