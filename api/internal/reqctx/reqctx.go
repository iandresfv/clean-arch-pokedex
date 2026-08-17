// Package reqctx carries per-request values through context.Context.
//
// It exists so that middleware (which sets these values) and handlers (which
// read them) can share the keys without either importing the other, which would
// invert the dependency between the two layers.
package reqctx

import (
	"context"
	"log/slog"
)

// ctxKey is unexported so no other package can collide with these keys. A
// string key would be reachable — and overwritable — by any package that
// happens to use the same literal.
type ctxKey int

const (
	requestIDKey ctxKey = iota
	loggerKey
)

// WithRequestID stores the correlation id for the current request.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the correlation id, or the empty string outside a request.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithLogger stores a logger already annotated with the request's fields.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// Logger returns the request-scoped logger, falling back to the default so a
// caller never has to nil-check. Every line it emits carries the request id,
// which is the only way to reconstruct one request from interleaved concurrent
// output.
func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
