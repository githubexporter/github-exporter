package exporter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v76/github"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.APIMetrics {
		ch <- m
	}
}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is performed on the /metrics endpoint
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	log.Info("collecting metrics")
	var data []*Datum

	repoMetrics, err := e.getRepoMetrics(ctx)
	if err != nil {
		log.Errorf("getting repository metrics: %v", err)
		return
	}

	data = append(data, repoMetrics...)

	userMetrics, err := e.getUserMetrics(ctx)
	if err != nil {
		log.Errorf("getting user metrics: %v", err)
		return
	}

	data = append(data, userMetrics...)

	orgMetrics, err := e.getOrgMetrics(ctx)
	if err != nil {
		log.Errorf("getting organisation metrics: %v", err)
		return
	}

	data = append(data, orgMetrics...)

	r, err := e.getRateLimits(ctx)
	if err != nil {
		log.Errorf("getting rate limit metrics: %v", err)
		return
	}

	// Set prometheus gauge metrics using the data gathered
	err = e.processMetrics(data, r, ch)
	if err != nil {
		log.Errorf("processing metrics: %v", err)
	}
}

func (e *Exporter) getRateLimits(ctx context.Context) (*[]RateLimit, error) {
	if !e.GitHubRateLimitEnabled {
		return nil, nil
	}

	rates, _, err := e.Client.RateLimit.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching rate limits: %w", err)
	}

	rateLimits := map[string]*github.Rate{
		"actions_runner_registration": rates.ActionsRunnerRegistration,
		"audit_log":                   rates.AuditLog,
		"code_scanning_upload":        rates.CodeScanningUpload,
		"code_search":                 rates.CodeSearch,
		"core":                        rates.Core,
		"dependency_snapshots":        rates.DependencySnapshots,
		"graphql":                     rates.GraphQL,
		"integration_manifest":        rates.IntegrationManifest,
		"scim":                        rates.SCIM,
		"search":                      rates.Search,
		"source_import":               rates.SourceImport,
	}

	var rls []RateLimit

	for resource, rate := range rateLimits {
		if rate == nil {
			continue
		}
		r := RateLimit{
			Resource:  resource,
			Limit:     float64(rate.Limit),
			Remaining: float64(rate.Remaining),
			Reset:     float64(rate.Reset.Unix()),
		}
		rls = append(rls, r)
	}

	return &rls, nil
}

// getRepoMetrics fetches metrics for the configured repositories
func (e *Exporter) getRepoMetrics(ctx context.Context) ([]*Datum, error) {
	var data []*Datum
	for _, m := range e.Repositories {
		// Split the repository string into owner and name
		parts := strings.Split(m, "/")
		if len(parts) != 2 {
			log.Errorf("Invalid repository format: %s", m)
			continue
		}

		repo, _, err := e.Client.Repositories.Get(ctx, parts[0], parts[1])
		if err != nil {
			log.Errorf("Error fetching repository data: %v", err)
			continue
		}

		d, err := e.parseRepo(ctx, *repo)
		if err != nil {
			log.Errorf("Error parsing repository data: %v", err)
			continue
		}

		data = append(data, d)
	}

	return data, nil
}

// getUserMetrics fetches metrics for the configured users
func (e *Exporter) getUserMetrics(ctx context.Context) ([]*Datum, error) {
	var data []*Datum
	for _, m := range e.Users {
		repos, _, err := e.Client.Repositories.ListByUser(ctx, m, nil)
		if err != nil {
			log.Errorf("Error fetching user data: %v", err)
			continue
		}

		for _, repo := range repos {
			d, err := e.parseRepo(ctx, *repo)
			if err != nil {
				log.Errorf("Error parsing user repository data: %v", err)
				continue
			}

			data = append(data, d)
		}

	}
	return data, nil
}

// getOrgMetrics fetches metrics for the configured organisations
func (e *Exporter) getOrgMetrics(ctx context.Context) ([]*Datum, error) {
	var data []*Datum
	for _, m := range e.Organisations {
		repos, _, err := e.Client.Repositories.ListByOrg(ctx, m, nil)
		if err != nil {
			log.Errorf("Error fetching organisation data: %v", err)
			continue
		}

		for _, repo := range repos {
			d, err := e.parseRepo(ctx, *repo)
			if err != nil {
				log.Errorf("Error parsing organisation repository data: %v", err)
				continue
			}

			data = append(data, d)
		}
	}

	return data, nil
}

func (e *Exporter) parseRepo(ctx context.Context, repo github.Repository) (*Datum, error) {
	repoOwner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()

	rel, _, err := e.Client.Repositories.ListReleases(ctx, repoOwner, repoName, nil)
	if err != nil {
		return nil, fmt.Errorf("listing releases: %w", err)
	}

	var releases []Release
	for _, release := range rel {
		var assets []Asset
		for _, asset := range release.Assets {
			a := Asset{
				Name:      asset.GetName(),
				Size:      asset.GetSize(),
				Downloads: asset.GetDownloadCount(),
				CreatedAt: asset.GetCreatedAt().Format(time.RFC3339),
			}
			assets = append(assets, a)
		}

		r := Release{
			Name:   release.GetName(),
			Tag:    release.GetTagName(),
			Assets: assets,
		}
		releases = append(releases, r)
	}

	pullRequests, _, err := e.Client.PullRequests.List(ctx, repoOwner, repoName, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching pull requests: %w", err)
	}
	var pulls []Pull
	for _, pr := range pullRequests {
		p := Pull{
			Url: pr.GetURL(),
			User: User{
				Login: pr.GetUser().GetLogin(),
			},
		}
		pulls = append(pulls, p)
	}

	d := &Datum{
		Name: repo.GetName(),
		Owner: User{
			Login: repo.GetOwner().GetLogin(),
		},
		License: License{
			Key: repo.GetLicense().GetKey(),
		},
		Language:   repo.GetLanguage(),
		Archived:   repo.GetArchived(),
		Private:    repo.GetPrivate(),
		Fork:       repo.GetFork(),
		Forks:      repo.GetForksCount(),
		Stars:      repo.GetStargazersCount(),
		OpenIssues: repo.GetOpenIssuesCount(),
		Watchers:   repo.GetSubscribersCount(),
		Size:       repo.GetSize(),
		Releases:   releases,
		Pulls:      pulls,
	}

	return d, nil
}
