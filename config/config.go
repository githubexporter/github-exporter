package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v76/github"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

// Config struct holds runtime configuration required for the application
type Config struct {
	MetricsPath            string   `envconfig:"METRICS_PATH" required:"false" default:"/metrics"`
	ListenPort             string   `envconfig:"LISTEN_PORT" required:"false" default:"9171"`
	LogLevel               string   `envconfig:"LOG_LEVEL" required:"false" default:"INFO"`
	AppName                string   `envconfig:"APP_NAME" required:"false" default:"github-exporter"`
	ApiUrl                 *url.URL `envconfig:"API_URL" required:"false" default:"https://api.github.com"`
	Repositories           []string `envconfig:"REPOS" required:"false"`
	Organisations          []string `envconfig:"ORGS" required:"false"`
	Users                  []string `envconfig:"USERS" required:"false"`
	GithubToken            string   `envconfig:"GITHUB_TOKEN" required:"false"`
	GithubTokenFile        string   `envconfig:"GITHUB_TOKEN_FILE" required:"false"`
	GitHubApp              bool     `envconfig:"GITHUB_APP" required:"false" default:"false"`
	GitHubRateLimitEnabled bool     `envconfig:"GITHUB_RATE_LIMIT_ENABLED" required:"false" default:"true"`
	*GitHubAppConfig       `ignored:"true"`
}

type GitHubAppConfig struct {
	GitHubAppKeyPath        string `envconfig:"GITHUB_APP_KEY_PATH" required:"false" default:""`
	GitHubAppId             int64  `envconfig:"GITHUB_APP_ID" required:"false" default:""`
	GitHubAppInstallationId int64  `envconfig:"GITHUB_APP_INSTALLATION_ID" required:"false" default:""`
}

// Init populates the Config struct based on environmental runtime configuration
func Init() (*Config, error) {

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("processing envconfig: %v", err)
	}

	// Parse and set log level
	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parsing log level: %v", err)
	}
	log.SetLevel(level)

	// Trim whitespace from repositories, organisations, and users
	cfg.Repositories = mapSlice(cfg.Repositories, strings.TrimSpace)
	cfg.Organisations = mapSlice(cfg.Organisations, strings.TrimSpace)
	cfg.Users = mapSlice(cfg.Users, strings.TrimSpace)

	// Process GitHub App config if enabled
	if cfg.GitHubApp {
		var appCfg GitHubAppConfig
		if err := envconfig.Process("", &appCfg); err != nil {
			return nil, fmt.Errorf("processing GitHub App envconfig: %v", err)
		}
		cfg.GitHubAppConfig = &appCfg
		token, err := cfg.APITokenFromGitHubApp()
		if err != nil {
			return nil, fmt.Errorf("generating API token from GitHub App config: %v", err)
		}
		cfg.GithubToken = token
	}

	// Read token from file if not set directly
	if cfg.GithubToken == "" && cfg.GithubTokenFile != "" {
		tokenBytes, err := os.ReadFile(cfg.GithubTokenFile)
		if err != nil {
			return nil, fmt.Errorf("reading GitHub token from file: %v", err)
		}
		cfg.GithubToken = strings.TrimSpace(string(tokenBytes))
	}

	return &cfg, nil
}

func (c *Config) TargetURLs() []string {

	var urls []string

	opts := map[string]string{"per_page": "100"} // Used to set the Github API to return 100 results per page (max)

	if len(c.Repositories) == 0 && len(c.Organisations) == 0 && len(c.Users) == 0 {
		log.Info("No targets specified. Only rate limit endpoint will be scraped")
	}

	// Append repositories to the array
	if len(c.Repositories) > 0 {
		for _, x := range c.Repositories {
			y := *c.ApiUrl
			y.Path = path.Join(y.Path, "repos", x)
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	// Append GitHub organisations to the array
	if len(c.Organisations) > 0 {
		for _, x := range c.Organisations {
			y := *c.ApiUrl
			y.Path = path.Join(y.Path, "orgs", x, "repos")
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	if len(c.Users) > 0 {
		for _, x := range c.Users {
			y := *c.ApiUrl
			y.Path = path.Join(y.Path, "users", x, "repos")
			q := y.Query()
			for k, v := range opts {
				q.Add(k, v)
			}
			y.RawQuery = q.Encode()
			urls = append(urls, y.String())
		}
	}

	return urls
}

// APITokenFromGitHubApp generating api token from github app configuration.
func (c *Config) APITokenFromGitHubApp() (string, error) {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, c.GitHubAppId, c.GitHubAppInstallationId, c.GitHubAppKeyPath)
	if err != nil {
		return "", err
	}

	strToken, err := itr.Token(context.Background())
	if err != nil {
		return "", err
	}

	return strToken, nil
}

func (c *Config) GetClient() (*github.Client, error) {
	var httpClient *http.Client

	// Add custom transport for GitHub App authentication if enabled
	if c.GitHubApp {
		itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, c.GitHubAppId, c.GitHubAppInstallationId, c.GitHubAppKeyPath)
		if err != nil {
			return nil, fmt.Errorf("creating GitHub App installation transport: %v", err)
		}

		httpClient = &http.Client{Transport: itr}
	}

	client := github.NewClient(httpClient)

	if c.GithubToken != "" {
		client = client.WithAuthToken(c.GithubToken)
	}

	return client, nil
}

// mapSlice applies a function to each element of a slice and returns a new slice with the results.
func mapSlice[T any, M any](input []T, f func(T) M) []M {
	result := make([]M, len(input))
	for i, e := range input {
		result[i] = f(e)
	}
	return result
}
