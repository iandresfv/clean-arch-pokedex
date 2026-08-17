// Package middleware holds cross-cutting HTTP concerns as composable wrappers
// of the form func(http.Handler) http.Handler.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// RequestIDHeader is the header used to propagate the correlation id.
const RequestIDHeader = "X-Request-ID"

// maxInboundRequestIDLen bounds an id accepted from the caller. Without a cap,
// a client could push an arbitrarily long string into every log line for the
// request, which is a cheap way to inflate log storage.
const maxInboundRequestIDLen = 128

// RequestID attaches a correlation id to every request and stores a logger
// already annotated with it.
//
// Correlation is the point: with concurrent requests interleaving in the log,
// a shared id is the only way to reconstruct what happened during one of them.
// It is also what makes the requestId in an error response actionable — a user
// can quote it and the matching lines can be found.
//
// An inbound X-Request-ID is honoured so an id survives across services, and
// echoed back on the response so the client can record it too.
func RequestID(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(RequestIDHeader)
			if id == "" || len(id) > maxInboundRequestIDLen {
				id = newRequestID()
			}

			w.Header().Set(RequestIDHeader, id)

			ctx := reqctx.WithRequestID(r.Context(), id)
			ctx = reqctx.WithLogger(ctx, logger.With("request_id", id))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// newRequestID returns 16 hex characters from crypto/rand.
//
// No UUID dependency: 8 random bytes are ample for correlating log lines, and
// crypto/rand is already in the standard library. rand.Read is documented never
// to fail as of Go 1.24, which is why its error is not checked.
func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
