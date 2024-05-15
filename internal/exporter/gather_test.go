package exporter

import (
	"context"
	"github.com/githubexporter/github-exporter/internal/config"
	"github.com/githubexporter/github-exporter/internal/logging"
	"github.com/google/go-github/v61/github"
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
		Time: time.Now().Round(time.Second),
	}
	rateLimits := github.RateLimits{
		Core: &github.Rate{
			Limit:     limit,
			Remaining: remaining,
			Reset:     reset,
		},
	}

	response := new(struct {
		Resources *github.RateLimits `json:"resources"`
	})

	response.Resources = &rateLimits

	exp := getMockedExporter(mock.GetRateLimit, response)
	res, err := exp.getRateLimits()
	require.NoError(t, err)

	assert.Equal(t, remaining, res.Core.Remaining)
	assert.Equal(t, limit, res.Core.Limit)
	assert.Equal(t, reset, res.Core.Reset)

}

func TestGetRepos(t *testing.T) {
	t.Skip("Not implemented")
}

func TestGetReleases(t *testing.T) {
	t.Skip("Not implemented")
}

// TODO - make more testy. abstract elsewhere?
func getMockedExporter(ep mock.EndpointPattern, response interface{}) Exporter {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			ep,
			response,
		),
	)

	ctx := context.Background()
	cfg, err := config.Init(ctx)
	if err != nil {
		panic(err)
	}

	logger, err := logging.New(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	if err != nil {
		panic(err)
	}

	metrics := GetMetrics()
	githubClient := github.NewClient(mockedHTTPClient)
	return NewExporter(ctx, logger, cfg, githubClient, metrics)
}
