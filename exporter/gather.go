package exporter

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// newClient provides an authenticated Github
// client for use in our API interactions
func (e *Exporter) newClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: e.Config.APIToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// GatherRates returns the GitHub API rate limits
// A free call that doesn't count towards your usage
func (e *Exporter) gatherRates(client *github.Client) {
	limits, _, err := client.RateLimits(context.Background())
	if err != nil {
		e.Log.Errorf("Error gathering API Rate limits: %v", err)
	}

	e.RateLimits.Limit = float64(limits.Core.Limit)
	e.RateLimits.Remaining = float64(limits.Core.Remaining)
	e.RateLimits.Reset = float64(limits.Core.Reset.Unix())

	if e.RateLimits.Remaining == 0 {
		e.Log.Errorf("Error - Github API Rate limit exceeded, review rates and usage")
	}
}

func (e *Exporter) gatherByOrg(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Config.Organisations) == 0 {
		e.Log.Info("No Organisations specified, skipping collection")
		return
	}

	// Requests are limited so we get as many objects per page as possible
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Loop through the organizations
	for _, y := range e.Config.Organisations {

		// Skip any undefined orgs
		if y == "" {
			continue
		}

		e.Log.Infof("Gathering metrics for GitHub Org %s", y)

		// Support pagination
		var allRepos []*github.Repository

		for {
			repos, resp, err := client.Repositories.ListByOrg(context.Background(), y, opt)
			if err != nil {
				e.Log.Errorf("Error listing repositories by org: %v", err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		e.enrichMetrics(client, allRepos)

	}
}

func (e *Exporter) gatherByUser(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Config.Users) == 0 {
		e.Log.Info("No Users specified, skipping collection")
		return
	}

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Loop through the Users passed in
	for _, y := range e.Config.Users {

		// Skip any undefined users
		if y == "" {
			continue
		}

		e.Log.Info("Gathering metrics for GitHub User ", y)

		// Support pagination
		var allRepos []*github.Repository

		for {
			repos, resp, err := client.Repositories.List(context.Background(), y, opt)
			if err != nil {
				e.Log.Errorf("Error listing repositories by user: %v", err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		e.enrichMetrics(client, allRepos)

	}
}

func (e *Exporter) gatherByRepo(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Config.Repositories) == 0 {
		e.Log.Info("No individual repositories specified, skipping collection")
		return
	}

	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Loop through the Users passed in
	for _, y := range e.Config.Repositories {

		// Skip any undefined users
		if y == "" {
			continue
		}

		// Prepare the arguemtns for the get
		parts := strings.Split(y, "/")
		owner := parts[0]
		repo := parts[1]

		e.Log.Infof("Gathering metrics for GitHub Repo %s", y)

		// Support pagination
		var allRepos []*github.Repository

		for {
			// collect basic repository information for the repo
			metrics, resp, err := client.Repositories.Get(context.Background(), owner, repo)
			if err != nil {
				e.Log.Errorf("Error collecting repository metrics: %v", err)
			}

			allRepos = append(allRepos, metrics)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		// Enrich the metrics
		e.enrichMetrics(client, allRepos)

	}
}

func (e *Exporter) metricEnabled(option string) bool {

	for _, v := range e.Config.AdditionalMetrics {
		if v == option {
			return true
		}
	}

	return false
}

// Adds metrics not available in the standard repository response
// Also adds them to the metrics struct format for processing
func (e *Exporter) enrichMetrics(client *github.Client, repos []*github.Repository) {

	// Let's then range over the repositories fed to the struct
	for _, y := range repos {

		// TODO Stage a better word?
		// TODO - Fix pagination
		pulls := 0.0
		if e.metricEnabled("pulls") {
			p, _, err := client.PullRequests.List(context.Background(), *y.Owner.Login, *y.Name, nil)
			if err != nil {
				e.Log.Errorf("Error obtaining pull metrics: %v", err)
			}

			pulls = float64(len(p))
		}

		// TODO - Fix pagination
		releases := 0.0
		if e.metricEnabled("releases") {
			r, _, err := client.Repositories.ListReleases(context.Background(), *y.Owner.Login, *y.Name, nil)
			if err != nil {
				e.Log.Errorf("Error obtaining release metrics: %v", err)
			}

			releases = float64(len(r))
		}

		// TODO - Fix pagination
		commits := 0.0
		if e.metricEnabled("commits") {
			c, _, err := client.Repositories.ListCommits(context.Background(), *y.Owner.Login, *y.Name, nil)
			if err != nil {
				e.Log.Errorf("Error obtaining commit metrics: %v", err)
			}
			releases = float64(len(c))

		}

		e.Repositories = append(e.Repositories, RepositoryMetrics{
			Name:            e.derefString(y.Name),
			Owner:           e.derefString(y.Owner.Login),
			Archived:        strconv.FormatBool(e.derefBool(y.Archived)),
			Private:         strconv.FormatBool(e.derefBool(y.Private)),
			Fork:            strconv.FormatBool(e.derefBool(y.Fork)),
			ForksCount:      float64(e.derefInt(y.ForksCount)),
			WatchersCount:   float64(e.derefInt(y.WatchersCount)),
			StargazersCount: float64(e.derefInt(y.StargazersCount)),
			PullsCount:      pulls,
			CommitsCount:    commits,
			OpenIssuesCount: float64(e.derefInt(y.OpenIssuesCount)),
			Size:            float64(e.derefInt(y.Size)),
			Releases:        releases,
			License:         y.License.GetKey(),
			Language:        e.derefString(y.Language),
		})

	}

}
