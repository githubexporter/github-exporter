package test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/infinityworks/github-exporter/config"
	"github.com/infinityworks/github-exporter/exporter"
	web "github.com/infinityworks/github-exporter/http"
	"github.com/infinityworks/go-common/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/steinfletcher/apitest"
)

var (
	log *logrus.Logger
)

func TestHomepage(t *testing.T) {
	test, collector := apiTest(withConfig("a/b"))
	defer prometheus.Unregister(&collector)

	test.Get("/").
		Expect(t).
		Assert(bodyContains("GitHub Prometheus Metrics Exporter")).
		Status(http.StatusOK).
		End()
}

func apiTest(conf config.Config) (*apitest.APITest, exporter.Exporter) {

	log = logger.Start(conf.Config)

	exp := exporter.New(conf, log)

	server := web.NewServer(exp)

	return apitest.New().
		Report(apitest.SequenceDiagram()).
		Handler(server.Handler), *exp
}

func withConfig(repos string) config.Config {
	_ = os.Setenv("REPOS", repos)
	_ = os.Setenv("GITHUB_TOKEN", "12345")
	return config.Init()
}

func bodyContains(substr string) func(*http.Response, *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		response := string(bytes)
		if !strings.Contains(response, substr) {
			return fmt.Errorf("response did not contain substring '%s'", substr)
		}
		return nil
	}
}
