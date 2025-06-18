package fauxinnati

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/blang/semver/v4"
)

type Server struct {
	mux *http.ServeMux
}

func NewServer() *Server {
	s := &Server{
		mux: http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/", s.handleRoot)
	s.mux.HandleFunc("/api/upgrades_info/graph", s.handleGraph)
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/readyz", s.handleReadyz)
}

func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting fauxinnati server on %s\n", addr)
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	channel := query.Get("channel")
	version := query.Get("version")
	arch := query.Get("arch")

	if channel == "" || version == "" {
		http.Error(w, "Missing required parameters: channel and version", http.StatusBadRequest)
		return
	}

	parsedVersion, err := semver.Parse(version)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid version format: %v", err), http.StatusBadRequest)
		return
	}

	var graph Graph
	switch channel {
	case "version-not-found":
		graph = s.generateVersionNotFoundGraph(parsedVersion, arch, channel)
	case "channel-head":
		graph = s.generateChannelHeadGraph(parsedVersion, arch, channel)
	case "simple":
		graph = s.generateSimpleGraph(parsedVersion, arch, channel)
	case "risks-always":
		graph = s.generateRisksAlwaysGraph(parsedVersion, arch, channel)
	case "risks-matching":
		graph = s.generateRisksMatchingGraph(parsedVersion, arch, channel)
	case "risks-nonmatching":
		graph = s.generateRisksNonmatchingGraph(parsedVersion, arch, channel)
	case "smoke-test":
		graph = s.generateSmokeTestGraph(parsedVersion, arch, channel)
	default:
		graph = s.generateEmptyGraph()
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty print the JSON response
	if err := encoder.Encode(graph); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) generateVersionNotFoundGraph(baseVersion semver.Version, arch string, channel string) Graph {
	versionA := baseVersion
	versionA.Minor++
	versionA.Patch = 0

	versionB := versionA
	versionB.Patch = 1

	versionC := versionA
	versionC.Patch = 2

	nodeA := Node{
		Version: versionA,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionA.Major*1000+versionA.Minor*100+versionA.Patch),
		},
	}

	nodeB := Node{
		Version: versionB,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionB.Major*1000+versionB.Minor*100+versionB.Patch),
		},
	}

	nodeC := Node{
		Version: versionC,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionC.Major*1000+versionC.Minor*100+versionC.Patch),
		},
	}

	if arch != "" {
		nodeA.Metadata["release.openshift.io/architecture"] = arch
		nodeB.Metadata["release.openshift.io/architecture"] = arch
		nodeC.Metadata["release.openshift.io/architecture"] = arch
	}

	return Graph{
		Nodes: []Node{nodeA, nodeB, nodeC},
		Edges: []Edge{
			{0, 1}, // A -> B
			{1, 2}, // B -> C
		},
		ConditionalEdges: []ConditionalEdge{},
	}
}

func (s *Server) generateChannelHeadGraph(clientVersion semver.Version, arch string, channel string) Graph {
	// Client version is the head (node C)
	versionC := clientVersion

	// Node A: Previous minor version with patch 0
	versionA := clientVersion
	versionA.Minor--
	versionA.Patch = 0
	versionA.Pre = nil // Clear prerelease

	// Node B: Previous minor version with patch 1  
	versionB := versionA
	versionB.Patch = 1

	nodeA := Node{
		Version: versionA,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionA.Major*1000+versionA.Minor*100+versionA.Patch),
		},
	}

	nodeB := Node{
		Version: versionB,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionB.Major*1000+versionB.Minor*100+versionB.Patch),
		},
	}

	nodeC := Node{
		Version: versionC,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    s.formatChannelsForMetadata(versionC),
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionC.Major*1000+versionC.Minor*100+versionC.Patch),
		},
	}

	if arch != "" {
		nodeA.Metadata["release.openshift.io/architecture"] = arch
		nodeB.Metadata["release.openshift.io/architecture"] = arch
		nodeC.Metadata["release.openshift.io/architecture"] = arch
	}

	return Graph{
		Nodes: []Node{nodeA, nodeB, nodeC},
		Edges: []Edge{
			{0, 1}, // A -> B
			{1, 2}, // B -> C
		},
		ConditionalEdges: []ConditionalEdge{},
	}
}

