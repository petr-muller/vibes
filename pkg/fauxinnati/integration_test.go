package fauxinnati

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/petr-muller/vibes/pkg/testhelper"
)

func TestServer_Integration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "successful graph request: any-unknown-channel channel",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=any-unknown-channel&version=4.17.5",
			headers:        map[string]string{"Accept": "application/json"},
			expectedStatus: 200,
		},
		{
			name:           "successful graph request: version-not-found channel",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.17.5",
			headers:        map[string]string{"Accept": "application/json"},
			expectedStatus: 200,
		},
		{
			name:           "successful graph request: channel-head channel",
			url:            "/api/upgrades_info/graph?channel=channel-head&version=4.17.5",
			headers:        map[string]string{"Accept": "application/json"},
			expectedStatus: 200,
		},
		{
			name:           "successful graph request: simple channel",
			url:            "/api/upgrades_info/graph?channel=simple&version=4.17.5",
			headers:        map[string]string{"Accept": "application/json"},
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			testServer := httptest.NewServer(server.mux)
			defer testServer.Close()

			req, err := http.NewRequest(tt.method, testServer.URL+tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedStatus == 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}
				testhelper.CompareWithFixture(t, body)
			}
		})
	}
}

func TestServer_FullWorkflow(t *testing.T) {
	tests := []struct {
		name        string
		description string
		steps       []struct {
			method         string
			url            string
			expectedStatus int
		}
	}{
		{
			name:        "multiple requests workflow",
			description: "Test multiple sequential requests",
			steps: []struct {
				method         string
				url            string
				expectedStatus int
			}{
				{"GET", "/api/upgrades_info/graph?channel=version-not-found&version=4.17.0", 200},
				{"GET", "/api/upgrades_info/graph?channel=version-not-found&version=4.18.0", 200},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			testServer := httptest.NewServer(server.mux)
			defer testServer.Close()

			client := &http.Client{}

			for i, step := range tt.steps {
				req, err := http.NewRequest(step.method, testServer.URL+step.url, nil)
				if err != nil {
					t.Fatalf("Step %d: Failed to create request: %v", i, err)
				}

				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("Step %d: Failed to make request: %v", i, err)
				}
				_ = resp.Body.Close()

				if resp.StatusCode != step.expectedStatus {
					t.Errorf("Step %d: expected status %d, got %d", i, step.expectedStatus, resp.StatusCode)
				}
			}
		})
	}
}

func TestServer_ConcurrentRequests(t *testing.T) {
	tests := []struct {
		name            string
		numGoroutines   int
		requestsPerGoro int
		url             string
	}{
		{
			name:            "concurrent graph requests",
			numGoroutines:   5,
			requestsPerGoro: 2,
			url:             "/api/upgrades_info/graph?channel=version-not-found&version=4.17.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			testServer := httptest.NewServer(server.mux)
			defer testServer.Close()

			var successCount atomic.Int64
			var wg sync.WaitGroup
			for i := 0; i < tt.numGoroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.requestsPerGoro; j++ {
						resp, err := http.Get(testServer.URL + tt.url)
						if err == nil && resp.StatusCode == 200 {
							successCount.Add(1)
						}
						if resp != nil {
							_ = resp.Body.Close()
						}
					}
				}()
			}
			wg.Wait()
			if actual := successCount.Load(); actual < int64(tt.numGoroutines*tt.requestsPerGoro) {
				t.Errorf("Expected at least %d successful requests, got %d", tt.numGoroutines*tt.requestsPerGoro, actual)
			}
		})
	}
}

func TestServer_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		method         string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing parameters",
			url:            "/api/upgrades_info/graph",
			method:         "GET",
			expectedStatus: 400,
			expectedError:  "Missing required parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			testServer := httptest.NewServer(server.mux)
			defer testServer.Close()

			req, err := http.NewRequest(tt.method, testServer.URL+tt.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
