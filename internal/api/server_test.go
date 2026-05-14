package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/registry/internal/api"
)

func TestNulByteValidationMiddleware(t *testing.T) {
	// Create a simple handler that returns "OK"
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap with our middleware
	middleware := api.NulByteValidationMiddleware(handler)

	t.Run("normal path should pass through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/servers", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("path with query params should pass through", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/servers?cursor=abc123", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("path with NUL byte should return 400", func(t *testing.T) {
		// Create request with NUL byte in path by manually setting URL
		req := httptest.NewRequest(http.MethodGet, "/v0/servers/test", nil)
		req.URL.Path = "/v0/servers/\x00"
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "URL path contains null bytes") {
			t.Errorf("expected body to contain error message, got %q", w.Body.String())
		}
		// Verify JSON response format
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", w.Header().Get("Content-Type"))
		}
	})

	t.Run("query with NUL byte should return 400", func(t *testing.T) {
		// Create request with NUL byte in query by manually setting RawQuery
		req := httptest.NewRequest(http.MethodGet, "/v0/servers", nil)
		req.URL.RawQuery = "cursor=\x00"
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "query parameters contain null bytes") {
			t.Errorf("expected body to contain error message, got %q", w.Body.String())
		}
	})

	t.Run("path with embedded NUL byte should return 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/servers/test", nil)
		req.URL.Path = "/v0/servers/test\x00name"
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("query with URL-encoded NUL byte (%00) should return 400", func(t *testing.T) {
		// This is the exact case from issue #862: ?cursor=%00
		req := httptest.NewRequest(http.MethodGet, "/v0/servers?cursor=%00", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "query parameters contain null bytes") {
			t.Errorf("expected body to contain error message, got %q", w.Body.String())
		}
	})

	t.Run("query with URL-encoded NUL byte followed by text should return 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/servers?cursor=%00test", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("query with embedded URL-encoded NUL byte should return 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v0/servers?cursor=abc%00def", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("query with double-encoded NUL byte (%2500) should pass through", func(t *testing.T) {
		// %2500 decodes to %00 (literal string), not a NUL byte
		// This is intentionally allowed - double-decoding is the caller's responsibility
		req := httptest.NewRequest(http.MethodGet, "/v0/servers?cursor=%2500", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		// This should pass - %2500 is not a NUL byte injection attempt
		// When decoded once: %2500 -> %00 (the string "%00", not a NUL byte)
		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d (double-encoded should pass)", http.StatusOK, w.Code)
		}
	})

	t.Run("query with valid percent-encoding should pass through", func(t *testing.T) {
		// Ensure we don't false-positive on valid encodings like %20 (space)
		req := httptest.NewRequest(http.MethodGet, "/v0/servers?search=hello%20world", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("path with URL-encoded NUL byte (%00) should return 400", func(t *testing.T) {
		// Handlers call url.PathUnescape() which would decode %00 to \x00
		req := httptest.NewRequest(http.MethodGet, "/v0/servers/%00/versions", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
		if !strings.Contains(w.Body.String(), "URL path contains null bytes") {
			t.Errorf("expected body to contain error message, got %q", w.Body.String())
		}
	})

	t.Run("path with URL-encoded NUL byte among other encodings should return 400", func(t *testing.T) {
		// %0a is newline, %00 is NUL - should still catch the NUL
		req := httptest.NewRequest(http.MethodGet, "/v0/servers/test%0a%00name/versions", nil)
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestTrailingSlashMiddleware(t *testing.T) {
	// Create a simple handler that returns "OK"
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap with our middleware
	middleware := api.TrailingSlashMiddleware(handler)

	tests := []struct {
		name             string
		path             string
		expectedStatus   int
		expectedLocation string
		expectRedirect   bool
	}{
		{
			name:           "root path should not redirect",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectRedirect: false,
		},
		{
			name:           "path without trailing slash should pass through",
			path:           "/v0/servers",
			expectedStatus: http.StatusOK,
			expectRedirect: false,
		},
		{
			name:             "path with trailing slash should redirect",
			path:             "/v0/servers/",
			expectedStatus:   http.StatusPermanentRedirect,
			expectedLocation: "/v0/servers",
			expectRedirect:   true,
		},
		{
			name:             "nested path with trailing slash should redirect",
			path:             "/v0/servers/123/",
			expectedStatus:   http.StatusPermanentRedirect,
			expectedLocation: "/v0/servers/123",
			expectRedirect:   true,
		},
		{
			name:             "deep nested path with trailing slash should redirect",
			path:             "/v0/auth/github/token/",
			expectedStatus:   http.StatusPermanentRedirect,
			expectedLocation: "/v0/auth/github/token",
			expectRedirect:   true,
		},
		{
			name:           "path with query params and no trailing slash should pass through",
			path:           "/v0/servers?limit=10",
			expectedStatus: http.StatusOK,
			expectRedirect: false,
		},
		{
			name:             "path with query params and trailing slash should redirect preserving query params",
			path:             "/v0/servers/?limit=10",
			expectedStatus:   http.StatusPermanentRedirect,
			expectedLocation: "/v0/servers?limit=10",
			expectRedirect:   true,
		},
		{
			// Regression test for GHSA-v8vw-gw5j-w7m6: a protocol-relative
			// path like "//evil.com/" must not redirect off-host.
			name:             "protocol-relative path should not redirect off-host",
			path:             "//evil.com/",
			expectedStatus:   http.StatusPermanentRedirect,
			expectedLocation: "/evil.com",
			expectRedirect:   true,
		},
		{
			name:             "path with multiple leading slashes should be collapsed",
			path:             "///evil.com/foo/",
			expectedStatus:   http.StatusPermanentRedirect,
			expectedLocation: "/evil.com/foo",
			expectRedirect:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectRedirect {
				location := w.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("expected Location header %q, got %q", tt.expectedLocation, location)
				}
			}
		})
	}
}

func TestEncodedSlashMiddleware(t *testing.T) {
	// Inner handler captures the path and rawPath it receives after middleware transforms them.
	var gotPath, gotRawPath string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawPath = r.URL.RawPath
		w.WriteHeader(http.StatusOK)
	})

	mw := api.EncodedSlashMiddleware(inner)

	tests := []struct {
		name string
		// path is set on r.URL.Path (the decoded path Go's HTTP parser produces)
		path string
		// rawPath, when non-empty, simulates a request that arrived with percent-encoding intact
		rawPath     string
		wantPath    string
		wantRawPath string
	}{
		// ── Envoy-decoded paths: RawPath is empty, literal "/" appears in serverName ──────────
		{
			name:        "v0 GET server version - Envoy decoded %2F",
			path:        "/v0/servers/com.example/my-server/versions/latest",
			wantPath:    "/v0/servers/com.example%2Fmy-server/versions/latest",
			wantRawPath: "/v0/servers/com.example%2Fmy-server/versions/latest",
		},
		{
			name:        "v0.1 GET server version - Envoy decoded %2F",
			path:        "/v0.1/servers/com.example/my-server/versions/latest",
			wantPath:    "/v0.1/servers/com.example%2Fmy-server/versions/latest",
			wantRawPath: "/v0.1/servers/com.example%2Fmy-server/versions/latest",
		},
		{
			name:        "v0.1 GET specific version - Envoy decoded %2F",
			path:        "/v0.1/servers/com.cortexapps/cortex-mcp/versions/1.2.3",
			wantPath:    "/v0.1/servers/com.cortexapps%2Fcortex-mcp/versions/1.2.3",
			wantRawPath: "/v0.1/servers/com.cortexapps%2Fcortex-mcp/versions/1.2.3",
		},
		{
			name:        "v0.1 GET all versions list - Envoy decoded %2F",
			path:        "/v0.1/servers/com.example/my-server/versions",
			wantPath:    "/v0.1/servers/com.example%2Fmy-server/versions",
			wantRawPath: "/v0.1/servers/com.example%2Fmy-server/versions",
		},
		{
			name:        "v0.1 PATCH version status - Envoy decoded %2F",
			path:        "/v0.1/servers/com.example/my-server/versions/1.0.0/status",
			wantPath:    "/v0.1/servers/com.example%2Fmy-server/versions/1.0.0/status",
			wantRawPath: "/v0.1/servers/com.example%2Fmy-server/versions/1.0.0/status",
		},
		{
			name:        "v0.1 PATCH server status - Envoy decoded %2F",
			path:        "/v0.1/servers/com.example/my-server/status",
			wantPath:    "/v0.1/servers/com.example%2Fmy-server/status",
			wantRawPath: "/v0.1/servers/com.example%2Fmy-server/status",
		},
		{
			name:        "serverName with dots and underscores - Envoy decoded %2F",
			path:        "/v0.1/servers/io.github.acme/my_tool-v2/versions/latest",
			wantPath:    "/v0.1/servers/io.github.acme%2Fmy_tool-v2/versions/latest",
			wantRawPath: "/v0.1/servers/io.github.acme%2Fmy_tool-v2/versions/latest",
		},
		// ── Properly encoded paths: RawPath set, should pass through unchanged ────────────────
		{
			name:        "already encoded %2F in RawPath - pass through unchanged",
			path:        "/v0.1/servers/com.example/my-server/versions/latest",
			rawPath:     "/v0.1/servers/com.example%2Fmy-server/versions/latest",
			wantPath:    "/v0.1/servers/com.example/my-server/versions/latest",
			wantRawPath: "/v0.1/servers/com.example%2Fmy-server/versions/latest",
		},
		// ── Paths with no serverName - pass through unchanged ─────────────────────────────────
		{
			name:        "list servers path - no serverName, pass through",
			path:        "/v0/servers",
			wantPath:    "/v0/servers",
			wantRawPath: "",
		},
		{
			name:        "list servers with query - no serverName, pass through",
			path:        "/v0/servers?cursor=abc",
			wantPath:    "/v0/servers",
			wantRawPath: "",
		},
		{
			name:        "non-server path - pass through unchanged",
			path:        "/v0/health",
			wantPath:    "/v0/health",
			wantRawPath: "",
		},
		{
			name:        "root path - pass through unchanged",
			path:        "/",
			wantPath:    "/",
			wantRawPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath = ""
			gotRawPath = ""

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.rawPath != "" {
				req.URL.RawPath = tt.rawPath
			}
			w := httptest.NewRecorder()

			mw.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
			if gotPath != tt.wantPath {
				t.Errorf("URL.Path:\n  got  %q\n  want %q", gotPath, tt.wantPath)
			}
			if gotRawPath != tt.wantRawPath {
				t.Errorf("URL.RawPath:\n  got  %q\n  want %q", gotRawPath, tt.wantRawPath)
			}
		})
	}
}
