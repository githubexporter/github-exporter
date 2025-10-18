package main

import (
	"github.com/githubexporter/github-exporter/config"
	"github.com/githubexporter/github-exporter/exporter"
	"github.com/githubexporter/github-exporter/http"
	
	"github.com/google/go-github/v76/github"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting Exporter")

	applicationCfg, err := config.Init()
	if err != nil {
		log.Fatalf("Error initializing configuration: %v", err)
	}

	metrics := exporter.AddMetrics()

	exp := exporter.Exporter{
		APIMetrics: metrics,
		Config:     *applicationCfg,
		Client:     github.NewClient(nil),
	}

	http.NewServer(exp).Start()
}
