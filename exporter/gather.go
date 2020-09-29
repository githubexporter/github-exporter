package exporter

import (
	"context"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/oauth2"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, m := range e.Metrics {
		ch <- m
	}

}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is peformed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.Log.Info("Initialising capture of metrics")

	e.newClient()

	e.gatherByOrg()
	e.gatherByUser()
	e.gatherByRepo()
	e.gatherRates()

	// Set prometheus gauge metrics using the data gathered
	e.processMetrics(ch)

	e.Log.Info("All Metrics successfully collected")

}

// newClient provides an authenticated Github
// client for use in our API interactions
func (e *Exporter) newClient() {
	ctx := context.Background()

	// Embed our authentication token in the client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: e.Config.APIToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	e.Client = github.NewClient(tc)
}

// GatherRates returns the GitHub API rate limits
// A free call that doesn't count towards your usage
func (e *Exporter) gatherRates() {

	limits, _, err := e.Client.RateLimits(context.Background())
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

// gatherByOrg specifically makes use of collection through organisation APIS
// the same function also gathers any associated metrics with the organisation
func (e *Exporter) gatherByOrg() {

	// Only execute if these have been defined
	if len(e.Config.Organisations) == 0 {
		e.Log.Info("No Organisations specified, skipping collection")
		return
	}

	// Loop through the organizations
	for _, y := range e.Config.Organisations {

		// Skip any undefined orgs
		if y == "" {
			continue
		}

		e.Log.Infof("Gathering metrics for GitHub Org %s", y)

		repos := e.fetchOrgRepos(y)
		members := e.fetchOrgMembers(y)
		collaborators := e.fetchOrgCollaborators(y)
		invites := e.fetchOrgInvites(y)

		for _, r := range repos {

			// Check this hasn't been collected prior
			if e.isDuplicateRepository(e.ProcessedRepos, r.Owner.GetLogin(), *r.Name) {
				continue
			}

			// enrich the metrics with the optional metric set
			am := e.optionalRepositoryMetrics(r)

			e.stageRepositoryMetrics(r, am)

			e.ProcessedRepos = append(e.ProcessedRepos, ProcessedRepos{
				Owner: r.Owner.GetLogin(),
				Name:  *r.Name,
			})
		}

		e.Organisations = append(e.Organisations, OrganisationMetrics{
			Name:                       y,
			MembersCount:               float64(len(members)),
			OutsideCollaboratorsCount:  float64(len(collaborators)),
			PendingOrgInvitationsCount: float64(len(invites)),
		})

	}
}

func (e *Exporter) gatherByUser() {

	// Only execute if these have been defined
	if len(e.Config.Users) == 0 {
		e.Log.Info("No Users specified, skipping collection")
		return
	}

	// Loop through the Users passed in
	for _, y := range e.Config.Users {

		// Skip any undefined users
		if y == "" {
			continue
		}

		e.Log.Info("Gathering metrics for GitHub User ", y)

		repos := e.fetchUserRepos(y)

		for _, y := range repos {

			if e.isDuplicateRepository(e.ProcessedRepos, y.Owner.GetLogin(), *y.Name) {
				continue
			}

			// enrich the metrics with the optional metric set
			am := e.optionalRepositoryMetrics(y)
			e.stageRepositoryMetrics(y, am)

			e.ProcessedRepos = append(e.ProcessedRepos, ProcessedRepos{
				Owner: y.Owner.GetLogin(),
				Name:  *y.Name,
			})
		}
	}
}

func (e *Exporter) gatherByRepo() {

	// Only execute if these have been defined
	if len(e.Config.Repositories) == 0 {
		e.Log.Info("No individual repositories specified, skipping collection")
		return
	}

	// Loop through the Users passed in
	for _, y := range e.Config.Repositories {

		// Skip any undefined users
		if y == "" {
			continue
		}

		// Prepare the arguemtns for the get
		parts := strings.Split(y, "/")
		o := parts[0]
		r := parts[1]

		e.Log.Infof("Gathering metrics for GitHub Repo %s", y)

		if e.isDuplicateRepository(e.ProcessedRepos, o, r) {
			continue
		}

		repo := e.fetchIndividualRepo(o, r)

		// enrich the metrics with the optional metric set
		am := e.optionalRepositoryMetrics(repo)
		e.stageRepositoryMetrics(repo, am)

		e.ProcessedRepos = append(e.ProcessedRepos, ProcessedRepos{
			Owner: o,
			Name:  r,
		})

	}
}

func (e *Exporter) optionalMetricEnabled(option string) bool {

	for _, v := range e.Config.OptionalMetrics {
		if v == option {
			return true
		}
	}

	return false
}

// isDuplicateRepository provides protection from collecting the same metrics twice
// Doing so causes error in the prometheus SDK
func (e *Exporter) isDuplicateRepository(repos []ProcessedRepos, o, r string) bool {

	for _, n := range repos {
		if r == n.Name && o == n.Owner {
			e.Log.Infof("Duplicate collection detected for %s/%s", o, r)
			return true
		}
	}
	return false

}

// Adds metrics not available in the standard repository response
// Also adds them to the metrics struct format for processing
func (e *Exporter) optionalRepositoryMetrics(repo *github.Repository) OptionalRepositoryMetrics {

	// TODO - Fix pagination
	pulls := 0.0
	if e.optionalMetricEnabled("pulls") {
		p, _, err := e.Client.PullRequests.List(context.Background(), *repo.Owner.Login, *repo.Name, nil)
		if err != nil {
			e.Log.Errorf("Error obtaining pull metrics: %v", err)
		}

		pulls = float64(len(p))
	}

	// TODO - Fix pagination
	releases := 0.0
	if e.optionalMetricEnabled("releases") {
		r, _, err := e.Client.Repositories.ListReleases(context.Background(), *repo.Owner.Login, *repo.Name, nil)
		if err != nil {
			e.Log.Errorf("Error obtaining release metrics: %v", err)
		}

		releases = float64(len(r))
	}

	// TODO - Fix pagination
	commits := 0.0
	if e.optionalMetricEnabled("commits") {
		c, _, err := e.Client.Repositories.ListCommits(context.Background(), *repo.Owner.Login, *repo.Name, nil)
		if err != nil {
			e.Log.Errorf("Error obtaining commit metrics: %v", err)
		}
		releases = float64(len(c))

	}

	return OptionalRepositoryMetrics{
		PullsCount:   pulls,
		CommitsCount: commits,
		Releases:     releases,
	}

}

func (e *Exporter) stageRepositoryMetrics(repo *github.Repository, am OptionalRepositoryMetrics) {

	e.Repositories = append(e.Repositories, RepositoryMetrics{
		Name:                      e.derefString(repo.Name),
		Owner:                     e.derefString(repo.Owner.Login),
		Archived:                  strconv.FormatBool(e.derefBool(repo.Archived)),
		Private:                   strconv.FormatBool(e.derefBool(repo.Private)),
		Fork:                      strconv.FormatBool(e.derefBool(repo.Fork)),
		ForksCount:                float64(e.derefInt(repo.ForksCount)),
		WatchersCount:             float64(e.derefInt(repo.WatchersCount)),
		StargazersCount:           float64(e.derefInt(repo.StargazersCount)),
		OptionalRepositoryMetrics: am,
		OpenIssuesCount:           float64(e.derefInt(repo.OpenIssuesCount)),
		Size:                      float64(e.derefInt(repo.Size)),
		License:                   repo.License.GetKey(),
		Language:                  e.derefString(repo.Language),
	})

}

func (e *Exporter) fetchOrgRepos(org string) []*github.Repository {

	// Support pagination
	var allRepos []*github.Repository
	opt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		repos, resp, err := e.Client.Repositories.ListByOrg(context.Background(), org, opt)
		if err != nil {
			e.Log.Errorf("Error listing repositories by org: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos

}

func (e *Exporter) fetchOrgMembers(org string) []*github.User {

	// Support pagination
	var allMembers []*github.User

	opt := &github.ListMembersOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		members, resp, err := e.Client.Organizations.ListMembers(context.Background(), org, opt)
		if err != nil {
			e.Log.Errorf("Error listing members by org: %v", err)
		}

		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage

	}

	return allMembers
}

func (e *Exporter) fetchOrgCollaborators(org string) []*github.User {

	// Support pagination
	var allCollabs []*github.User

	opt := &github.ListOutsideCollaboratorsOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		collabs, resp, err := e.Client.Organizations.ListOutsideCollaborators(context.Background(), org, opt)
		if err != nil {
			e.Log.Errorf("Error listing members by org: %v", err)
		}
		allCollabs = append(allCollabs, collabs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allCollabs
}

func (e *Exporter) fetchOrgInvites(org string) []*github.Invitation {

	// Support pagination
	var allInvites []*github.Invitation

	opt := &github.ListOptions{PerPage: 100}

	for {

		invites, resp, err := e.Client.Organizations.ListPendingOrgInvitations(context.Background(), org, opt)
		if err != nil {
			e.Log.Errorf("Error listing members by org: %v", err)
		}
		allInvites = append(allInvites, invites...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allInvites
}

func (e *Exporter) fetchUserRepos(user string) []*github.Repository {

	// Support pagination
	var allRepos []*github.Repository
	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		repos, resp, err := e.Client.Repositories.List(context.Background(), user, opt)
		if err != nil {
			e.Log.Errorf("Error listing repositories by user: %v", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos

}

func (e *Exporter) fetchIndividualRepo(owner, repo string) *github.Repository {

	// collect basic repository information for the repo
	r, _, err := e.Client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		e.Log.Errorf("Error collecting repository metrics: %v", err)
	}

	return r
}
