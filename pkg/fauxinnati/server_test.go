package fauxinnati

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/petr-muller/vibes/pkg/testhelper"
)

func findVersion(graph Graph, version string) *Node {
	for _, node := range graph.Nodes {
		if node.Version.EQ(semver.MustParse(version)) {
			return &node
		}
	}
	return nil
}

func getVersionIndex(graph Graph, version string) int {
	sv := semver.MustParse(version)
	index := -1
	for i, node := range graph.Nodes {
		if node.Version.EQ(sv) {
			index = i
			break
		}
	}
	return index
}

func edgesTo(graph Graph, version string) []string {
	index := getVersionIndex(graph, version)

	var originSemVers []semver.Version
	for _, edge := range graph.Edges {
		if edge[1] == index {
			originSemVers = append(originSemVers, graph.Nodes[edge[0]].Version)
		}
	}

	return toStrings(originSemVers)
}

func edgesFrom(graph Graph, version string) []string {
	index := getVersionIndex(graph, version)

	var targetSemVers []semver.Version
	for _, edge := range graph.Edges {
		if edge[0] == index {
			targetSemVers = append(targetSemVers, graph.Nodes[edge[1]].Version)
		}
	}

	return toStrings(targetSemVers)
}

func toStrings(targetSemVers []semver.Version) []string {
	sort.Slice(targetSemVers, func(i, j int) bool {
		return targetSemVers[i].LT(targetSemVers[j])
	})

	edges := make([]string, len(targetSemVers))
	for i, sv := range targetSemVers {
		edges[i] = sv.String()
	}
	return edges
}

func TestServer_handleGraph(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		validateGraph  func(t *testing.T, graph Graph)
	}{
		{
			name:           "POST is disallowed",
			method:         "POST",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64",
			expectedStatus: 405,
		},
		{
			name:           "GET unknown channel returns 200 and an empty graph (like OSUS)",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=unknown&version=4.17.5&arch=amd64",
			expectedStatus: 200,
			validateGraph: func(t *testing.T, graph Graph) {
				if len(graph.Nodes) != 0 {
					t.Errorf("expected empty graph, got %d nodes", len(graph.Nodes))
				}
				if len(graph.Edges) != 0 {
					t.Errorf("expected empty graph, got %d edges", len(graph.Edges))
				}
				if len(graph.ConditionalEdges) != 0 {
					t.Errorf("expected empty graph, got %d conditional edges", len(graph.ConditionalEdges))
				}
			},
		},
		{
			name:           "GET version-not-found channel returns 200 and a three-node graph derived from version 4.17.5 but not including it",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64",
			expectedStatus: 200,
			validateGraph: func(t *testing.T, graph Graph) {
				v4175 := findVersion(graph, "4.17.5")
				if v4175 != nil {
					t.Errorf("expected version 4.17.5 not to be in the graph, but it was found")
				}
				if len(graph.Nodes) != 3 {
					t.Errorf("expected 3 nodes in the graph, got %d", len(graph.Nodes))
				}
				if len(graph.Edges) != 2 {
					t.Errorf("expected 2 edges in the graph, got %d", len(graph.Edges))
				}
			},
		},
		{
			name:           "GET version-not-found channel returns 200 and a three-node graph derived from version 4.20.0-ec.2 but not including it",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.20.0-ec.2&arch=amd64",
			expectedStatus: 200,
			validateGraph: func(t *testing.T, graph Graph) {
				v4175 := findVersion(graph, "4.17.5")
				if v4175 != nil {
					t.Errorf("expected version 4.17.5 not to be in the graph, but it was found")
				}
				if len(graph.Nodes) != 3 {
					t.Errorf("expected 3 nodes in the graph, got %d", len(graph.Nodes))
				}
				if len(graph.Edges) != 2 {
					t.Errorf("expected 2 edges in the graph, got %d", len(graph.Edges))
				}
			},
		},
		{
			name:           "GET channel-head channel returns 200 and a three-node graph with version 4.20.0-ec.2 as a channel head",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=channel-head&version=4.20.0-ec.2&arch=amd64",
			expectedStatus: 200,
			validateGraph: func(t *testing.T, graph Graph) {
				v4200ec2 := findVersion(graph, "4.20.0-ec.2")
				if v4200ec2 == nil {
					t.Errorf("expected version 4.20.0-ec.2 to be in the graph, but it was not found")
				}
				if len(graph.Nodes) != 3 {
					t.Errorf("expected 3 nodes in the graph, got %d", len(graph.Nodes))
				}
				if len(graph.Edges) != 2 {
					t.Errorf("expected 2 edges in the graph, got %d", len(graph.Edges))
				}
				if diff := cmp.Diff([]string{}, edgesFrom(graph, "4.20.0-ec.2")); diff != "" {
					t.Errorf("edges from 4.20.0-ec.2 mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff([]string{"4.19.1"}, edgesTo(graph, "4.20.0-ec.2")); diff != "" {
					t.Errorf("edges to 4.20.0-ec.2 mismatch (-want +got):\n%s", diff)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			server.handleGraph(w, req)

			result := w.Result()
			if result.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}

			if tt.expectedStatus == 200 {
				body, err := io.ReadAll(result.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}
				var graph Graph
				if err := json.Unmarshal(body, &graph); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}
				tt.validateGraph(t, graph)
			}
		})
	}
}

func TestServer_generateVersionNotFoundGraph(t *testing.T) {
	tests := []struct {
		name        string
		baseVersion semver.Version
		arch        string
		channel     string
	}{
		{
			name:        "generates A->B->C graph derived from version 4.17.5 but not including it",
			baseVersion: semver.MustParse("4.17.5"),
			arch:        "amd64",
			channel:     "version-not-found",
		},
		{
			name:        "generates A->B->C graph derived from version 4.15.0-ec.1 but not including it",
			baseVersion: semver.MustParse("4.15.0-ec.1"),
			arch:        "amd64",
			channel:     "version-not-found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			result := server.generateVersionNotFoundGraph(tt.baseVersion, tt.arch, tt.channel)
			testhelper.CompareWithFixture(t, result)
		})
	}
}

func TestServer_generateEmptyGraph(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "generates empty graph",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			result := server.generateEmptyGraph()
			testhelper.CompareWithFixture(t, result)
		})
	}
}

func TestServer_setupRoutes(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		expectedStatus int
	}{
		{
			name:           "graph endpoint exists",
			path:           "/api/upgrades_info/graph",
			method:         "GET",
			expectedStatus: 400,
		},
		{
			name:           "healthz endpoint exists",
			path:           "/healthz",
			method:         "GET",
			expectedStatus: 200,
		},
		{
			name:           "readyz endpoint exists",
			path:           "/readyz",
			method:         "GET",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			result := w.Result()
			if result.StatusCode != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d", tt.expectedStatus, result.StatusCode)
			}
		})
	}
}
