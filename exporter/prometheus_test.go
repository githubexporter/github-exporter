package exporter

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/githubexporter/github-exporter/config"
	"github.com/google/go-github/v76/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTempRSAKey generates a 2048-bit RSA private key and writes it as a
// PKCS#1 PEM file to a temp path. The file is removed automatically when the
// test finishes.
func writeTempRSAKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "test-github-app-key-*.pem")
	require.NoError(t, err)
	defer f.Close()

	err = pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	require.NoError(t, err)

	t.Cleanup(func() { os.Remove(f.Name()) })

	return f.Name()
}

func newTestExporter(cfg config.Config) *Exporter {
	return &Exporter{
		Client: github.NewClient(nil),
		Config: cfg,
	}
}

// TestCollectDoesNotPanic checks that a scrape ends cleanly, with no metrics,
// when a GitHub API response can't be used. A panic in Collect would crash the
// exporter.
func TestCollectDoesNotPanic(t *testing.T) {
	const repo = `{"name":"myRepo","owner":{"login":"myOrg"}}`
	const forbidden = `{"message":"Resource not accessible by personal access token"}`

	responses := map[string]string{
		"/repos/myOrg/myRepo":          repo,
		"/users/myOrg/repos":           "[" + repo + "]",
		"/orgs/myOrg/repos":            "[" + repo + "]",
		"/repos/myOrg/myRepo/releases": "[]",
		"/repos/myOrg/myRepo/pulls":    "[]",
		"/rate_limit":                  `{"resources":{"core":{"limit":1,"remaining":1,"reset":1}}}`,
	}

	tests := []struct {
		name   string
		path   string
		status int
		body   string
	}{
		{"repository forbidden", "/repos/myOrg/myRepo", http.StatusForbidden, forbidden},
		{"user repositories forbidden", "/users/myOrg/repos", http.StatusForbidden, forbidden},
		{"organisation repositories forbidden", "/orgs/myOrg/repos", http.StatusForbidden, forbidden},
		{"releases forbidden", "/repos/myOrg/myRepo/releases", http.StatusForbidden, forbidden},
		{"pull requests forbidden", "/repos/myOrg/myRepo/pulls", http.StatusForbidden, forbidden},
		{"rate limit forbidden", "/rate_limit", http.StatusForbidden, forbidden},
		{"rate limit without resources", "/rate_limit", http.StatusOK, `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tt.path {
					w.WriteHeader(tt.status)
					fmt.Fprint(w, tt.body)
					return
				}
				fmt.Fprint(w, responses[r.URL.Path])
			}))
			defer server.Close()

			baseURL, err := url.Parse(server.URL + "/")
			require.NoError(t, err)

			cfg := config.Config{
				Repositories:             []string{"myOrg/myRepo"},
				Organisations:            []string{"myOrg"},
				Users:                    []string{"myOrg"},
				GitHubRateLimitEnabled:   true,
				FetchRepoReleasesEnabled: true,
				FetchReposConcurrency:    1,
				FetchOrgsConcurrency:     1,
				FetchOrgReposConcurrency: 1,
				FetchUsersConcurrency:    1,
			}
			e := newTestExporter(cfg)
			e.APIMetrics = AddMetrics(&cfg)
			e.Client.BaseURL = baseURL

			ch := make(chan prometheus.Metric, 100)
			assert.NotPanics(t, func() { e.Collect(ch) })
			assert.Empty(t, ch)
		})
	}
}

func TestRateLimitRefresh(t *testing.T) {
	// Shared rate-limit slices used across sub-tests.
	aboveThreshold := []RateLimit{{Resource: "core", Remaining: 200, Limit: 5000}}
	atThreshold := []RateLimit{{Resource: "core", Remaining: 100, Limit: 5000}}
	belowThreshold := []RateLimit{{Resource: "core", Remaining: 50, Limit: 5000}}
	noCoreEntry := []RateLimit{{Resource: "search", Remaining: 5, Limit: 30}}

	baseCfg := config.Config{
		GitHubResultsPerPage:     100,
		FetchReposConcurrency:    1,
		FetchOrgsConcurrency:     1,
		FetchOrgReposConcurrency: 1,
		FetchUsersConcurrency:    1,
	}

	t.Run("no-op when GitHubRateLimit is zero", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 0
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&atThreshold)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when GitHubRateLimit is negative", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = -1
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&atThreshold)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when rates is nil", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 100
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(nil)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when core remaining is above threshold", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 100 // remaining=200 > threshold=100
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&aboveThreshold)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when rates has no core entry", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 100
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&noCoreEntry)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when GitHubApp is false and core is at threshold", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 100 // remaining=100 <= threshold=100
		cfg.GitHubApp = false
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&atThreshold)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when GitHubApp is false and core is below threshold", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 100 // remaining=50 < threshold=100
		cfg.GitHubApp = false
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&belowThreshold)

		assert.Same(t, original, e.Client)
	})

	t.Run("no-op when GitHubApp is true but GetClient fails", func(t *testing.T) {
		cfg := baseCfg
		cfg.GitHubRateLimit = 100
		cfg.GitHubApp = true
		cfg.GitHubAppConfig = &config.GitHubAppConfig{
			GitHubAppKeyPath:        "/nonexistent/key.pem",
			GitHubAppId:             1,
			GitHubAppInstallationId: 1,
		}
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&atThreshold)

		assert.Same(t, original, e.Client)
	})

	t.Run("replaces client when GitHubApp is true and core is at threshold", func(t *testing.T) {
		keyPath := writeTempRSAKey(t)
		cfg := baseCfg
		cfg.GitHubRateLimit = 100 // remaining=100 <= threshold=100
		cfg.GitHubApp = true
		cfg.GitHubAppConfig = &config.GitHubAppConfig{
			GitHubAppKeyPath:        keyPath,
			GitHubAppId:             1,
			GitHubAppInstallationId: 1,
		}
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&atThreshold)

		assert.NotSame(t, original, e.Client, "expected client to be replaced after rate limit refresh")
	})

	t.Run("replaces client when GitHubApp is true and core is below threshold", func(t *testing.T) {
		keyPath := writeTempRSAKey(t)
		cfg := baseCfg
		cfg.GitHubRateLimit = 100 // remaining=50 < threshold=100
		cfg.GitHubApp = true
		cfg.GitHubAppConfig = &config.GitHubAppConfig{
			GitHubAppKeyPath:        keyPath,
			GitHubAppId:             1,
			GitHubAppInstallationId: 1,
		}
		e := newTestExporter(cfg)
		original := e.Client

		e.rateLimitRefresh(&belowThreshold)

		assert.NotSame(t, original, e.Client, "expected client to be replaced after rate limit refresh")
	})
}
