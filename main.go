package main

import (
	conf "github.com/benri-io/jira-exporter/config"
	"github.com/benri-io/jira-exporter/exporter"
	"github.com/benri-io/jira-exporter/http"
	"github.com/benri-io/jira-exporter/logger"
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
	log = logger.Start(&applicationCfg)
	logger.SetDefaultLogger(log)
	mets = exporter.AddMetrics()
}

func main() {
	log.Info("Starting Exporter")

	exp := exporter.Exporter{
		APIMetrics: mets,
		Config:     applicationCfg,
	}

	var done = make(chan struct{}, 1)
	go func() {
		http.NewServer(exp).Start()
		done <- struct{}{}
	}()

	ch := make(chan prometheus.Metric)
	exp.Collect(ch)
	logrus.Infof("Done collecting data")
	<-done
}