func (s *Server) generateSimpleGraph(queriedVersion semver.Version, arch string, channel string) Graph {
	// A is the queried version
	versionA := queriedVersion

	// B: Same minor, patch bumped by one, drop prerelease
	versionB := queriedVersion
	versionB.Patch++
	versionB.Pre = nil

	// C: Minor bumped by one, patch set to zero, drop prerelease
	versionC := queriedVersion
	versionC.Minor++
	versionC.Patch = 0
	versionC.Pre = nil

	nodeA := Node{
		Version: versionA,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    s.formatChannelsForMetadata(versionA),
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionA.Major*1000+versionA.Minor*100+versionA.Patch),
		},
	}

	nodeB := Node{
		Version: versionB,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionB.Major*1000+versionB.Minor*100+versionB.Patch),
		},
	}

	nodeC := Node{
		Version: versionC,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionC.Major*1000+versionC.Minor*100+versionC.Patch),
		},
	}

	if arch != "" {
		nodeA.Metadata["release.openshift.io/architecture"] = arch
		nodeB.Metadata["release.openshift.io/architecture"] = arch
		nodeC.Metadata["release.openshift.io/architecture"] = arch
	}

	return Graph{
		Nodes: []Node{nodeA, nodeB, nodeC},
		Edges: []Edge{
			{0, 1}, // A -> B
			{0, 2}, // A -> C
		},
		ConditionalEdges: []ConditionalEdge{},
	}
}

func (s *Server) generateRisksAlwaysGraph(queriedVersion semver.Version, arch string, channel string) Graph {
	// A is the queried version
	versionA := queriedVersion

	// B: Same minor, patch bumped by one, drop prerelease
	versionB := queriedVersion
	versionB.Patch++
	versionB.Pre = nil

	// C: Minor bumped by one, patch set to zero, drop prerelease
	versionC := queriedVersion
	versionC.Minor++
	versionC.Patch = 0
	versionC.Pre = nil

	nodeA := Node{
		Version: versionA,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    s.formatChannelsForMetadata(versionA),
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionA.Major*1000+versionA.Minor*100+versionA.Patch),
		},
	}

	nodeB := Node{
		Version: versionB,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionB.Major*1000+versionB.Minor*100+versionB.Patch),
		},
	}

	nodeC := Node{
		Version: versionC,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionC.Major*1000+versionC.Minor*100+versionC.Patch),
		},
	}

	if arch != "" {
		nodeA.Metadata["release.openshift.io/architecture"] = arch
		nodeB.Metadata["release.openshift.io/architecture"] = arch
		nodeC.Metadata["release.openshift.io/architecture"] = arch
	}

	// Create conditional edges with SyntheticRisk that applies always
	conditionalEdges := []ConditionalEdge{
		{
			Edges: []ConditionalUpdate{
				{
					From: versionA.String(),
					To:   versionB.String(),
				},
				{
					From: versionA.String(),
					To:   versionC.String(),
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk",
					Name:    "SyntheticRisk", 
					Message: "This is a synthetic risk that always applies for testing purposes",
					MatchingRules: []MatchingRule{
						{
							Type: "Always",
						},
					},
				},
			},
		},
	}

	return Graph{
		Nodes:            []Node{nodeA, nodeB, nodeC},
		Edges:            []Edge{}, // No unconditional edges, only conditional
		ConditionalEdges: conditionalEdges,
	}
}

func (s *Server) generateRisksMatchingGraph(queriedVersion semver.Version, arch string, channel string) Graph {
	// A is the queried version
	versionA := queriedVersion

	// B: Same minor, patch bumped by one, drop prerelease
	versionB := queriedVersion
	versionB.Patch++
	versionB.Pre = nil

	// C: Minor bumped by one, patch set to zero, drop prerelease
	versionC := queriedVersion
	versionC.Minor++
	versionC.Patch = 0
	versionC.Pre = nil

	nodeA := Node{
		Version: versionA,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    s.formatChannelsForMetadata(versionA),
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionA.Major*1000+versionA.Minor*100+versionA.Patch),
		},
	}

	nodeB := Node{
		Version: versionB,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionB.Major*1000+versionB.Minor*100+versionB.Patch),
		},
	}

	nodeC := Node{
		Version: versionC,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionC.Major*1000+versionC.Minor*100+versionC.Patch),
		},
	}

	if arch != "" {
		nodeA.Metadata["release.openshift.io/architecture"] = arch
		nodeB.Metadata["release.openshift.io/architecture"] = arch
		nodeC.Metadata["release.openshift.io/architecture"] = arch
	}

	// Create conditional edges with SyntheticRisk using PromQL that always evaluates to 1
	conditionalEdges := []ConditionalEdge{
		{
			Edges: []ConditionalUpdate{
				{
					From: versionA.String(),
					To:   versionB.String(),
				},
				{
					From: versionA.String(),
					To:   versionC.String(),
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk-promql",
					Name:    "SyntheticRisk", 
					Message: "This is a synthetic risk with PromQL that always matches in OpenShift clusters",
					MatchingRules: []MatchingRule{
						{
							Type: "PromQL",
							PromQL: &PromQLQuery{
								PromQL: "vector(1)",
							},
						},
					},
				},
			},
		},
	}

	return Graph{
		Nodes:            []Node{nodeA, nodeB, nodeC},
		Edges:            []Edge{}, // No unconditional edges, only conditional
		ConditionalEdges: conditionalEdges,
	}
}

