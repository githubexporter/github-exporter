package main

import "C"
import (
	"context"
	"github.com/fatih/structs"
	conf "github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shurcooL/githubv4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"net/http"
)

var (
	c    conf.Config
	mets map[string]*prometheus.Desc
)

func init() {
	viper.AutomaticEnv()
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(conf.LogLevel())

	c = conf.Init()
	mets = exporter.AddMetrics()
}

func main() {

	log.WithFields(structs.Map(c)).Info("Starting Exporter")

	applicationCfg := exporter.Exporter{
		APIMetrics: mets,
		Config:     c,
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: viper.GetString("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)
	exporter.Query(client)

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(&applicationCfg)

	// Setup HTTP handler
	//http.Handle(c.MetricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(
			`<html>
				<head><title>Github Exporter</title></head>
				<body>
					<h1>GitHub Prometheus Metrics Exporter</h1>
					<p>For more information, visit <a href=https://github.com/infinityworks/github-exporter>GitHub</a></p>
					<p><a href='` + c.MetricsPath + `'>Metrics</a></p>
				</body>
			</html>`))
	})
	log.Fatal(http.ListenAndServe(":"+c.ListenPort, nil))
}
