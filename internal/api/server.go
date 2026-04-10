package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/cors"

	v0 "github.com/modelcontextprotocol/registry/internal/api/handlers/v0"
	"github.com/modelcontextprotocol/registry/internal/api/router"
	"github.com/modelcontextprotocol/registry/internal/config"
	"github.com/modelcontextprotocol/registry/internal/service"
	"github.com/modelcontextprotocol/registry/internal/telemetry"
)

// NulByteValidationMiddleware rejects requests containing NUL bytes in URL path or query parameters.
// This prevents PostgreSQL encoding errors (SQLSTATE 22021) and returns a proper 400 Bad Request.
// Checks for both literal NUL bytes (\x00) and URL-encoded form (%00).
func NulByteValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check URL path for literal NUL bytes or URL-encoded %00
		// Path needs %00 check because handlers call url.PathUnescape() which would decode it
		if containsNulByte(r.URL.Path) {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request: URL path contains null bytes")
			return
		}

		// Check raw query string for literal NUL bytes or URL-encoded %00
		if containsNulByte(r.URL.RawQuery) {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request: query parameters contain null bytes")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse writes a JSON error response using huma's ErrorModel format
// for consistency with the rest of the API.
func writeErrorResponse(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errModel := &huma.ErrorModel{
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	}
	_ = json.NewEncoder(w).Encode(errModel)
}

// containsNulByte checks if a string contains a NUL byte, either as a literal \x00
// or URL-encoded as %00.
func containsNulByte(s string) bool {
	// Check for literal NUL byte
	if strings.ContainsRune(s, '\x00') {
		return true
	}
	// Check for URL-encoded NUL byte (%00)
	// Using Contains directly since %00 has no case variation (both hex digits are 0)
	return strings.Contains(s, "%00")
}

// serverNamePathRx matches routes where Envoy has decoded a %2F-encoded slash
// in the {serverName} path segment (e.g. "com.example/my-server" instead of
// "com.example%2Fmy-server"). The pattern captures:
//
//   - m[1]: API version prefix ("v0" or "v0.1")
//   - m[2]: namespace portion of serverName (e.g. "com.example")
//   - m[3]: name portion of serverName (e.g. "my-server")
//   - m[4]: remainder of the path after serverName (may be empty, e.g. "/versions/latest")
//
// Character classes are intentionally broader than the canonical namespacePattern /
// namePartPattern in internal/validators/validators.go. Over-matching here is safe:
// any structurally invalid server name still reaches the service layer which rejects it.
// What matters is that we only fire when there are exactly two consecutive path segments
// after /servers/ that together look like a two-part server name.
var serverNamePathRx = regexp.MustCompile(
	`^/(v0(?:\.1)?)/servers/([A-Za-z0-9._-]+)/([A-Za-z0-9._-]+)(/.*)?$`,
)

// EncodedSlashMiddleware re-encodes a decoded slash in {serverName} path segments.
//
// Azure Container Apps (Envoy proxy) decodes %2F → "/" in request paths before
// forwarding to the Go application. Server names follow the pattern "namespace/name"
// (e.g. "com.example/my-server"), which clients URL-encode as "com.example%2Fmy-server".
// After Envoy decodes this, the Go HTTP mux sees one extra path segment and cannot
// match any registered route, returning a routing-level 404.
//
// This middleware detects the decoded form (r.URL.RawPath == "", indicating no
// percent-encoding survived to the Go layer) and re-encodes the slash between
// namespace and name as %2F so that the mux can route correctly.
//
// # Intentional url.URL.Path convention
//
// We set r.URL.Path to the percent-encoded form (containing literal "%2F") even
// though url.URL.Path is conventionally the decoded path. net/http.ServeMux routes
// on r.URL.Path directly — not via url.EscapedPath() — so placing "%2F" in Path
// causes the mux to treat "namespace%2Fname" as a single path segment, matching
// the registered {serverName} wildcard. RawPath is set to the same value so that
// huma's path-param extraction, which calls url.PathUnescape(RawPath), correctly
// decodes "%2F" back to "/" and delivers "namespace/name" to the handler.
//
// Do NOT call r.URL.String() or r.URL.EscapedPath() on this request downstream;
// their output will be wrong because PathUnescape(RawPath) != Path by design.
func EncodedSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only re-encode when RawPath is empty, which means the request arrived
		// without any percent-encoding (i.e. Envoy already decoded %2F → /).
		// If RawPath is non-empty the client sent a properly-encoded URL; leave it alone.
		if r.URL.RawPath == "" {
			if m := serverNamePathRx.FindStringSubmatch(r.URL.Path); m != nil {
				// Reconstruct the path with %2F between namespace (m[2]) and name (m[3]).
				encoded := "/" + m[1] + "/servers/" + m[2] + "%2F" + m[3] + m[4]
				r = r.Clone(r.Context())
				// See doc comment above for why Path is intentionally set to the encoded form.
				r.URL.Path = encoded
				r.URL.RawPath = encoded
			}
		}
		next.ServeHTTP(w, r)
	})
}

// TrailingSlashMiddleware redirects requests with trailing slashes to their canonical form
func TrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only redirect if the path is not "/" and ends with a "/"
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			// Create a copy of the URL and remove the trailing slash
			newURL := *r.URL
			newURL.Path = strings.TrimSuffix(r.URL.Path, "/")

			// Use 308 Permanent Redirect to preserve the request method
			http.Redirect(w, r, newURL.String(), http.StatusPermanentRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Server represents the HTTP server
type Server struct {
	config   *config.Config
	registry service.RegistryService
	humaAPI  huma.API
	server   *http.Server
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, registryService service.RegistryService, metrics *telemetry.Metrics, versionInfo *v0.VersionBody) *Server {
	// Create HTTP mux and Huma API
	mux := http.NewServeMux()

	api := router.NewHumaAPI(cfg, registryService, mux, metrics, versionInfo)

	// Configure CORS with permissive settings for public API
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Type", "Content-Length"},
		AllowCredentials: false, // Must be false when AllowedOrigins is "*"
		MaxAge:           86400, // 24 hours
	})

	// Wrap the mux with middleware stack
	// Order: NulByteValidation -> EncodedSlash -> TrailingSlash -> CORS -> Mux
	handler := NulByteValidationMiddleware(EncodedSlashMiddleware(TrailingSlashMiddleware(corsHandler.Handler(mux))))

	server := &Server{
		config:   cfg,
		registry: registryService,
		humaAPI:  api,
		server: &http.Server{
			Addr:              cfg.ServerAddress,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}

	return server
}

// Start begins listening for incoming HTTP requests
func (s *Server) Start() error {
	log.Printf("HTTP server starting on %s", s.config.ServerAddress)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
