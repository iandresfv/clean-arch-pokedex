package middleware

import (
	"net/http"
	"strconv"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/httperr"
)

// SecurityHeaders sets the response headers that constrain how a browser may
// treat this API's responses.
//
// They matter even for a JSON API: a browser that can be persuaded to render a
// response as HTML, or to embed it in a frame, turns a data endpoint into an
// attack surface.
func SecurityHeaders(tlsEnabled bool) Middleware {
	// HSTS is only meaningful over TLS. Sending it on plain HTTP is ignored by
	// browsers and reads as a false positive in a security audit.
	const hstsValue = "max-age=31536000; includeSubDomains"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()

			// Without nosniff a browser may ignore Content-Type and guess, so a
			// crafted value inside a JSON field could be executed as HTML.
			h.Set("X-Content-Type-Options", "nosniff")

			// Nothing here is meant to be framed; DENY removes clickjacking as
			// a category rather than enumerating allowed ancestors.
			h.Set("X-Frame-Options", "DENY")

			// Keeps the full URL — which carries search terms and ids — from
			// leaking to third-party origins in the Referer header.
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// An API serves data, never markup or scripts, so the policy denies
			// everything. frame-ancestors repeats X-Frame-Options for browsers
			// that honour CSP instead.
			h.Set("Content-Security-Policy",
				"default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")

			// Browsers expose device APIs to any document by default; an API has
			// no use for them.
			h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			// X-Powered-By style disclosure: net/http does not add one, but a
			// proxy might, so it is cleared explicitly.
			h.Del("X-Powered-By")

			if tlsEnabled {
				h.Set("Strict-Transport-Security", hstsValue)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MaxBodySize rejects oversized request bodies.
//
// The API is read-only today, so any body is unexpected; the limit exists so a
// future write endpoint inherits a bound rather than having to remember one.
func MaxBodySize(limit int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				w.Header().Set("Connection", "close")
				httperr.Write(w, r, http.StatusRequestEntityTooLarge, "Payload Too Large",
					"request body exceeds "+strconv.FormatInt(limit, 10)+" bytes")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
