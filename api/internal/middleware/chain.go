package middleware

import (
	"net/http"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/httperr"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// Middleware wraps a handler with behaviour that runs before and after it.
type Middleware func(http.Handler) http.Handler

// Chain composes middlewares so that the first argument is the outermost
// wrapper and a request passes through them in the order written.
//
// The loop runs backwards because wrapping is inside-out: the last middleware
// must be applied to the handler first so that it ends up closest to it.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Timeout bounds how long a handler may run.
//
// Cancelling the request context propagates: an in-flight database query
// observes the cancellation and stops, instead of holding a pooled connection
// for a client that has already been answered.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// http.TimeoutHandler writes a plain-text body on expiry, which
			// would be the one response in the API not shaped like the others.
			// Wrapping it lets the timeout be reported in the same envelope.
			timed := http.TimeoutHandler(next, d, "")

			rec := &timeoutRecorder{ResponseWriter: w, r: r}
			timed.ServeHTTP(rec, r)
		})
	}
}

// timeoutRecorder replaces the plain-text 503 that http.TimeoutHandler emits
// with the API's problem+json envelope.
type timeoutRecorder struct {
	http.ResponseWriter
	r        *http.Request
	timedOut bool
}

func (tr *timeoutRecorder) WriteHeader(status int) {
	if status == http.StatusServiceUnavailable && tr.ResponseWriter.Header().Get("Content-Type") == "" {
		tr.timedOut = true
		reqctx.Logger(tr.r.Context()).Warn("request timed out",
			"method", tr.r.Method, "path", tr.r.URL.Path)
		httperr.Write(tr.ResponseWriter, tr.r, http.StatusServiceUnavailable,
			"Request Timeout", "the request took too long to process")
		return
	}
	tr.ResponseWriter.WriteHeader(status)
}

func (tr *timeoutRecorder) Write(b []byte) (int, error) {
	// The envelope has already been written; discard TimeoutHandler's body.
	if tr.timedOut {
		return len(b), nil
	}
	return tr.ResponseWriter.Write(b)
}

// Unwrap keeps http.ResponseController able to reach the real writer.
func (tr *timeoutRecorder) Unwrap() http.ResponseWriter { return tr.ResponseWriter }
