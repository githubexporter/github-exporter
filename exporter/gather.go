package exporter

import (
	"context"
	"fmt"
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

func (e *Exporter) gatherRates(client *github.Client) {
	limits, _, err := client.RateLimits(context.Background())
	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	e.RateLimits.Limit = float64(limits.Core.Limit)
	e.RateLimits.Remaining = float64(limits.Core.Remaining)
	e.RateLimits.Reset = float64(limits.Core.Reset.Unix())

}

func (e *Exporter) gatherByOrg(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Config.Organisations) == 0 {
		println("No Organisations specified, skipping collection")
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

		println("Gathering metrics for ", y)

		// Support pagination
		var allRepos []*github.Repository

		for {
			repos, resp, err := client.Repositories.ListByOrg(context.Background(), y, opt)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		em := enrichMetrics(client, allRepos)

		e.Repositories = append(e.Repositories, em...)
	}
}

func (e *Exporter) gatherByUser(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Config.Users) == 0 {
		println("No Users specified, skipping collection")
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

		println("Gathering metrics for ", y)

		// Support pagination
		var allRepos []*github.Repository

		for {
			repos, resp, err := client.Repositories.List(context.Background(), y, opt)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			allRepos = append(allRepos, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		em := enrichMetrics(client, allRepos)

		e.Repositories = append(e.Repositories, em...)
	}
}

func (e *Exporter) gatherByRepo(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Config.Repositories) == 0 {
		println("No individual repositories specified, skipping collection")
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

		println("Gathering metrics for ", y)

		// Support pagination
		var allRepos []*github.Repository

		for {
			// collect basic repository information for the repo
			metrics, resp, err := client.Repositories.Get(context.Background(), owner, repo)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}

			allRepos = append(allRepos, metrics)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		// Enrich the metrics
		em := enrichMetrics(client, allRepos)

		e.Repositories = append(e.Repositories, em...)
	}
}

// Adds metrics not available in the standard repository response
// Also adds them to the metrics struct format for processing
func enrichMetrics(client *github.Client, repos []*github.Repository) []RepositoryMetrics {

	// First, let's create an empty struct we can return
	em := []RepositoryMetrics{}

	// Let's then range over the repositories fed to the struct
	for _, y := range repos {

		// TODO - Fix pagination
		pulls, _, err := client.PullRequests.List(context.Background(), *y.Owner.Login, *y.Name, nil)
		if err != nil {
			fmt.Print(err)
		}

		// TODO - Fix pagination
		releases, _, err := client.Repositories.ListReleases(context.Background(), *y.Owner.Login, *y.Name, nil)
		if err != nil {
			fmt.Print(err)
		}

		// TODO - Fix pagination
		commits, _, err := client.Repositories.ListCommits(context.Background(), *y.Owner.Login, *y.Name, nil)
		if err != nil {
			fmt.Print(err)
		}

		em = append(em, RepositoryMetrics{
			Name:            derefString(y.Name),
			Owner:           derefString(y.Owner.Login),
			Archived:        strconv.FormatBool(derefBool(y.Archived)),
			Private:         strconv.FormatBool(derefBool(y.Private)),
			Fork:            strconv.FormatBool(derefBool(y.Fork)),
			ForksCount:      float64(derefInt(y.ForksCount)),
			WatchersCount:   float64(derefInt(y.WatchersCount)),
			StargazersCount: float64(derefInt(y.StargazersCount)),
			PullsCount:      float64(len(pulls)),
			CommitsCount:    float64(len(commits)),
			OpenIssuesCount: float64(derefInt(y.OpenIssuesCount)),
			Size:            float64(derefInt(y.Size)),
			Releases:        float64(len(releases)),
			License:         y.License.GetKey(),
			Language:        derefString(y.Language),
		})
	}

	return em
}
