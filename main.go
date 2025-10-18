package main

import (
	"github.com/githubexporter/github-exporter/config"
	"github.com/githubexporter/github-exporter/exporter"
	"github.com/githubexporter/github-exporter/http"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Starting Exporter")

	applicationCfg, err := config.Init()
	if err != nil {
		log.Fatalf("Error initializing configuration: %v", err)
	}

	exp, err := exporter.NewExporter(applicationCfg)
	if err != nil {
		log.Fatalf("Error initializing exporter: %v", err)
	}

	http.NewServer(*exp).Start()
}
