package http_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSwaggerEndpoints(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantContains   string
		wantContentType string
	}{
		{
			name:           "GET /docs returns Swagger UI HTML",
			path:           "/docs",
			wantStatus:     http.StatusOK,
			wantContains:   "swagger-ui",
			wantContentType: "text/html",
		},
		{
			name:           "GET /docs/ returns Swagger UI HTML",
			path:           "/docs/",
			wantStatus:     http.StatusOK,
			wantContains:   "swagger-ui",
			wantContentType: "text/html",
		},
		{
			name:           "GET /swagger returns Swagger UI HTML",
			path:           "/swagger",
			wantStatus:     http.StatusOK,
			wantContains:   "swagger-ui",
			wantContentType: "text/html",
		},
		{
			name:           "GET /swagger/ returns Swagger UI HTML",
			path:           "/swagger/",
			wantStatus:     http.StatusOK,
			wantContains:   "swagger-ui",
			wantContentType: "text/html",
		},
		{
			name:           "GET /openapi.yaml returns OpenAPI spec",
			path:           "/openapi.yaml",
			wantStatus:     http.StatusOK,
			wantContains:   "openapi:",
			wantContentType: "application/yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatalf("GET %s: %v", tt.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("GET %s status = %d, want %d", tt.path, resp.StatusCode, tt.wantStatus)
			}

			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, tt.wantContentType) {
				t.Errorf("GET %s Content-Type = %q, want to contain %q", tt.path, contentType, tt.wantContentType)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}

			if !strings.Contains(string(body), tt.wantContains) {
				t.Errorf("GET %s body doesn't contain %q", tt.path, tt.wantContains)
			}
		})
	}
}
