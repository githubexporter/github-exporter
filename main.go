package main

import (
	conf "github.com/githubexporter/github-exporter/config"
	"github.com/githubexporter/github-exporter/exporter"
	"github.com/githubexporter/github-exporter/http"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting Exporter")

	applicationCfg, err := conf.Init()
	if err != nil {
		log.Fatalf("Error initializing configuration: %v", err)
	}

	metrics := exporter.AddMetrics()

	exp := exporter.Exporter{
		APIMetrics: metrics,
		Config:     *applicationCfg,
	}

	http.NewServer(exp).Start()
}
