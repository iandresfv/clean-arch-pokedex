package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// responseRecorder captures the status code and response size.
//
// http.ResponseWriter offers no way to read back what was written, so
// observing the outcome of a request requires wrapping it.
type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rr *responseRecorder) WriteHeader(status int) {
	// Only the first call counts: a handler that writes a body after
	// WriteHeader triggers an implicit second call, and recording that would
	// report the wrong status.
	if rr.status == 0 {
		rr.status = status
		rr.ResponseWriter.WriteHeader(status)
	}
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if rr.status == 0 {
		rr.status = http.StatusOK
	}
	n, err := rr.ResponseWriter.Write(b)
	rr.bytes += n
	return n, err
}

// Unwrap exposes the underlying writer to http.ResponseController, which is how
// Go 1.20+ reaches optional interfaces such as Flusher and Hijacker. Without
// it, wrapping would silently disable streaming and connection upgrades.
func (rr *responseRecorder) Unwrap() http.ResponseWriter { return rr.ResponseWriter }

// Logging records one structured line per request.
//
// The level is derived from the status so that a log query for problems is
// `level >= WARN` rather than a substring match on the message.
func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			duration := time.Since(start)
			log := reqctx.Logger(r.Context())

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", float64(duration.Microseconds()) / 1000,
				"bytes", rec.bytes,
				"remote_addr", clientIP(r),
			}
			if q := r.URL.RawQuery; q != "" {
				attrs = append(attrs, "query", q)
			}

			switch {
			// Probes run every few seconds and would otherwise dominate log
			// volume under an orchestrator.
			case isProbePath(r.URL.Path):
				log.Debug("request", attrs...)
			case rec.status >= http.StatusInternalServerError:
				log.Error("request", attrs...)
			case rec.status >= http.StatusBadRequest:
				log.Warn("request", attrs...)
			default:
				log.Info("request", attrs...)
			}
		})
	}
}

func isProbePath(path string) bool {
	return path == "/health" || path == "/ready"
}

// clientIP returns the peer address without its port.
//
// This is the raw connection address. Resolving the real client behind a proxy
// requires a trusted-proxy list and lives with the rate limiter, which is the
// component whose correctness depends on it.
func clientIP(r *http.Request) string {
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
