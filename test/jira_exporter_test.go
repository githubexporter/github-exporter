package test

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalIssuesResponse(t *testing.T) {
	var data = readFile("testdata/jira_response.json")
	var sr SearchResponse
	err := json.Unmarshal(data, &sr)
	if err != nil {
		panic(err)
	}
	if len(sr.Issues) != 104 {
		t.Fatal("Failed to parse the correct number of issues")
	}
}
