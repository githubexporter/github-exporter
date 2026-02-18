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

	"github.com/google/go-github/v76/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/steinfletcher/apitest"
)

func TestHomepage(t *testing.T) {
	test, collector := apiTest(withConfig())
	defer prometheus.Unregister(&collector)

	test.Get("/").
		Expect(t).
		Assert(bodyContains("GitHub Prometheus Metrics Exporter")).
		Status(http.StatusOK).
		End()
}

func TestGithubExporter(t *testing.T) {
	test, collector := apiTest(withConfig())
	defer prometheus.Unregister(&collector)

	test.Mocks(
		githubRepos(),
		githubReleases(),
		githubPulls(),
		githubUserRepos(),
		githubUserReleases(),
		githubUserPulls(),
		githubOrgRepos(),
		githubReleases(),
		githubPulls(),
		githubRateLimit(),
	).
		Get("/metrics").
		Expect(t).
		Assert(bodyContains(`github_rate_limit{resource="code_search"} 60`)).
		Assert(bodyContains(`github_rate_limit{resource="core"} 60`)).
		Assert(bodyContains(`github_rate_limit{resource="graphql"} 0`)).
		Assert(bodyContains(`github_rate_limit{resource="integration_manifest"} 5000`)).
		Assert(bodyContains(`github_rate_limit{resource="search"} 10`)).
		Assert(bodyContains(`github_rate_remaining{resource="code_search"} 60`)).
		Assert(bodyContains(`github_rate_remaining{resource="core"} 60`)).
		Assert(bodyContains(`github_rate_remaining{resource="graphql"} 0`)).
		Assert(bodyContains(`github_rate_remaining{resource="integration_manifest"} 5000`)).
		Assert(bodyContains(`github_rate_remaining{resource="search"} 10`)).
		Assert(bodyContains(`github_rate_reset{resource="code_search"} 3e+09`)).
		Assert(bodyContains(`github_rate_reset{resource="core"} 3e+09`)).
		Assert(bodyContains(`github_rate_reset{resource="graphql"} 3e+09`)).
		Assert(bodyContains(`github_rate_reset{resource="integration_manifest"} 3e+09`)).
		Assert(bodyContains(`github_rate_reset{resource="search"} 3e+09`)).
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
		Assert(bodyContains(`github_repo_forks{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myUser"} 10`)).
		Assert(bodyContains(`github_repo_pull_request_count{repo="myRepo",user="myUser"} 3`)).
		Assert(bodyContains(`github_repo_open_issues{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myUser"} 2`)).
		Assert(bodyContains(`github_repo_size_kb{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myUser"} 946`)).
		Assert(bodyContains(`github_repo_stars{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myUser"} 120`)).
		Assert(bodyContains(`github_repo_watchers{archived="false",fork="false",language="Go",license="mit",private="false",repo="myRepo",user="myUser"} 5`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-02-28T08:25:53Z",name="myRepo_1.3.0_checksums.txt",release="1.3.0",repo="myRepo",tag="1.3.0",user="myUser"} 7292`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-02-28T08:25:53Z",name="myRepo_1.3.0_windows_amd64.tar.gz",release="1.3.0",repo="myRepo",tag="1.3.0",user="myUser"} 21`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-05-02T15:22:16Z",name="myRepo_2.0.0_checksums.txt",release="2.0.0",repo="myRepo",tag="2.0.0",user="myUser"} 14564`)).
		Assert(bodyContains(`github_repo_release_downloads{created_at="2019-05-02T15:22:16Z",name="myRepo_2.0.0_windows_amd64.tar.gz",release="2.0.0",repo="myRepo",tag="2.0.0",user="myUser"} 55`)).
		Status(http.StatusOK).
		End()
}

func TestGithubExporterHttpErrorHandling(t *testing.T) {
	test, collector := apiTest(withConfig())
	defer prometheus.Unregister(&collector)

	// Test that the exporter returns when an error occurs
	// Ideally a new gauge should be added to keep track of scrape errors
	// following prometheus exporter guidelines
	test.Mocks(
		githubRepos(),
		githubReleases(),
		githubPullsError(),
	).
		Get("/metrics").
		Expect(t).
		Status(http.StatusOK).
		End()
}

func apiTest(conf config.Config) (*apitest.APITest, exporter.Exporter) {
	exp := exporter.Exporter{
		APIMetrics: exporter.AddMetrics(&conf),
		Config:     conf,
		Client:     github.NewClient(nil),
	}
	server := web.NewServer(exp)

	return apitest.New().
		Report(apitest.SequenceDiagram()).
		Handler(server.Handler), exp
}

func withConfig() config.Config {
	_ = os.Setenv("REPOS", "myOrg/myRepo")
	_ = os.Setenv("ORGS", "myOrg")
	_ = os.Setenv("USERS", "myUser")

	_ = os.Setenv("GITHUB_TOKEN", "12345")
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}
	return *cfg
}

func githubRepos() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo").
		RespondWith().
		Times(1).
		Body(readFile("testdata/my_repo_response.json")).
		Status(200).
		End()
}

func githubUserRepos() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/users/myUser/repos").
		RespondWith().
		Times(1).
		Body(readFile("testdata/user_repos_response.json")).
		Status(200).
		End()
}

func githubUserReleases() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myUser/myRepo/releases").
		RespondWith().
		Times(1).
		Body(readFile("testdata/user_releases_response.json")).
		Status(200).
		End()
}

func githubUserPulls() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myUser/myRepo/pulls").
		RespondWith().
		Times(1).
		Body(readFile("testdata/user_pulls_response.json")).
		Status(http.StatusOK).
		End()
}

func githubOrgRepos() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/orgs/myOrg/repos").
		RespondWith().
		Times(1).
		Body(readFile("testdata/org_repos_response.json")).
		Status(200).
		End()
}

func githubRateLimit() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/rate_limit").
		RespondWith().
		Times(1).
		Body(readFile("testdata/rate_limit_response.json")).
		Status(http.StatusOK).
		End()
}

func githubReleases() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo/releases").
		RespondWith().
		Times(1).
		Body(readFile("testdata/releases_response.json")).
		Status(http.StatusOK).
		End()
}

func githubPulls() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo/pulls").
		RespondWith().
		Times(1).
		Body(readFile("testdata/pulls_response.json")).
		Status(http.StatusOK).
		End()
}

func githubPullsError() *apitest.Mock {
	return apitest.NewMock().
		Get("https://api.github.com/repos/myOrg/myRepo/pulls").
		RespondWith().
		Times(1).
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
			return fmt.Errorf("response did not contain substring '%s'", response)
		}
		return nil
	}
}