func (s *Server) generateRisksNonmatchingGraph(queriedVersion semver.Version, arch string, channel string) Graph {
	// A is the queried version
	versionA := queriedVersion

	// B: Same minor, patch bumped by one, drop prerelease
	versionB := queriedVersion
	versionB.Patch++
	versionB.Pre = nil

	// C: Minor bumped by one, patch set to zero, drop prerelease
	versionC := queriedVersion
	versionC.Minor++
	versionC.Patch = 0
	versionC.Pre = nil

	nodeA := Node{
		Version: versionA,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    s.formatChannelsForMetadata(versionA),
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionA.Major*1000000+versionA.Minor*1000+versionA.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionA.Major*1000+versionA.Minor*100+versionA.Patch),
		},
	}

	nodeB := Node{
		Version: versionB,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionB.Major*1000000+versionB.Minor*1000+versionB.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionB.Major*1000+versionB.Minor*100+versionB.Patch),
		},
	}

	nodeC := Node{
		Version: versionC,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    channel,
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionC.Major*1000000+versionC.Minor*1000+versionC.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionC.Major*1000+versionC.Minor*100+versionC.Patch),
		},
	}

	if arch != "" {
		nodeA.Metadata["release.openshift.io/architecture"] = arch
		nodeB.Metadata["release.openshift.io/architecture"] = arch
		nodeC.Metadata["release.openshift.io/architecture"] = arch
	}

	// Create conditional edges with SyntheticRisk using PromQL that never evaluates to true
	conditionalEdges := []ConditionalEdge{
		{
			Edges: []ConditionalUpdate{
				{
					From: versionA.String(),
					To:   versionB.String(),
				},
				{
					From: versionA.String(),
					To:   versionC.String(),
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk-promql-nonmatching",
					Name:    "SyntheticRisk", 
					Message: "This is a synthetic risk with PromQL that never matches in OpenShift clusters",
					MatchingRules: []MatchingRule{
						{
							Type: "PromQL",
							PromQL: &PromQLQuery{
								PromQL: "vector(0)",
							},
						},
					},
				},
			},
		},
	}

	return Graph{
		Nodes:            []Node{nodeA, nodeB, nodeC},
		Edges:            []Edge{}, // No unconditional edges, only conditional
		ConditionalEdges: conditionalEdges,
	}
}

