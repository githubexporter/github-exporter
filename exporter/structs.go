package exporter

import (
	"net/http"

	"github.com/benri-io/jira-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is used to store Metrics data and embeds the config struct.
// This is done so that the relevant functions have easy access to the
// user defined runtime configuration when the Collect method is called.
type Exporter struct {
	APIMetrics map[string]*prometheus.Desc
	config.Config
}

// Data is used to store an array of Datums.
// This is useful for the JSON array detection
type Data []Datum

// Datum is used to store data from all the relevant endpoints in the API
type Datum struct {
	Issues []IssueMetric
}

type IssueMetric struct {
	Project   string
	Epic      string
	Owner     string
	Creator   string
	IssueType string
	Assigned  string
	Status    string
	Priority  string
	Labels    []string
	Votes     int8
}

// RateLimits is used to store rate limit data into a struct
// This data is later represented as a metric, captured at the end of a scrape
type RateLimits struct {
	Limit     float64
	Remaining float64
	Reset     float64
}

// Response struct is used to store http.Response and associated data
type Response struct {
	url      string
	response *http.Response
	body     []byte
	err      error
}

type Issue struct {
	Expand string      `json:"expand"`
	Id     string      `json:id`
	Self   string      `json:"self"`
	Key    string      `json:"key"`
	Fields Field       `json:"fields"`
	Parent interface{} `json:"parent"`
}

type Priority struct {
	Self    string `json:"self"`
	Name    string `json:"name"`
	Id      string `json:"id"`
	IconURL string `json:"iconUrl"`
}

type Status struct {
	Self           string         `json:"self"`
	IconURL        string         `json:"iconUrl"`
	Description    string         `json:"description"`
	Name           string         `json:"name"`
	Id             int            `json:"id"`
	StatusCategory StatusCategory `json:"statusCategory"`
}

type StatusCategory struct {
	Self      string `json:"self"`
	Id        string `json:"id"`
	Key       string `json:"key"`
	ColorName string `json:"colorName"`
	Name      string `json:"name"`
}

type Field struct {
	Summary   string               `json:"summary"`
	Status    Status               `json:"status"`
	Priority  Priority             `json:"priority"`
	IssueType IssueTypeDescription `json:"issuetype"`
}

type SearchResponse struct {
	Expand     string  `json:"expand"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// Example
//"issuetype": {
//   "self": "https://benri.atlassian.net/rest/api/3/issuetype/10007",
//   "id": "10007",
//   "description": "Subtasks track small pieces of work that are part of a larger task.",
//   "iconUrl": "https://benri.atlassian.net/rest/api/2/universal_avatar/view/type/issuetype/avatar/10316?size=medium",
//   "name": "Subtask",
//   "subtask": true,
//   "avatarId": 10316,
//   "entityId": "2c4923b2-0754-499c-ab8e-0d1fefa20d99",
//   "hierarchyLevel": -1
// },
type IssueTypeDescription struct {
	Self           string `json:"self"`
	Id             int    `json:"id"`
	Description    string `json:"description"`
	IconURL        string `json:"iconUrl"`
	Name           string `json:"name"`
	Subtask        bool   `json:"subtask"`
	AvatarId       int    `json:"avatarId"`
	EntityId       string `json:"entityId"`
	HeirarchyLevel int    `json:"hierarchyLevel"`
}

type Project struct {
	Name string `json:"name"`
}
