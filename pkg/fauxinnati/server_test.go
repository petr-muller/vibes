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
		{
			name:           "GET channel-head channel returns 200 and a three-node graph with version 4.20.0-ec.2 as a channel head",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=channel-head&version=4.20.0-ec.2&arch=amd64",
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
