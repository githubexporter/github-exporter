package config

import (
	"context"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/kelseyhightower/envconfig"
	"net/http"
	"net/url"
	"strings"

	"os"
)

// Config struct holds runtime configuration required for the application
type Config struct {
	MetricsPath      string   `envconfig:"METRICS_PATH" required:"false" default:"/metrics"`
	ListenPort       string   `envconfig:"LISTEN_PORT" required:"false" default:"9171"`
	LogLevel         string   `envconfig:"LOG_LEVEL" required:"false" default:"INFO"`
	LogFormat        string   `envconfig:"LOG_FORMAT" required:"false" default:"json"`
	AppName          string   `envconfig:"APP_NAME" required:"false" default:"github-exporter"`
	ApiUrl           *url.URL `envconfig:"API_URL" required:"false" default:"https://api.github.com"`
	Repositories     []string `envconfig:"REPOS" required:"false" default:""`
	Organisations    []string `envconfig:"ORGS" required:"false" default:""`
	Users            []string `envconfig:"USERS" required:"false" default:""`
	GithubToken      string   `envconfig:"GITHUB_TOKEN" required:"false" default:""`
	GithubTokenFile  string   `envconfig:"GITHUB_TOKEN_FILE" required:"false" default:""`
	GitHubApp        bool     `envconfig:"GITHUB_APP" required:"false" default:"false"`
	TargetURLs       []string
	*GitHubAppConfig `ignored:"true"`
}

type GitHubAppConfig struct {
	GitHubAppKeyPath        string  `envconfig:"GITHUB_APP_KEY_PATH" required:"false" default:""`
	GitHubAppId             int64   `envconfig:"GITHUB_APP_ID" required:"false" default:""`
	GitHubAppInstallationId int64   `envconfig:"GITHUB_APP_INSTALLATION_ID" required:"false" default:""`
	GitHubRateLimit         float64 `envconfig:"GITHUB_RATE_LIMIT" required:"false" default:"15000"`
}

// Init populates the Config struct based on environmental runtime configuration
func Init(ctx context.Context) (*Config, error) {
	cfg := Config{}
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("processing config: %w", err)
	}

	ghAppCfg := GitHubAppConfig{}
	if cfg.GitHubApp {
		err = envconfig.Process("", &ghAppCfg)
		if err != nil {
			return nil, fmt.Errorf("processing github app config: %w", err)
		}
		cfg.GitHubAppConfig = &ghAppCfg

		token, err := getAPITokenFromGitHubApp(ctx, ghAppCfg.GitHubAppId, ghAppCfg.GitHubAppInstallationId, ghAppCfg.GitHubAppKeyPath)
		if err != nil {
			return nil, fmt.Errorf("getting api token from github app config: %w", err)
		}
		cfg.GithubToken = token

	} else if cfg.GithubTokenFile != "" {
		token, err := getAPITokenFromFile(cfg.GithubTokenFile)
		if err != nil {
			return nil, fmt.Errorf("setting api token from file: %w", err)
		}
		cfg.GithubToken = token
	}

	// TODO - validate comma-separated inputs e.g. orgs
	// TODO - improve and document token behaviour
	// TODO - set scrape URLs (or use gh package?)
	// TODO - validate API URL
	return &cfg, nil
}

// getAPITokenFromGitHubApp generating api token from github app configuration.
func getAPITokenFromGitHubApp(ctx context.Context, gitHubAppId int64, gitHubAppInstallationId int64, gitHubAppKeyPath string) (string, error) {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, gitHubAppId, gitHubAppInstallationId, gitHubAppKeyPath)
	if err != nil {
		return "", err
	}

	// TODO - pass context from main?
	// TODO - refresh behaviour
	token, err := itr.Token(ctx)
	if err != nil {
		return "", err
	}
	return token, nil
}

// getAPITokenFromFile accepts a file containing an oauth2 token for usage in http.request
func getAPITokenFromFile(tokenFile string) (string, error) {
	b, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
