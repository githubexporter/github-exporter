package main

import (
	"context"
	"github.com/githubexporter/github-exporter/internal/config"
	"github.com/githubexporter/github-exporter/internal/exporter"
	"github.com/githubexporter/github-exporter/internal/http"
	"github.com/githubexporter/github-exporter/internal/logging"
	"github.com/google/go-github/v61/github"
	"os"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Init(ctx)
	if err != nil {
		panic(err)
	}

	logger, err := logging.New(cfg.LogLevel, cfg.LogFormat, os.Stdout)
	if err != nil {
		panic(err)
	}

	metrics := exporter.GetMetrics()
	githubClient := github.NewClient(nil)
	exp := exporter.NewExporter(ctx, logger, cfg, githubClient, metrics)

	logger.Info("Starting Exporter")
	server := http.NewServer(exp)
	server.Run()
}
