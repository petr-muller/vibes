package fauxinnati

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

func TestServer_Integration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		headers        map[string]string
		expectedStatus int
		expectedType   string
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name:           "successful graph request",
			method:         "GET",
			url:            "/api/upgrades_info/graph?channel=version-not-found&version=4.17.5",
			headers:        map[string]string{"Accept": "application/json"},
			expectedStatus: 200,
			expectedType:   "application/json",
			validateBody:   nil,
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
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
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
				resp.Body.Close()

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
							resp.Body.Close()
						}
					}
				}()
			}
			wg.Wait()
			if successCount.Load() < int64(tt.numGoroutines*tt.requestsPerGoro) {
				t.Errorf("Expected at least %d successful requests, got %d", tt.numGoroutines*tt.requestsPerGoro, successCount)
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
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}
