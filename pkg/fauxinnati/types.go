package fauxinnati

import (
	"github.com/blang/semver/v4"
)

type Graph struct {
	Nodes            []Node            `json:"nodes"`
	Edges            []Edge            `json:"edges"`
	ConditionalEdges []ConditionalEdge `json:"conditionalEdges"`
}

type Node struct {
	Version  semver.Version    `json:"version"`
	Image    string            `json:"payload"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Edge [2]int

type ConditionalEdge struct {
	Edges []ConditionalUpdate     `json:"edges"`
	Risks []ConditionalUpdateRisk `json:"risks"`
}

type ConditionalUpdate struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ConditionalUpdateRisk struct {
	URL           string         `json:"url"`
	Name          string         `json:"name"`
	Message       string         `json:"message"`
	MatchingRules []MatchingRule `json:"matchingRules"`
}

type MatchingRule struct {
	Type   string       `json:"type"`
	PromQL *PromQLQuery `json:"promql,omitempty"`
}

type PromQLQuery struct {
	PromQL string `json:"promql"`
}

type ChannelInfo struct {
	Name        string
	Description string
	Example     string
	CurlCommand string
}
