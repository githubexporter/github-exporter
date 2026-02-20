package exporter

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/githubexporter/github-exporter/config"
	"github.com/google/go-github/v76/github"
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
