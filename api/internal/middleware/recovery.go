package middleware

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/httperr"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// Recovery converts a panic into a 500 response and a logged stack trace.
//
// net/http already recovers panics in handler goroutines, but it closes the
// connection without writing anything: the client sees a dropped connection
// rather than an error, and the stack trace goes to the server's default
// logger unstructured. Recovering explicitly produces a proper response and a
// trace attached to the request's correlation id.
//
// It is the outermost middleware so that a panic raised inside any other
// middleware is also caught.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// nolint:contextcheck // The recovery path deliberately takes no
			// context parameter: it runs while unwinding a panicking goroutine
			// and only reads the request's own context through reqctx.
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				// ErrAbortHandler is the standard library's signal for a
				// deliberate abort. Swallowing it would suppress intended
				// behaviour, so it is re-panicked for net/http to handle.
				if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(rec)
				}

				reqctx.Logger(r.Context()).Error("panic recovered",
					"panic", rec,
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				httperr.Write(w, r, http.StatusInternalServerError,
					"Internal Server Error", "an unexpected error occurred")
			}()

			next.ServeHTTP(w, r)
		})
	}
}
