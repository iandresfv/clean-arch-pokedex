package middleware

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
)

// CORS answers cross-origin requests according to the configured allowlist.
//
// Cross-Origin Resource Sharing is enforced by the browser, not the server: a
// page served from localhost:5173 may not read a response from localhost:8080
// unless that response carries headers granting permission. curl is unaffected,
// which is why a broken CORS setup passes every command-line test and fails
// only in the application.
//
// A preflight is the OPTIONS request a browser sends first for any request that
// is not "simple", asking whether the real one is permitted.
func CORS(cfg config.CORS) func(http.Handler) http.Handler {
	maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Origin is echoed rather than answered with "*" because a wildcard
			// is rejected outright by browsers when credentials are enabled.
			// Echoing makes the response origin-specific, which is why Vary
			// below is mandatory.
			if origin != "" && isOriginAllowed(origin, cfg.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if cfg.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			// Without Vary, a shared cache could serve one origin's response —
			// including its Allow-Origin header — to a different origin.
			w.Header().Add("Vary", "Origin")

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, "+RequestIDHeader)
				w.Header().Set("Access-Control-Max-Age", maxAge)
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")

				// A preflight carries no body and must not reach the router.
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Lets the client read the correlation id, which is otherwise
			// hidden: scripts may only access a short list of response headers
			// unless the server opts others in.
			w.Header().Set("Access-Control-Expose-Headers", RequestIDHeader)

			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(origin string, allowed []string) bool {
	return slices.Contains(allowed, origin)
}
