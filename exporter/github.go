package exporter

import (
	"context"
	"fmt"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/viper"
	"log"
)

type starGazers struct {
	Stargazers struct {
		TotalCount githubv4.Int
	}
}

type issues struct {
	Issues struct {
		TotalCount githubv4.Int
	}
}

type watchers struct {
	Watchers struct {
		TotalCount githubv4.Int
	}
}

type forkCount struct {
	ForkCount githubv4.Int
}

var query struct {
	Repository struct {
		starGazers `graphql:"...@include(if:$stars)"`
		issues     `graphql:"...@include(if:$issues)"`
		watchers   `graphql:"...@include(if:$watchers)"`
		forkCount  `graphql:"...@include(if:$forks)"`
	} `graphql:"repository(owner: $owner, name: $name)"`

	RateLimit struct {
		Cost      githubv4.Int
		Limit     githubv4.Int
		Remaining githubv4.Int
		ResetAt   githubv4.DateTime
	}
}

// Query GitHub API
func Query(client *githubv4.Client) {
	variables := map[string]interface{}{
		"owner": githubv4.String("infinityworks"),
		"name":  githubv4.String("github-exporter"),

		"stars":    githubv4.Boolean(viper.GetBool("STARS")),
		"issues":   githubv4.Boolean(viper.GetBool("ISSUES")),
		"watchers": githubv4.Boolean(viper.GetBool("WATCHERS")),
		"forks":    githubv4.Boolean(viper.GetBool("FORKS")),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		log.Println(err)
	}

	fmt.Println()
	fmt.Println("Stars     - ", query.Repository.Stargazers.TotalCount)
	fmt.Println("Issues    - ", query.Repository.Issues.TotalCount)
	fmt.Println("Watchers  - ", query.Repository.Watchers.TotalCount)
	fmt.Println("Forks     - ", query.Repository.ForkCount)
	fmt.Println("Ratelimit - ",
		query.RateLimit.Cost,
		query.RateLimit.Limit,
		query.RateLimit.Remaining,
		query.RateLimit.ResetAt)

	fmt.Println(query)

}
