package transport

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Robotech-Org/driver-management/internal/wiring"
)

func TestDocsAndSwaggerRoutes(t *testing.T) {
	// Ensure we can locate docs from test directory
	_ = os.Chdir("../..")
	router := NewRoutes(&wiring.DependencyContainer{})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedLoc    string
		checkBodySub   string
	}{
		{
			name:           "Root redirects to /swagger/",
			path:           "/",
			expectedStatus: http.StatusFound,
			expectedLoc:    "/swagger/",
		},
		{
			name:           "/swagger redirects to /swagger/",
			path:           "/swagger",
			expectedStatus: http.StatusMovedPermanently,
			expectedLoc:    "/swagger/",
		},
		{
			name:           "/docs redirects to /swagger/",
			path:           "/docs",
			expectedStatus: http.StatusMovedPermanently,
			expectedLoc:    "/swagger/",
		},
		{
			name:           "/docs/ redirects to /swagger/",
			path:           "/docs/",
			expectedStatus: http.StatusMovedPermanently,
			expectedLoc:    "/swagger/",
		},
		{
			name:           "/swagger/ serves Swagger UI HTML",
			path:           "/swagger/",
			expectedStatus: http.StatusOK,
			checkBodySub:   "<title>Swagger UI</title>",
		},
		{
			name:           "/swagger.yaml serves API spec",
			path:           "/swagger.yaml",
			expectedStatus: http.StatusOK,
			checkBodySub:   "Driver Management API",
		},
		{
			name:           "/swagger/swagger.yaml serves API spec",
			path:           "/swagger/swagger.yaml",
			expectedStatus: http.StatusOK,
			checkBodySub:   "Driver Management API",
		},
		{
			name:           "/swagger/swagger-ui.css serves CSS",
			path:           "/swagger/swagger-ui.css",
			expectedStatus: http.StatusOK,
			checkBodySub:   ".swagger-ui",
		},
		{
			name:           "/swagger/swagger-ui-bundle.js serves JS bundle",
			path:           "/swagger/swagger-ui-bundle.js",
			expectedStatus: http.StatusOK,
			checkBodySub:   "SwaggerUIBundle",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("path %s: expected status %d, got %d", tc.path, tc.expectedStatus, rec.Code)
			}
			if tc.expectedLoc != "" {
				loc := rec.Header().Get("Location")
				if loc != tc.expectedLoc {
					t.Fatalf("path %s: expected Location header %s, got %s", tc.path, tc.expectedLoc, loc)
				}
			}
			if tc.checkBodySub != "" {
				body := rec.Body.String()
				if !strings.Contains(body, tc.checkBodySub) {
					t.Fatalf("path %s: expected body to contain %q", tc.path, tc.checkBodySub)
				}
			}
		})
	}
}
