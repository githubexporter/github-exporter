package exporter

import (
	"context"
	"github.com/githubexporter/github-exporter/internal/config"
	"github.com/githubexporter/github-exporter/internal/logging"
	"github.com/google/go-github/v59/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestGetRateLimits(t *testing.T) {
	limit := 5000
	remaining := 4000
	reset := github.Timestamp{
		Time: time.Now(),
	}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetRateLimit,
			github.RateLimits{
				Core: &github.Rate{
					Limit:     limit,
					Remaining: remaining,
					Reset:     reset,
				},
			},
		),
	)

	ctx := context.Background()
	cfg, err := config.Init(ctx)
	require.NoError(t, err)

	logger, err := logging.New(cfg, os.Stdout)
	require.NoError(t, err)

	metrics := GetMetrics()
	githubClient := github.NewClient(mockedHTTPClient)
	exp := NewExporter(ctx, logger, cfg, githubClient, metrics)

	rateLimits, err := exp.getRateLimits()

	t.Log(rateLimits)
	assert.Equal(t, remaining, rateLimits.Core.Remaining)
	assert.Equal(t, reset, rateLimits.Core.Reset)
}

func TestGetRepos(t *testing.T) {
	t.Skip("Not implemented")
}

func TestGetReleases(t *testing.T) {
	t.Skip("Not implemented")
}
