package http

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/githubexporter/github-exporter/internal/exporter"
)

type Server struct {
	logger   *slog.Logger
	handler  http.Handler
	exporter exporter.Exporter
}

func NewServer(exporter exporter.Exporter) *Server {
	r := http.NewServeMux()

	// Register Metrics from each of the endpoints
	// This invokes the Collect method through the prometheus client libraries.
	prometheus.MustRegister(&exporter)

	r.Handle(exporter.Config.MetricsPath, promhttp.Handler())
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
<head><title>Github Exporter</title></head>
<body>
<h1>GitHub Prometheus Metrics Exporter</h1>
<p>For more information, visit <a href=https://github.com/githubexporter/github-exporter>GitHub</a></p>
<p><a href='` + exporter.Config.MetricsPath + `'>Metrics</a></p>
</body>
</html>
		`))
	})

	return &Server{handler: r, exporter: exporter}
}

func (s *Server) Run() {
	err := http.ListenAndServe(":"+s.exporter.Config.ListenPort, s.handler)
	s.logger.Error("http server error", "error", err.Error())
}
