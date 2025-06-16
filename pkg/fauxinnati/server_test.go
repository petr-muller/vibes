package fauxinnati

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/petr-muller/vibes/pkg/testhelper"
)

func TestServer_handleGraph(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
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
		},
		{
			name:           "GET version-not-found channel returns 200 and a three-node graph derived from version 4.17.5 but not including it",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.17.5&arch=amd64",
			expectedStatus: 200,
		},
		{
			name:           "GET version-not-found channel returns 200 and a three-node graph derived from version 4.20.0-ec.2 but not including it",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.20.0-ec.2&arch=amd64",
			expectedStatus: 200,
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
				testhelper.CompareWithFixture(t, body)
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
		expected    Graph
	}{
		{
			name:        "generates A->B->C graph from version 4.17.5",
			baseVersion: semver.MustParse("4.17.5"),
			arch:        "amd64",
			channel:     "version-not-found",
			expected:    Graph{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			result := server.generateVersionNotFoundGraph(tt.baseVersion, tt.arch, tt.channel)

			if len(result.Nodes) != 3 {
				t.Errorf("expected 3 nodes, got %d", len(result.Nodes))
			}
			if len(result.Edges) != 2 {
				t.Errorf("expected 2 edges, got %d", len(result.Edges))
			}
		})
	}
}

func TestServer_setupRoutes(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		hasRoute bool
	}{
		{
			name:     "graph endpoint exists",
			path:     "/api/upgrades_info/graph",
			method:   "GET",
			hasRoute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			server.mux.ServeHTTP(w, req)

			result := w.Result()
			if tt.hasRoute && result.StatusCode == 404 {
				t.Errorf("expected route to exist, got 404")
			}
		})
	}
}
