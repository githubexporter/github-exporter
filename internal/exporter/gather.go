package exporter

import (
	"fmt"
	"github.com/google/go-github/v61/github"
	"strings"
)

const resultsPerPage = 100

func (e *Exporter) getRateLimits() (*github.RateLimits, error) {
	rateLimits, _, err := e.githubClient.RateLimit.Get(e.ctx)
	if err != nil {
		return nil, fmt.Errorf("getting rate limits: %w", err)
	}

	return rateLimits, nil
}

func (e *Exporter) getRepos() ([]*github.Repository, error) {
	listOptions := github.ListOptions{PerPage: resultsPerPage}

	orgOpts := &github.RepositoryListByOrgOptions{
		ListOptions: listOptions,
	}

	userOpts := &github.RepositoryListByUserOptions{
		ListOptions: listOptions,
	}

	var allRepos []*github.Repository

	for _, o := range e.Config.Organisations {
		repos, resp, err := e.githubClient.Repositories.ListByOrg(e.ctx, o, orgOpts)
		if err != nil {
			return nil, fmt.Errorf("getting repos for org (%s): %w", o, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		orgOpts.Page = resp.NextPage
	}

	for _, u := range e.Config.Users {
		repos, resp, err := e.githubClient.Repositories.ListByUser(e.ctx, u, userOpts)
		if err != nil {
			return nil, fmt.Errorf("getting repos for user (%s): %w", u, err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		userOpts.Page = resp.NextPage
	}

	for _, r := range e.Config.Repositories {
		// TODO - move this out of here
		repoArr := strings.Split(r, "/")

		user := repoArr[0]
		repoName := repoArr[1]
		repo, _, err := e.githubClient.Repositories.Get(e.ctx, user, repoName)
		if err != nil {
			return nil, fmt.Errorf("getting repo (%s): %w", r, err)
		}
		allRepos = append(allRepos, repo)
	}
	return allRepos, nil
}

func (e *Exporter) getPullRequestCount(owner string, repo string) (int, error) {
	opts := &github.PullRequestListOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: resultsPerPage},
	}
	pulls, _, err := e.githubClient.PullRequests.List(e.ctx, owner, repo, opts)
	if err != nil {
		return 0, err
	}

	return len(pulls), nil
}

func (e *Exporter) getReleases(owner string, repo string) ([]*github.RepositoryRelease, error) {
	opts := &github.ListOptions{PerPage: resultsPerPage}

	// TODO - check response codes
	releases, _, err := e.githubClient.Repositories.ListReleases(e.ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	return releases, nil
}