func (s *Server) generateSmokeTestGraph(queriedVersion semver.Version, arch string, channel string) Graph {
	// E is the queried version
	versionE := queriedVersion

	nodeE := Node{
		Version: versionE,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionE.Major*1000000+versionE.Minor*1000+versionE.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    s.formatChannelsForMetadata(versionE),
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionE.Major*1000000+versionE.Minor*1000+versionE.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionE.Major*1000+versionE.Minor*100+versionE.Patch),
		},
	}

	if arch != "" {
		nodeE.Metadata["release.openshift.io/architecture"] = arch
	}

	// D is one version back (decrement minor, reset patch to 0, drop prerelease)
	versionD := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor - 1,
		Patch: 0,
	}
	
	// F is one patch ahead of D (so D=4.16.0, F=4.16.1)
	versionF := semver.Version{
		Major: versionD.Major,
		Minor: versionD.Minor,
		Patch: versionD.Patch + 1,
	}
	
	// G is one patch ahead of E (so E=4.17.5, G=4.17.6)
	versionG := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor,
		Patch: versionE.Patch + 1,
	}
	
	// H is one minor ahead of E (so E=4.17.5, H=4.18.0)
	versionH := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor + 1,
		Patch: 0,
	}
	
	// I is 4.17.7 (for conditional edge with RiskA:Always)
	versionI := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor,
		Patch: 7,
	}
	
	// J is 4.18.1 (for conditional edge with RiskA:Always)
	versionJ := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor + 1,
		Patch: 1,
	}
	
	// K is 4.17.8 (for conditional edge with RiskBMatches:PromQL)
	versionK := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor,
		Patch: 8,
	}
	
	// L is 4.18.2 (for conditional edge with RiskBMatches:PromQL)
	versionL := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor + 1,
		Patch: 2,
	}
	
	// M is 4.17.9 (for conditional edge with RiskCNoMatch:PromQL)
	versionM := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor,
		Patch: 9,
	}
	
	// N is 4.18.3 (for conditional edge with RiskCNoMatch:PromQL)
	versionN := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor + 1,
		Patch: 3,
	}
	
	// O is 4.17.10 (for conditional edge with combined risks)
	versionO := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor,
		Patch: 10,
	}
	
	// P is 4.18.4 (for conditional edge with combined risks)
	versionP := semver.Version{
		Major: versionE.Major,
		Minor: versionE.Minor + 1,
		Patch: 4,
	}

	nodeD := Node{
		Version: versionD,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionD.Major*1000000+versionD.Minor*1000+versionD.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionD.Major*1000000+versionD.Minor*1000+versionD.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionD.Major*1000+versionD.Minor*100+versionD.Patch),
		},
	}
	
	nodeF := Node{
		Version: versionF,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionF.Major*1000000+versionF.Minor*1000+versionF.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionF.Major*1000000+versionF.Minor*1000+versionF.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionF.Major*1000+versionF.Minor*100+versionF.Patch),
		},
	}
	
	nodeG := Node{
		Version: versionG,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionG.Major*1000000+versionG.Minor*1000+versionG.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionG.Major*1000000+versionG.Minor*1000+versionG.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionG.Major*1000+versionG.Minor*100+versionG.Patch),
		},
	}
	
	nodeH := Node{
		Version: versionH,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionH.Major*1000000+versionH.Minor*1000+versionH.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionH.Major*1000000+versionH.Minor*1000+versionH.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionH.Major*1000+versionH.Minor*100+versionH.Patch),
		},
	}
	
	nodeI := Node{
		Version: versionI,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionI.Major*1000000+versionI.Minor*1000+versionI.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionI.Major*1000000+versionI.Minor*1000+versionI.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionI.Major*1000+versionI.Minor*100+versionI.Patch),
		},
	}
	
	nodeJ := Node{
		Version: versionJ,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionJ.Major*1000000+versionJ.Minor*1000+versionJ.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionJ.Major*1000000+versionJ.Minor*1000+versionJ.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionJ.Major*1000+versionJ.Minor*100+versionJ.Patch),
		},
	}
	
	nodeK := Node{
		Version: versionK,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionK.Major*1000000+versionK.Minor*1000+versionK.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionK.Major*1000000+versionK.Minor*1000+versionK.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionK.Major*1000+versionK.Minor*100+versionK.Patch),
		},
	}
	
	nodeL := Node{
		Version: versionL,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionL.Major*1000000+versionL.Minor*1000+versionL.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionL.Major*1000000+versionL.Minor*1000+versionL.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionL.Major*1000+versionL.Minor*100+versionL.Patch),
		},
	}
	
	nodeM := Node{
		Version: versionM,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionM.Major*1000000+versionM.Minor*1000+versionM.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionM.Major*1000000+versionM.Minor*1000+versionM.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionM.Major*1000+versionM.Minor*100+versionM.Patch),
		},
	}
	
	nodeN := Node{
		Version: versionN,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionN.Major*1000000+versionN.Minor*1000+versionN.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionN.Major*1000000+versionN.Minor*1000+versionN.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionN.Major*1000+versionN.Minor*100+versionN.Patch),
		},
	}
	
	nodeO := Node{
		Version: versionO,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionO.Major*1000000+versionO.Minor*1000+versionO.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionO.Major*1000000+versionO.Minor*1000+versionO.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionO.Major*1000+versionO.Minor*100+versionO.Patch),
		},
	}
	
	nodeP := Node{
		Version: versionP,
		Image:   fmt.Sprintf("quay.io/openshift-release-dev/ocp-release@sha256:%064x", versionP.Major*1000000+versionP.Minor*1000+versionP.Patch),
		Metadata: map[string]string{
			"io.openshift.upgrades.graph.release.channels":    "smoke-test",
			"io.openshift.upgrades.graph.release.manifestref": fmt.Sprintf("sha256:%064x", versionP.Major*1000000+versionP.Minor*1000+versionP.Patch),
			"url": fmt.Sprintf("https://access.redhat.com/errata/RHSA-2024:%05d", versionP.Major*1000+versionP.Minor*100+versionP.Patch),
		},
	}

	if arch != "" {
		nodeD.Metadata["release.openshift.io/architecture"] = arch
		nodeF.Metadata["release.openshift.io/architecture"] = arch
		nodeG.Metadata["release.openshift.io/architecture"] = arch
		nodeH.Metadata["release.openshift.io/architecture"] = arch
		nodeI.Metadata["release.openshift.io/architecture"] = arch
		nodeJ.Metadata["release.openshift.io/architecture"] = arch
		nodeK.Metadata["release.openshift.io/architecture"] = arch
		nodeL.Metadata["release.openshift.io/architecture"] = arch
		nodeM.Metadata["release.openshift.io/architecture"] = arch
		nodeN.Metadata["release.openshift.io/architecture"] = arch
		nodeO.Metadata["release.openshift.io/architecture"] = arch
		nodeP.Metadata["release.openshift.io/architecture"] = arch
	}

	// Create conditional edges with mixed risk types
	conditionalEdges := []ConditionalEdge{
		{
			Edges: []ConditionalUpdate{
				{
					From: versionE.String(),
					To:   versionI.String(), // E -> I (4.17.5 -> 4.17.7)
				},
				{
					From: versionE.String(),
					To:   versionJ.String(), // E -> J (4.17.5 -> 4.18.1)
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk-smoke",
					Name:    "RiskA",
					Message: "This is a synthetic risk with Always type for smoke testing",
					MatchingRules: []MatchingRule{
						{
							Type: "Always",
						},
					},
				},
			},
		},
		{
			Edges: []ConditionalUpdate{
				{
					From: versionE.String(),
					To:   versionK.String(), // E -> K (4.17.5 -> 4.17.8)
				},
				{
					From: versionE.String(),
					To:   versionL.String(), // E -> L (4.17.5 -> 4.18.2)
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk-smoke-promql",
					Name:    "RiskBMatches",
					Message: "This is a synthetic risk with PromQL that matches for smoke testing",
					MatchingRules: []MatchingRule{
						{
							Type: "PromQL",
							PromQL: &PromQLQuery{
								PromQL: "vector(1)",
							},
						},
					},
				},
			},
		},
		{
			Edges: []ConditionalUpdate{
				{
					From: versionE.String(),
					To:   versionM.String(), // E -> M (4.17.5 -> 4.17.9)
				},
				{
					From: versionE.String(),
					To:   versionN.String(), // E -> N (4.17.5 -> 4.18.3)
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk-smoke-promql-nomatch",
					Name:    "RiskCNoMatch",
					Message: "This is a synthetic risk with PromQL that never matches for smoke testing",
					MatchingRules: []MatchingRule{
						{
							Type: "PromQL",
							PromQL: &PromQLQuery{
								PromQL: "vector(0)",
							},
						},
					},
				},
			},
		},
		{
			Edges: []ConditionalUpdate{
				{
					From: versionE.String(),
					To:   versionO.String(), // E -> O (4.17.5 -> 4.17.10)
				},
				{
					From: versionE.String(),
					To:   versionP.String(), // E -> P (4.17.5 -> 4.18.4)
				},
			},
			Risks: []ConditionalUpdateRisk{
				{
					URL:     "https://docs.openshift.com/synthetic-risk-smoke-combined-a",
					Name:    "RiskA",
					Message: "This is RiskA part of combined risks for smoke testing",
					MatchingRules: []MatchingRule{
						{
							Type: "Always",
						},
					},
				},
				{
					URL:     "https://docs.openshift.com/synthetic-risk-smoke-combined-b",
					Name:    "RiskBMatches",
					Message: "This is RiskBMatches part of combined risks for smoke testing",
					MatchingRules: []MatchingRule{
						{
							Type: "PromQL",
							PromQL: &PromQLQuery{
								PromQL: "vector(1)",
							},
						},
					},
				},
				{
					URL:     "https://docs.openshift.com/synthetic-risk-smoke-combined-c",
					Name:    "RiskCNoMatch",
					Message: "This is RiskCNoMatch part of combined risks for smoke testing",
					MatchingRules: []MatchingRule{
						{
							Type: "PromQL",
							PromQL: &PromQLQuery{
								PromQL: "vector(0)",
							},
						},
					},
				},
			},
		},
	}

	return Graph{
		Nodes:            []Node{nodeD, nodeE, nodeF, nodeG, nodeH, nodeI, nodeJ, nodeK, nodeL, nodeM, nodeN, nodeO, nodeP},
		Edges:            []Edge{{0, 1}, {0, 2}, {1, 3}, {1, 4}}, // D -> E, D -> F, E -> G, E -> H
		ConditionalEdges: conditionalEdges,
	}
}

