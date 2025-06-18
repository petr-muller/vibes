package fauxinnati

import (
	"fmt"

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

// AIDEV-NOTE: Node constructor helpers to reduce code duplication across graph generation methods
// These helpers encapsulate the repetitive patterns for generating realistic OpenShift node metadata

// generateImageSHA256 creates a deterministic SHA256 hash for a version's payload image
func generateImageSHA256(version semver.Version) string {
	return fmt.Sprintf("%064x", version.Major*1000000+version.Minor*1000+version.Patch)
}

// generateManifestRef creates a deterministic manifest reference for a version
func generateManifestRef(version semver.Version) string {
	return fmt.Sprintf("sha256:%064x", version.Major*1000000+version.Minor*1000+version.Patch)
}

// generateErrataURL creates a deterministic RHSA errata URL for a version
func generateErrataURL(version semver.Version) string {
	return fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", version.Major*1000+version.Minor*100+version.Patch)
}

// NewNode creates a new Node with standard OpenShift metadata for the given version and channel
func NewNode(version semver.Version, channel string) Node {
	return Node{
		Version: version,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%s", generateImageSHA256(version)),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": generateManifestRef(version),
			"url": generateErrataURL(version),
		},
	}
}

// NewNodeWithChannelsMetadata creates a new Node with channels metadata from formatChannelsForMetadata
func NewNodeWithChannelsMetadata(version semver.Version, channelsMetadata string) Node {
	return Node{
		Version: version,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%s", generateImageSHA256(version)),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channelsMetadata,
			"io.openshift.upgrades.graph.release.manifestref": generateManifestRef(version),
			"url": generateErrataURL(version),
		},
	}
}

// SetArchitecture adds architecture metadata to a node if arch is not empty
func (n *Node) SetArchitecture(arch string) {
	if arch != "" {
		n.Metadata["release.openshift.io/architecture"] = arch
	}
}
