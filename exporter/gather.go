package exporter

import (
	"context"
	"fmt"

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

func (e *Exporter) gatherOrgMetrics(client *github.Client) {

	// Only execute if these have been defined
	if len(e.Organisations) == 0 {
		return
	}

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	// Loop through the organizations passed in
	for _, y := range e.Config.Organisations {

		// collect basic repository information for the org
		repos, resp, err := client.Repositories.ListByOrg(context.Background(), y, opt)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}

		em := enrichMetrics(client, repos)
		println("Remaining requests...", resp.Remaining)

		e.Repositories = append(e.Repositories, em...)

	}
}

// Adds metrics not available in the standard repository response
func enrichMetrics(client *github.Client, repos []*github.Repository) []*RepositoryMetrics {

	// First, let's create an empty struct we can return
	em := []*RepositoryMetrics{}

	// Let's then range over the repositories fed to the struct
	for _, y := range repos {

		pulls, _, err := client.PullRequests.List(context.Background(), *y.Owner.Login, *y.Name, nil)
		if err != nil {
			fmt.Print(err)
		}

		em = append(em, &RepositoryMetrics{
			PullsCount: float64(len(pulls)),
		})
	}

	return em
}

// gatherData - Collects the data from the API and stores into struct
// func (e *Exporter) gatherData(client *github.Client) ([]*RepositoryMetrics, *RateMetrics, error) {

// 	opt := &github.RepositoryListByOrgOptions{
// 		ListOptions: github.ListOptions{
// 			Page:    1,
// 			PerPage: 100,
// 		},
// 	}

// 	repoMetrics := []*RepositoryMetrics{}
// 	// x, y, err := client.Users.
// 	// user, _, err := client.Organizations.ListMembers(context.Background(), "infinityworks", nil)

// 	// // TODO pagination limits?
// 	// fmt.Printf("Number of members: %d", len(user))

// 	// client.Organizations.ListOrgMemberships()

// 	repos, resp, err := client.Repositories.ListByOrg(context.Background(), "infinityworks", opt)
// 	fmt.Printf("Size of repos: %d\n", len(repos))
// 	if err != nil {
// 		fmt.Printf("Error: %v", err)
// 	}

// 	rateMetrics := RateMetrics{
// 		Limit:     float64(resp.Limit),
// 		Remaining: float64(resp.Remaining),
// 		Reset:     float64(resp.Reset.Unix()),
// 	}

// 	for _, y := range repos {

// 		pulls, _, err := client.PullRequests.List(context.Background(), *y.Owner.Login, *y.Name, nil)
// 		if err != nil {
// 			fmt.Print(err)
// 		}

// 		repoMetrics = append(repoMetrics, &RepositoryMetrics{
// 			Name:            *y.Name,
// 			Owner:           *y.Owner.Login,
// 			License:         y.License.GetKey(),
// 			Language:        derefString(y.Language),
// 			Archived:        *y.Archived,
// 			Private:         *y.Private,
// 			Fork:            *y.Fork,
// 			ForksCount:      float64(*y.ForksCount),
// 			StarsCount:      float64(*y.StargazersCount),
// 			OpenIssuesCount: float64(*y.OpenIssuesCount),
// 			WatchersCount:   float64(*y.WatchersCount),
// 			Size:            float64(*y.Size),
// 			PullsCount:      float64(len(pulls)),
// 		})

// 	}

// 	// //return data, rates, err
// 	return repoMetrics, &rateMetrics, nil

// }

func derefString(s *string) string {
	if s != nil {
		return *s
	}

	return ""
}
