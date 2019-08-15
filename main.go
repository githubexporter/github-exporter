package main

import (
	"net/http"

	"github.com/fatih/structs"
	conf "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	"github.com/infinityworks/go-common/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	log = logger.Start(&applicationCfg)
}

func main() {

	log.WithFields(structs.Map(applicationCfg)).Info("Starting Exporter")

	exporter := exporter.Exporter{
		APIMetrics: mets,
		Config:     applicationCfg,
	}

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(&exporter)

	// Setup HTTP handler
	http.Handle(applicationCfg.MetricsPath(), promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		                <head><title>Github Exporter</title></head>
		                <body>
		                   <h1>GitHub Prometheus Metrics Exporter</h1>
						   <p>For more information, visit <a href=https://github.com/infinityworks/github-exporter>GitHub</a></p>
		                   <p><a href='` + applicationCfg.MetricsPath() + `'>Metrics</a></p>
		                   </body>
		                </html>
		              `))
	})
	log.Fatal(http.ListenAndServe(":"+applicationCfg.ListenPort(), nil))
}
