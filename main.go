package main

import (
	"github.com/fatih/structs"
	conf "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	"github.com/infinityworks/github-exporter/http"
	"github.com/infinityworks/go-common/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	log            *logrus.Logger
	applicationCfg conf.Config
	mets           map[string]*prometheus.Desc
)

func init() {
	applicationCfg = conf.Init()
	mets = exporter.AddMetrics()
	log = logger.Start(applicationCfg.Config)
}

func main() {
	log.WithFields(structs.Map(applicationCfg)).Info("Starting Exporter")

	exp := exporter.Exporter{
		Metrics: mets,
		Config:  applicationCfg,
		Log:     log,
	}

	http.NewServer(exp).Start()
}
