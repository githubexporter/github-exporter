package config

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedCfg *Config
		expectedErr error
	}{
		{
			name: "default config",
			expectedCfg: &Config{
				MetricsPath: "/metrics",
				ListenPort:  "9171",
				LogLevel:    "INFO",
				ApiUrl: &url.URL{
					Scheme: "https",
					Host:   "api.github.com",
				},
				Repositories:             []string{},
				Organisations:            []string{},
				Users:                    []string{},
				GitHubResultsPerPage:     100,
				GithubToken:              "",
				GithubTokenFile:          "",
				GitHubApp:                false,
				GitHubAppConfig:          nil,
				GitHubRateLimitEnabled:   true,
				GitHubRateLimit:          0,
				FetchRepoReleasesEnabled: true,
				FetchOrgsConcurrency:     1,
				FetchOrgReposConcurrency: 1,
				FetchReposConcurrency:    1,
				FetchUsersConcurrency:    1,
			},
			expectedErr: nil,
		},
		{
			name: "non-default config",
			envVars: map[string]string{
				"METRICS_PATH":                "/otherendpoint",
				"LISTEN_PORT":                 "1111",
				"LOG_LEVEL":                   "DEBUG",
				"API_URL":                     "https://example.com",
				"REPOS":                       "repo1, repo2",
				"ORGS":                        "org1,org2 ",
				"USERS":                       " user1, user2 ",
				"GITHUB_RESULTS_PER_PAGE":     "50",
				"GITHUB_TOKEN":                "token",
				"GITHUB_RATE_LIMIT_ENABLED":   "false",
				"FETCH_REPO_RELEASES_ENABLED": "false",
				"FETCH_ORGS_CONCURRENCY":      "2",
				"FETCH_ORG_REPOS_CONCURRENCY": "3",
				"FETCH_REPOS_CONCURRENCY":     "4",
				"FETCH_USERS_CONCURRENCY":     "5",
			},
			expectedCfg: &Config{
				MetricsPath: "/otherendpoint",
				ListenPort:  "1111",
				LogLevel:    "DEBUG",
				ApiUrl: &url.URL{
					Scheme: "https",
					Host:   "example.com",
				},
				Repositories: []string{
					"repo1",
					"repo2",
				},
				Organisations: []string{
					"org1",
					"org2",
				},
				Users: []string{
					"user1",
					"user2",
				},
				GitHubResultsPerPage:     50,
				GithubToken:              "token",
				GithubTokenFile:          "",
				GitHubApp:                false,
				GitHubAppConfig:          nil,
				GitHubRateLimitEnabled:   false,
				GitHubRateLimit:          0,
				FetchRepoReleasesEnabled: false,
				FetchOrgsConcurrency:     2,
				FetchOrgReposConcurrency: 3,
				FetchReposConcurrency:    4,
				FetchUsersConcurrency:    5,
			},
			expectedErr: nil,
		},
		{
			name: "github rate limit threshold config",
			envVars: map[string]string{
				"GITHUB_RATE_LIMIT": "15000",
			},
			expectedCfg: &Config{
				MetricsPath: "/metrics",
				ListenPort:  "9171",
				LogLevel:    "INFO",
				ApiUrl: &url.URL{
					Scheme: "https",
					Host:   "api.github.com",
				},
				Repositories:             []string{},
				Organisations:            []string{},
				Users:                    []string{},
				GitHubResultsPerPage:     100,
				GithubToken:              "",
				GithubTokenFile:          "",
				GitHubApp:                false,
				GitHubAppConfig:          nil,
				GitHubRateLimitEnabled:   true,
				GitHubRateLimit:          15000,
				FetchRepoReleasesEnabled: true,
				FetchOrgsConcurrency:     1,
				FetchOrgReposConcurrency: 1,
				FetchReposConcurrency:    1,
				FetchUsersConcurrency:    1,
			},
			expectedErr: nil,
		},
		{
			name:        "invalid url",
			expectedCfg: nil,
			envVars: map[string]string{
				"API_URL": "://invalid-url",
			},
			expectedErr: errors.New("processing envconfig: envconfig.Process: assigning API_URL to ApiUrl: converting '://invalid-url' to type url.URL. details: parse \"://invalid-url\": missing protocol scheme"),
		},
		{
			name:        "invalid github app id",
			expectedCfg: nil,
			envVars: map[string]string{
				"GITHUB_APP":    "true",
				"GITHUB_APP_ID": "not-an-integer",
			},
			expectedErr: errors.New("processing GitHub App envconfig: envconfig.Process: assigning GITHUB_APP_ID to GitHubAppId: converting 'not-an-integer' to type int64. details: strconv.ParseInt: parsing \"not-an-integer\": invalid syntax"),
		},
		{
			name:        "invalid log level",
			expectedCfg: nil,
			envVars: map[string]string{
				"LOG_LEVEL": "boop",
			},
			expectedErr: errors.New("parsing log level: not a valid logrus Level: \"boop\""),
		},
		{
			name: "github token file not found",
			envVars: map[string]string{
				"GITHUB_TOKEN_FILE": "/non/existent/file",
			},
			expectedCfg: nil,
			expectedErr: errors.New("reading GitHub token from file: open /non/existent/file: no such file or directory"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := Init()

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedCfg, cfg)
		})
	}
}

func TestGetClient(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		expectedHost string
		expectedPath string
		expectedErr  error
	}{
		{
			name:         "default config",
			expectedHost: "api.github.com",
			expectedPath: "/",
			expectedErr:  nil,
		},
		{
			name: "non-default config",
			envVars: map[string]string{
				"METRICS_PATH":                "/otherendpoint",
				"LISTEN_PORT":                 "1111",
				"LOG_LEVEL":                   "DEBUG",
				"API_URL":                     "https://example.com",
				"REPOS":                       "repo1, repo2",
				"ORGS":                        "org1,org2 ",
				"USERS":                       " user1, user2 ",
				"GITHUB_RESULTS_PER_PAGE":     "50",
				"GITHUB_TOKEN":                "token",
				"GITHUB_RATE_LIMIT_ENABLED":   "false",
				"FETCH_REPO_RELEASES_ENABLED": "false",
				"FETCH_ORGS_CONCURRENCY":      "2",
				"FETCH_ORG_REPOS_CONCURRENCY": "3",
				"FETCH_REPOS_CONCURRENCY":     "4",
				"FETCH_USERS_CONCURRENCY":     "5",
			},
			expectedHost: "example.com",
			expectedPath: "/api/v3/",
			expectedErr:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, _ := Init()
			client, _ := cfg.GetClient()

			assert.Equal(t, client.BaseURL.Host, tt.expectedHost)
			assert.Equal(t, client.BaseURL.Path, tt.expectedPath)
		})
	}
}
