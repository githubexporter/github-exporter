package exporter

import (
	"context"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/shurcooL/githubv4"
)

var query struct {
	Viewer struct {
		Login     githubv4.String
		CreatedAt githubv4.DateTime
	}
}

// Query GitHub API
func Query(client *githubv4.Client) {
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("    Login:", query.Viewer.Login)
	fmt.Println("CreatedAt:", query.Viewer.CreatedAt)
}
