package fauxinnati

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	s.mux.HandleFunc("/api/upgrades_info/graph", s.handleGraph)
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
	default:
		http.Error(w, fmt.Sprintf("Unsupported channel: %s", channel), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(graph); err != nil {
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
