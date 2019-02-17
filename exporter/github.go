package exporter

import (
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/shurcooL/githubv4"
)

var query struct {
	Repository struct {
		Stargazers struct {
			TotalCount githubv4.Int
		}
		Issues struct {
			TotalCount githubv4.Int
		}
		Watchers struct {
			TotalCount githubv4.Int
		}
		ForkCount githubv4.Int
	} `graphql:"repository(owner: $owner, name: $name)"`
}

// Query GitHub API
func Query(client *githubv4.Client) {
	variables := map[string]interface{}{
		"owner": githubv4.String("infinityworks"),
		"name":  githubv4.String("github-exporter"),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Stars    - ", query.Repository.Stargazers.TotalCount)
	fmt.Println("Issues   - ", query.Repository.Issues.TotalCount)
	fmt.Println("Watchers - ", query.Repository.Watchers.TotalCount)
	fmt.Println("Forks    - ", query.Repository.ForkCount)
}
