package exporter

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/prometheus/client_golang/prometheus"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {

	for _, m := range e.Metrics {
		ch <- m
	}

}

// Collect - called on by Prometheus Client library
// This function is called when a scrape is performed on the /metrics page
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	e.Log.Info("Initialising capture of metrics")

	// TBC HOW THIS WORKS
	var allRepos []*RepositoryMetrics
	var allOrgs []*OrganisationMetrics

	// Only execute if these have been defined
	if len(e.Config.Organisations) > 0 {

		// Loop through the organizations
		for _, org := range e.Config.Organisations {

			// Skip any undefined orgs
			if org == "" {
				continue
			}

			e.Log.Infof("Gathering metrics for GitHub Org %s", org)

			org, repos, err := gatherByOrg(e.Client, org, e.Config.OptionalMetrics)
			if err != nil {
				e.Log.Error(err)
				return
			}

			allOrgs = append(allOrgs, org)
			allRepos = append(allRepos, repos...)
		}

		for _, org := range allOrgs {
			ch <- prometheus.MustNewConstMetric(e.Metrics["MembersCount"], prometheus.GaugeValue, org.MembersCount, org.Name)
			ch <- prometheus.MustNewConstMetric(e.Metrics["OutsideCollaboratorsCount"], prometheus.GaugeValue, org.OutsideCollaboratorsCount, org.Name)
			ch <- prometheus.MustNewConstMetric(e.Metrics["PendingOrgInvitationsCount"], prometheus.GaugeValue, org.PendingOrgInvitationsCount, org.Name)
		}

	}

	// Only execute if these have been defined
	if len(e.Config.Users) > 0 {
		// Loop through the organizations
		for _, user := range e.Config.Users {

			// Skip any undefined users
			if user == "" {
				continue
			}

			e.Log.Info("Gathering metrics for GitHub User ", user)

			repos, err := fetchUserRepos(e.Client, user, e.Config.OptionalMetrics)
			if err != nil {
				e.Log.Error(err)
				return
			}

			allRepos = append(allRepos, repos...)
		}
	}

	// Only execute if these have been defined
	if len(e.Config.Repositories) > 0 {

		for _, repo := range e.Config.Repositories {

			// Skip any undefined repos
			if repo == "" {
				continue
			}

			e.Log.Infof("Gathering metrics for GitHub Repo %s", repo)

			repos, err := gatherByRepo(e.Client, repo, e.Config.OptionalMetrics)
			if err != nil {
				e.Log.Error(err)
				return
			}

			allRepos = append(allRepos, repos)
		}
	}

	// Range through Repository metrics
	for _, x := range allRepos {
		ch <- prometheus.MustNewConstMetric(e.Metrics["Stars"], prometheus.GaugeValue, x.StargazersCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["Forks"], prometheus.GaugeValue, x.ForksCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["Watchers"], prometheus.GaugeValue, x.WatchersCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["Size"], prometheus.GaugeValue, x.Size, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		ch <- prometheus.MustNewConstMetric(e.Metrics["OpenIssues"], prometheus.GaugeValue, x.OpenIssuesCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)

		if optionalMetricEnabled(e.Config.OptionalMetrics, "pulls") {
			ch <- prometheus.MustNewConstMetric(e.Metrics["PullRequests"], prometheus.GaugeValue, x.OptionalRepositoryMetrics.PullsCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		}
		if optionalMetricEnabled(e.Config.OptionalMetrics, "releases") {
			ch <- prometheus.MustNewConstMetric(e.Metrics["Releases"], prometheus.GaugeValue, x.OptionalRepositoryMetrics.Releases, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)

			for _, y := range x.OptionalRepositoryMetrics.ReleaseDownloads {
				ch <- prometheus.MustNewConstMetric(e.Metrics["ReleaseDownloads"], prometheus.GaugeValue, y.DownloadCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language, y.ReleaseName, y.AssetName, y.CreatedAt)
			}
		}
		if optionalMetricEnabled(e.Config.OptionalMetrics, "commits") {
			ch <- prometheus.MustNewConstMetric(e.Metrics["Commits"], prometheus.GaugeValue, x.OptionalRepositoryMetrics.CommitsCount, x.Name, x.Owner, x.Private, x.Fork, x.Archived, x.License, x.Language)
		}
	}

	rates, err := gatherRates(e.Client)
	if err != nil {
		e.Log.Error(err)
		return
	}

	ch <- prometheus.MustNewConstMetric(e.Metrics["Limit"], prometheus.GaugeValue, rates.Limit)
	ch <- prometheus.MustNewConstMetric(e.Metrics["Remaining"], prometheus.GaugeValue, rates.Remaining)
	ch <- prometheus.MustNewConstMetric(e.Metrics["Reset"], prometheus.GaugeValue, rates.Reset)

	e.Log.Info("All Metrics successfully collected")

}

// newClient provides an authenticated Github
// client for use in our API interactions
func newClient(token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	ctx := context.Background()

	// Embed our authentication token in the client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

// GatherRates returns the GitHub API rate limits
// A free call that doesn't count towards your usage
func gatherRates(client *github.Client) (*RateMetrics, error) {

	limits, _, err := client.RateLimits(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error gathering API Rate limits: %v", err)
	}

	if limits.Core.Remaining == 0 {
		return nil, fmt.Errorf("Error - Github API Rate limit exceeded, review rates and usage")
	}

	return &RateMetrics{
		Limit:     float64(limits.Core.Limit),
		Remaining: float64(limits.Core.Remaining),
		Reset:     float64(limits.Core.Reset.Unix()),
	}, nil

}

// gatherByOrg specifically makes use of collection through organisation APIS
// the same function also gathers any associated metrics with the organisation
func gatherByOrg(client *github.Client, org string, opts []string) (*OrganisationMetrics, []*RepositoryMetrics, error) {

	repos, err := fetchOrgRepos(client, org, opts)
	if err != nil {
		return nil, nil, err
	}

	members, err := fetchOrgMembers(client, org)
	if err != nil {
		return nil, nil, err
	}

	collaborators, err := fetchOrgCollaborators(client, org)
	if err != nil {
		return nil, nil, err
	}

	invites, err := fetchOrgInvites(client, org)
	if err != nil {
		return nil, nil, err
	}

	return &OrganisationMetrics{
		Name:                       org,
		MembersCount:               float64(len(members)),
		OutsideCollaboratorsCount:  float64(len(collaborators)),
		PendingOrgInvitationsCount: float64(len(invites)),
	}, repos, nil

}

func gatherByRepo(client *github.Client, repo string, opts []string) (*RepositoryMetrics, error) {

	// Prepare the arguments for the get
	parts := strings.Split(repo, "/")
	o := parts[0]
	r := parts[1]

	// TODO - check for duplicates
	// if e.isDuplicateRepository(e.ProcessedRepos, o, r) {
	// 	continue
	// }

	// collect basic repository information for the repo
	repoMetrics, _, err := client.Repositories.Get(context.Background(), o, r)
	if err != nil {
		return nil, fmt.Errorf("Error collecting repository metrics: %v", err)
	}

	om, err := optionalRepositoryMetrics(client, o, r, opts)
	if err != nil {
		return nil, fmt.Errorf("Error collecting optional metrics: %v", err)
	}
	return newRepositoryMetrics(repoMetrics, om), nil

}

func optionalMetricEnabled(options []string, option string) bool {

	for _, v := range options {
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
			return true
		}
	}
	return false

}

// Adds metrics not available in the standard repository response
// Also adds them to the metrics struct format for processing
func optionalRepositoryMetrics(client *github.Client, owner, name string, opts []string) (*OptionalRepositoryMetrics, error) {

	var (
		pulls            float64
		commits          float64
		releases         float64
		releaseDownloads []RepoReleaseDownloads
		err              error
	)

	if optionalMetricEnabled(opts, "pulls") {
		pulls, err = fetchRepoPulls(client, owner, name)
		if err != nil {
			return nil, err
		}
	}

	if optionalMetricEnabled(opts, "commits") {
		commits, err = fetchRepoCommits(client, owner, name)
		if err != nil {
			return nil, err
		}
	}

	if optionalMetricEnabled(opts, "releases") {
		releases, releaseDownloads, err = fetchRepoReleases(client, owner, name)
		if err != nil {
			return nil, err
		}
	}

	return &OptionalRepositoryMetrics{
		PullsCount:       pulls,
		CommitsCount:     commits,
		Releases:         releases,
		ReleaseDownloads: releaseDownloads,
	}, nil

}

func newRepositoryMetrics(repo *github.Repository, opt *OptionalRepositoryMetrics) *RepositoryMetrics {

	return &RepositoryMetrics{
		Name:                      derefString(repo.Name),
		Owner:                     derefString(repo.Owner.Login),
		Archived:                  strconv.FormatBool(derefBool(repo.Archived)),
		Private:                   strconv.FormatBool(derefBool(repo.Private)),
		Fork:                      strconv.FormatBool(derefBool(repo.Fork)),
		ForksCount:                float64(derefInt(repo.ForksCount)),
		WatchersCount:             float64(derefInt(repo.WatchersCount)),
		StargazersCount:           float64(derefInt(repo.StargazersCount)),
		OptionalRepositoryMetrics: *opt,
		OpenIssuesCount:           float64(derefInt(repo.OpenIssuesCount)),
		Size:                      float64(derefInt(repo.Size)),
		License:                   repo.License.GetKey(),
		Language:                  derefString(repo.Language),
	}

}

func fetchOrgRepos(client *github.Client, org string, opts []string) ([]*RepositoryMetrics, error) {

	// Support pagination
	var allRepos []*RepositoryMetrics
	opt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		repos, resp, err := client.Repositories.ListByOrg(context.Background(), org, opt)
		if err != nil {
			return nil, fmt.Errorf("Error listing repositories by org: %v", err)
		}

		for _, repo := range repos {
			om, err := optionalRepositoryMetrics(client, *repo.Owner.Login, *repo.Name, opts)
			if err != nil {
				return nil, fmt.Errorf("Error collecting optional metrics: %v", err)
			}
			allRepos = append(allRepos, newRepositoryMetrics(repo, om))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil

}

func fetchOrgMembers(client *github.Client, org string) ([]*github.User, error) {

	// Support pagination
	var allMembers []*github.User

	opt := &github.ListMembersOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		members, resp, err := client.Organizations.ListMembers(context.Background(), org, opt)
		if err != nil {
			return nil, fmt.Errorf("Error listing members by org: %v", err)
		}

		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage

	}

	return allMembers, nil
}

func fetchOrgCollaborators(client *github.Client, org string) ([]*github.User, error) {

	// Support pagination
	var allCollabs []*github.User

	opt := &github.ListOutsideCollaboratorsOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		collabs, resp, err := client.Organizations.ListOutsideCollaborators(context.Background(), org, opt)
		if err != nil {
			return nil, fmt.Errorf("Error listing collaborators by org: %v", err)
		}
		allCollabs = append(allCollabs, collabs...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allCollabs, nil
}

func fetchOrgInvites(client *github.Client, org string) ([]*github.Invitation, error) {

	// Support pagination
	var allInvites []*github.Invitation

	opt := &github.ListOptions{PerPage: 100}

	for {

		invites, resp, err := client.Organizations.ListPendingOrgInvitations(context.Background(), org, opt)
		if err != nil {
			return nil, fmt.Errorf("Error listing invites by org: %v", err)
		}
		allInvites = append(allInvites, invites...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allInvites, nil
}

func fetchUserRepos(client *github.Client, user string, opts []string) ([]*RepositoryMetrics, error) {

	// Support pagination
	var allRepos []*RepositoryMetrics
	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		repos, resp, err := client.Repositories.List(context.Background(), user, opt)
		if err != nil {
			return nil, fmt.Errorf("Error listing repositories by user: %v", err)
		}

		for _, repo := range repos {
			om, err := optionalRepositoryMetrics(client, *repo.Owner.Login, *repo.Name, opts)
			if err != nil {
				return nil, fmt.Errorf("Error collecting optional metrics: %v", err)
			}
			allRepos = append(allRepos, newRepositoryMetrics(repo, om))
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil

}

func fetchRepoPulls(client *github.Client, owner, repo string) (float64, error) {

	// Support pagination
	var totalPulls []*github.PullRequest
	opt := &github.PullRequestListOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		p, resp, err := client.PullRequests.List(context.Background(), owner, repo, opt)
		if err != nil {
			return 0.0, fmt.Errorf("Error obtaining pull metrics: %v", err)
		}

		totalPulls = append(totalPulls, p...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return float64(len(totalPulls)), nil

}

func fetchRepoCommits(client *github.Client, owner, repo string) (float64, error) {

	// Support pagination
	var totalCommits []*github.RepositoryCommit
	opt := &github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: 100}}

	for {

		c, resp, err := client.Repositories.ListCommits(context.Background(), owner, repo, opt)
		if err != nil {
			return 0.0, fmt.Errorf("Error obtaining commit metrics: %v", err)
		}

		totalCommits = append(totalCommits, c...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return float64(len(totalCommits)), nil

}

func fetchRepoReleases(client *github.Client, owner, repo string) (float64, []RepoReleaseDownloads, error) {

	// Support pagination
	var releases []*github.RepositoryRelease
	var downloads []RepoReleaseDownloads

	opt := &github.ListOptions{PerPage: 100}

	for {

		r, resp, err := client.Repositories.ListReleases(context.Background(), owner, repo, opt)
		if err != nil {
			return 0.0, nil, fmt.Errorf("Error obtaining release metrics: %v", err)
		}

		releases = append(releases, r...)
		for _, y := range r {

			for _, x := range y.Assets {

				downloads = append(downloads, RepoReleaseDownloads{
					ReleaseName:   y.GetName(),
					AssetName:     x.GetName(),
					CreatedAt:     x.CreatedAt.String(),
					DownloadCount: float64(*x.DownloadCount),
				})
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return float64(len(releases)), downloads, nil

}
