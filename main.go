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
	log = logger.Start(applicationCfg.Config)
}

func main() {
	log.WithFields(structs.Map(applicationCfg)).Info("Starting Exporter")

	exporter := exporter.New(applicationCfg, log)

	http.NewServer(exporter).Start()
}