func (s *Server) generateEmptyGraph() Graph {
	return Graph{
		Nodes:            []Node{},
		Edges:            []Edge{},
		ConditionalEdges: []ConditionalEdge{},
	}
}

// AIDEV-NOTE: Helper to determine which channels contain the queried version
// Currently channel-head, simple, risks-always, risks-matching, risks-nonmatching, and smoke-test contain the queried version, but this will expand
// as more channels are added that include the queried version in their graphs
func (s *Server) getChannelsContainingVersion(version semver.Version) []string {
	var channels []string
	
	// Channels that contain the queried version
	channels = append(channels, "channel-head")
	channels = append(channels, "risks-always")
	channels = append(channels, "risks-matching")
	channels = append(channels, "risks-nonmatching")
	channels = append(channels, "simple")
	channels = append(channels, "smoke-test")
	
	// Future channels that contain the queried version will be added here
	
	return channels
}

// AIDEV-NOTE: Format channel list for metadata field
// Returns comma-separated sorted list of channels containing the version
func (s *Server) formatChannelsForMetadata(version semver.Version) string {
	channels := s.getChannelsContainingVersion(version)
	sort.Strings(channels) // Ensure consistent ordering
	return strings.Join(channels, ",")
}

func (s *Server) generateRootHTML(host string) string {
	baseURL := fmt.Sprintf("https://%s", host)
	if host == "" {
		baseURL = "https://localhost:8080"
	}

	apiURL := fmt.Sprintf("%s/api/upgrades_info/graph", baseURL)
	exampleVersion := "4.18.42"

	// Generate live examples for each channel
	channelNames := []string{"version-not-found", "channel-head", "simple", "risks-always", "risks-matching", "risks-nonmatching", "smoke-test"}
	channelDescriptions := []string{
		"Three-node graph excluding the requested version. Creates a forward progression path.",
		"Three-node graph where the client's version is the head. Shows upgrade history.",
		"Three-node linear progression from the client's version. Basic upgrade path.",
		"Three-node graph with conditional edges that always block updates (Always matching rule).",
		"Three-node graph with PromQL conditional edges that match (PromQL: vector(1)).",
		"Three-node graph with PromQL conditional edges that don't match (PromQL: vector(0)).",
		"Comprehensive 13-node graph with mixed conditional edges for testing all Cincinnati features.",
	}
	
	var channels []ChannelInfo
	for i, name := range channelNames {
		curlCmd := fmt.Sprintf(`curl "%s?channel=%s&version=%s&arch=amd64"`, apiURL, name, exampleVersion)
		channels = append(channels, ChannelInfo{
			Name:        name,
			Description: channelDescriptions[i],
			Example:     s.generateChannelExample(name, exampleVersion),
			CurlCommand: curlCmd,
		})
	}

	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>fauxinnati - Mock Cincinnati Update Graph Server</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 2rem; line-height: 1.6; }
        .header { border-bottom: 2px solid #007acc; padding-bottom: 1rem; margin-bottom: 2rem; }
        .api-url { background: #f5f5f5; padding: 1rem; border-radius: 5px; font-family: monospace; margin: 1rem 0; }
        .channel { margin: 1.5rem 0; padding: 1rem; border: 1px solid #ddd; border-radius: 5px; }
        .channel h3 { margin-top: 0; color: #007acc; }
        .example { background: #f8f9fa; padding: 1rem; border-radius: 3px; font-family: monospace; font-size: 0.9em; margin-top: 0.5rem; white-space: pre-wrap; }
        .copy-button { background: #007acc; color: white; border: none; padding: 0.3rem 0.6rem; border-radius: 3px; cursor: pointer; font-size: 0.8em; }
        .copy-button:hover { background: #005a9f; }
        code { background: #f1f1f1; padding: 0.2rem 0.4rem; border-radius: 3px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔄 fauxinnati</h1>
        <p>Mock Cincinnati Update Graph Server for OpenShift</p>
    </div>

    <h2>📡 API Endpoint</h2>
    <div class="api-url">
        <strong>Base URL:</strong> {{.APIUrl}}
        <button class="copy-button" onclick="copyToClipboard('{{.APIUrl}}')">Copy</button>
    </div>

    <p><strong>Required Parameters:</strong></p>
    <ul>
        <li><code>channel</code> - Update channel name</li>
        <li><code>version</code> - Base version in semver format (e.g., <code>4.17.5</code>)</li>
    </ul>
    <p><strong>Optional Parameters:</strong></p>
    <ul>
        <li><code>arch</code> - Architecture (e.g., <code>amd64</code>)</li>
    </ul>

    <h2>📋 Available Channels</h2>
    <p>All examples below use version <strong>{{.ExampleVersion}}</strong> to show live graph structures:</p>

    {{range .Channels}}
    <div class="channel">
        <h3>{{.Name}}</h3>
        <p>{{.Description}}</p>
        <div class="example">{{.Example | safeHTML}}</div>
        <p><strong>Try it:</strong> <code>{{.CurlCommand}}</code> 
        <button class="copy-button" onclick="copyToClipboard('{{.CurlCommand}}')">Copy</button></p>
    </div>
    {{end}}

    <h2>ℹ️ About</h2>
    <p>fauxinnati implements the Cincinnati update graph protocol used by OpenShift clusters to discover available updates. Each channel demonstrates different graph topologies and conditional update scenarios.</p>

    <script>
        function copyToClipboard(text) {
            navigator.clipboard.writeText(text).then(function() {
                console.log('Copied to clipboard');
            });
        }
    </script>
</body>
</html>`

	data := struct {
		APIUrl         string
		ExampleVersion string
		Channels       []ChannelInfo
	}{
		APIUrl:         apiURL,
		ExampleVersion: exampleVersion,
		Channels:       channels,
	}

	t := template.Must(template.New("root").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).Parse(tmpl))
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<html><body><h1>Error generating page: %v</h1></body></html>", err)
	}

	return buf.String()
}

func (s *Server) generateChannelExample(channel, version string) string {
	parsedVersion, err := semver.Parse(version)
	if err != nil {
		return fmt.Sprintf("Error parsing version %s: %v", version, err)
	}

	var graph Graph
	switch channel {
	case "version-not-found":
		graph = s.generateVersionNotFoundGraph(parsedVersion, "amd64", channel)
	case "channel-head":
		graph = s.generateChannelHeadGraph(parsedVersion, "amd64", channel)
	case "simple":
		graph = s.generateSimpleGraph(parsedVersion, "amd64", channel)
	case "risks-always":
		graph = s.generateRisksAlwaysGraph(parsedVersion, "amd64", channel)
	case "risks-matching":
		graph = s.generateRisksMatchingGraph(parsedVersion, "amd64", channel)
	case "risks-nonmatching":
		graph = s.generateRisksNonmatchingGraph(parsedVersion, "amd64", channel)
	case "smoke-test":
		graph = s.generateSmokeTestGraph(parsedVersion, "amd64", channel)
	default:
		return "Unknown channel"
	}

	return s.graphToASCII(graph)
}

func (s *Server) graphToASCII(graph Graph) string {
	if len(graph.Nodes) == 0 {
		return "Empty graph"
	}

	var result strings.Builder
	
	// Show nodes
	result.WriteString("Nodes:\n")
	for i, node := range graph.Nodes {
		versionStr := node.Version.String()
		if versionStr == "4.18.42" {
			versionStr = "<strong>" + versionStr + "</strong>"
		}
		result.WriteString(fmt.Sprintf("  [%d] %s\n", i, versionStr))
	}
	
	// Show unconditional edges
	if len(graph.Edges) > 0 {
		result.WriteString("\nUnconditional Edges:\n")
		for _, edge := range graph.Edges {
			fromVersion := graph.Nodes[edge[0]].Version.String()
			toVersion := graph.Nodes[edge[1]].Version.String()
			if fromVersion == "4.18.42" {
				fromVersion = "<strong>" + fromVersion + "</strong>"
			}
			if toVersion == "4.18.42" {
				toVersion = "<strong>" + toVersion + "</strong>"
			}
			result.WriteString(fmt.Sprintf("  %s → %s\n", fromVersion, toVersion))
		}
	}
	
	// Show conditional edges with risks
	if len(graph.ConditionalEdges) > 0 {
		result.WriteString("\nConditional Edges:\n")
		for _, condEdge := range graph.ConditionalEdges {
			for _, edge := range condEdge.Edges {
				fromVersion := edge.From
				toVersion := edge.To
				if fromVersion == "4.18.42" {
					fromVersion = "<strong>" + fromVersion + "</strong>"
				}
				if toVersion == "4.18.42" {
					toVersion = "<strong>" + toVersion + "</strong>"
				}
				result.WriteString(fmt.Sprintf("  %s ⇢ %s", fromVersion, toVersion))
				if len(condEdge.Risks) > 0 {
					risk := condEdge.Risks[0] // Show first risk for simplicity
					if len(risk.MatchingRules) > 0 {
						rule := risk.MatchingRules[0]
						if rule.Type == "Always" {
							result.WriteString(" [Risk: Always]")
						} else if rule.Type == "PromQL" && rule.PromQL != nil {
							result.WriteString(fmt.Sprintf(" [Risk: %s]", rule.PromQL.PromQL))
						}
					}
				}
				result.WriteString("\n")
			}
		}
	}
	
	// Simple ASCII diagram for small graphs
	if len(graph.Nodes) <= 5 {
		result.WriteString("\nGraph Visualization:\n")
		result.WriteString(s.simpleGraphDiagram(graph))
	}
	
	return result.String()
}

func (s *Server) simpleGraphDiagram(graph Graph) string {
	if len(graph.Nodes) == 0 {
		return "No nodes"
	}
	
	// For simple 3-node linear graphs, show a simple diagram
	if len(graph.Nodes) == 3 {
		n0 := graph.Nodes[0].Version.String()
		n1 := graph.Nodes[1].Version.String()
		n2 := graph.Nodes[2].Version.String()
		
		// Check if it's a linear progression
		hasLinearEdges := false
		for _, edge := range graph.Edges {
			if (edge[0] == 0 && edge[1] == 1) || (edge[0] == 1 && edge[1] == 2) {
				hasLinearEdges = true
			}
		}
		
		if hasLinearEdges {
			if n0 == "4.18.42" { n0 = "<strong>" + n0 + "</strong>" }
			if n1 == "4.18.42" { n1 = "<strong>" + n1 + "</strong>" }
			if n2 == "4.18.42" { n2 = "<strong>" + n2 + "</strong>" }
			return fmt.Sprintf("  %s → %s → %s", n0, n1, n2)
		}
		
		// Check for branching pattern (0 -> 1, 0 -> 2)
		hasBranchingEdges := false
		for _, edge := range graph.Edges {
			if (edge[0] == 0 && edge[1] == 1) || (edge[0] == 0 && edge[1] == 2) {
				hasBranchingEdges = true
			}
		}
		
		if hasBranchingEdges {
			if n0 == "4.18.42" { n0 = "<strong>" + n0 + "</strong>" }
			if n1 == "4.18.42" { n1 = "<strong>" + n1 + "</strong>" }
			if n2 == "4.18.42" { n2 = "<strong>" + n2 + "</strong>" }
			return fmt.Sprintf("      %s\n     ↗\n  %s\n     ↘\n      %s", n1, n0, n2)
		}
	}
	
	// For smoke-test or complex graphs, show a summary
	if len(graph.Nodes) > 5 {
		return fmt.Sprintf("Complex graph with %d nodes, %d edges, %d conditional edge groups", 
			len(graph.Nodes), len(graph.Edges), len(graph.ConditionalEdges))
	}
	
	// Default: just list versions
	var versions []string
	for _, node := range graph.Nodes {
		versionStr := node.Version.String()
		if versionStr == "4.18.42" {
			versionStr = "<strong>" + versionStr + "</strong>"
		}
		versions = append(versions, versionStr)
	}
	return strings.Join(versions, " → ")
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	html := s.generateRootHTML(r.Host)
	w.Write([]byte(html))
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	s.healthCheck(w, r)
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	s.healthCheck(w, r)
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := r.Clone(r.Context())
	req.URL.Path = "/api/upgrades_info/graph"
	req.URL.RawQuery = "channel=stable-4.17&version=4.17.0"

	rec := &healthResponseRecorder{
		statusCode: http.StatusOK,
		headers:    make(http.Header),
	}

	s.handleGraph(rec, req)

	if rec.statusCode != http.StatusOK {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

type healthResponseRecorder struct {
	statusCode int
	headers    http.Header
}

func (h *healthResponseRecorder) Header() http.Header {
	return h.headers
}

func (h *healthResponseRecorder) Write([]byte) (int, error) {
	return 0, nil
}

func (h *healthResponseRecorder) WriteHeader(statusCode int) {
	h.statusCode = statusCode
}
