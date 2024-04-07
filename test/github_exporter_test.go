package test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/githubexporter/github-exporter/config"
	"github.com/githubexporter/github-exporter/exporter"
	web "github.com/githubexporter/github-exporter/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinfletcher/apitest"
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

func TestGithubExporter(t *testing.T) {
	test, collector := apiTest(withConfig("myOrg/myRepo"))
	defer prometheus.Unregister(&collector)

	test.Mocks(
		githubRepos(),
		githubRateLimit(),
		githubReleases(),
		githubPulls(),
	).
		Get("/metrics").
		Expect(t).
		Assert(bodyContains(`github_rate_limit 60`)).
		Assert(bodyContains(`github_rate_remaining 60`)).
		Assert(bodyContains(`github_rate_reset 1.566853865e+09`)).
		Assert(bodyContains(`github_repo_forks{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myOrg"} 10`)).
		Assert(bodyContains(`github_repo_pull_request_count{repo="myRepo",user="myOrg"} 3`)).
		Assert(bodyContains(`github_repo_open_issues{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myOrg"} 2`)).
		Assert(bodyContains(`github_repo_size_kb{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myOrg"} 946`)).
		Assert(bodyContains(`github_repo_stars{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myOrg"} 120`)).
		Assert(bodyContains(`github_repo_watchers{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myOrg"} 5`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-02-28T08:25:53Z",name="myRepo_1.3.0_checksums.txt",release="1.3.0",repo="myRepo",tag="1.3.0",user="myOrg"} 7292`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-02-28T08:25:53Z",name="myRepo_1.3.0_windows_amd64.tar.gz",release="1.3.0",repo="myRepo",tag="1.3.0",user="myOrg"} 21`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-05-02T15:22:16Z",name="myRepo_2.0.0_checksums.txt",release="2.0.0",repo="myRepo",tag="2.0.0",user="myOrg"} 14564`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-05-02T15:22:16Z",name="myRepo_2.0.0_windows_amd64.tar.gz",release="2.0.0",repo="myRepo",tag="2.0.0",user="myOrg"} 55`)).
		Status(http.StatusOK).
		End()
}

func TestGithubExporterHttpErrorHandling(t *testing.T) {
	test, collector := apiTest(withConfig("myOrg/myRepo"))
	defer prometheus.Unregister(&collector)

	// Test that the exporter returns when an error occurs
	// Ideally a new gauge should be added to keep track of scrape errors
	// following prometheus exporter guidelines
	test.Mocks(
		githubPullsError(),
	).
		Get("/metrics").
		Expect(t).
		Status(http.StatusOK).
		End()
}

func apiTest(conf config.Config) (*apitest.APITest, exporter.Exporter) {
	exp := exporter.Exporter{
		APIMetrics: exporter.AddMetrics(),
		Config:     conf,
	}
	server := web.NewServer(exp)

	return apitest.New().
		Report(apitest.SequenceDiagram()).
		Handler(server.Handler), exp
}

func withConfig(repos string) config.Config {
	_ = os.Setenv("REPOS", repos)
	_ = os.Setenv("GITHUB_TOKEN", "12345")
	return config.Init()
}

func githubRepos() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo").
		Header("Authorization", "token 12345").
		Query("per_page", "100").
		RespondWith().
		Times(2).
		Body(readFile("testdata/my_repo_response.json")).
		Status(200).
		End()
}

func githubRateLimit() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/rate_limit").
		Header("Authorization", "token 12345").
		RespondWith().
		Header("X-RateLimit-Limit", "60").
		Header("X-RateLimit-Remaining", "60").
		Header("X-RateLimit-Reset", "1566853865").
		Status(http.StatusOK).
		End()
}

func githubReleases() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo/releases").
		Header("Authorization", "token 12345").
		RespondWith().
		Times(2).
		Body(readFile("testdata/releases_response.json")).
		Status(http.StatusOK).
		End()
}

func githubPulls() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo/pulls").
		Header("Authorization", "token 12345").
		RespondWith().
		Times(2).
		Body(readFile("testdata/pulls_response.json")).
		Status(http.StatusOK).
		End()
}

func githubPullsError() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo/pulls").
		Header("Authorization", "token 12345").
		RespondWith().
		Status(http.StatusBadRequest).
		End()
}

func readFile(path string) string {
	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func bodyContains(substr string) func(*http.Response, *http.Request) error {
	return func(res *http.Response, req *http.Request) error {
		bytes, err := io.ReadAll(res.Body)
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
