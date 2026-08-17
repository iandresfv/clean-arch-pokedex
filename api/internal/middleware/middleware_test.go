package middleware

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestRecoveryReturnsProblemJSON(t *testing.T) {
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	h := Chain(panicking, RequestID(discardLogger()), Recovery())
	rec := httptest.NewRecorder()

	// The test itself must not die with the handler: an unrecovered panic in a
	// goroutine terminates the whole process, which is exactly what Recovery
	// exists to prevent.
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}

	var problem map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &problem); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if problem["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("body status = %v, want 500", problem["status"])
	}
	// The panic value must never reach the client.
	if strings.Contains(rec.Body.String(), "boom") {
		t.Errorf("panic value leaked to the client: %s", rec.Body.String())
	}
}

func TestRecoveryRepanicsOnAbortHandler(t *testing.T) {
	aborting := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		if rec := recover(); rec == nil {
			t.Error("ErrAbortHandler was swallowed; it must propagate to net/http")
		}
	}()

	h := Recovery()(aborting)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
}

func TestRequestIDGeneratedAndEchoed(t *testing.T) {
	var seen string
	h := RequestID(discardLogger())(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = reqctx.RequestID(r.Context())
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if seen == "" {
		t.Fatal("no request id placed in the context")
	}
	if got := rec.Header().Get(RequestIDHeader); got != seen {
		t.Errorf("response header = %q, context = %q; they must match", got, seen)
	}
	if len(seen) != 16 {
		t.Errorf("request id = %q, want 16 hex characters", seen)
	}
}

func TestRequestIDHonoursInboundHeader(t *testing.T) {
	tests := []struct {
		name     string
		inbound  string
		wantSame bool
	}{
		{name: "propagated across services", inbound: "abc123", wantSame: true},
		{name: "absent", inbound: "", wantSame: false},
		{
			name:     "oversized is replaced",
			inbound:  strings.Repeat("x", maxInboundRequestIDLen+1),
			wantSame: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen string
			h := RequestID(discardLogger())(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				seen = reqctx.RequestID(r.Context())
			}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			if tt.inbound != "" {
				req.Header.Set(RequestIDHeader, tt.inbound)
			}
			h.ServeHTTP(httptest.NewRecorder(), req)

			if got := seen == tt.inbound; got != tt.wantSame {
				t.Errorf("id reuse = %v, want %v (id was %q)", got, tt.wantSame, seen)
			}
		})
	}
}

func TestCORS(t *testing.T) {
	cfg := config.CORS{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowCredentials: true,
		MaxAge:           time.Hour,
	}
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		method     string
		origin     string
		preflight  bool
		wantStatus int
		wantOrigin string
	}{
		{
			name: "allowed origin", method: http.MethodGet,
			origin:     "http://localhost:5173",
			wantStatus: http.StatusOK, wantOrigin: "http://localhost:5173",
		},
		{
			name: "disallowed origin gets no grant", method: http.MethodGet,
			origin:     "https://evil.example",
			wantStatus: http.StatusOK, wantOrigin: "",
		},
		{
			name: "preflight short-circuits", method: http.MethodOptions,
			origin: "http://localhost:5173", preflight: true,
			wantStatus: http.StatusNoContent, wantOrigin: "http://localhost:5173",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tt.method, "/api/v1/pokemon", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.preflight {
				req.Header.Set("Access-Control-Request-Method", "GET")
			}

			rec := httptest.NewRecorder()
			CORS(cfg)(ok).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
				t.Errorf("Allow-Origin = %q, want %q", got, tt.wantOrigin)
			}
			// Vary is mandatory whenever the response depends on Origin, or a
			// shared cache may serve one origin's grant to another.
			if !strings.Contains(rec.Header().Get("Vary"), "Origin") {
				t.Error("Vary does not include Origin")
			}
		})
	}
}

// TestCORSHeadersPresentOnErrorResponses guards the chain ordering: with CORS
// applied inside Recovery, a 500 would arrive without a grant and the browser
// would report a CORS failure that hides the real error.
func TestCORSHeadersPresentOnErrorResponses(t *testing.T) {
	cfg := config.CORS{AllowedOrigins: []string{"http://localhost:5173"}, MaxAge: time.Hour}
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })

	h := Chain(panicking,
		Recovery(),
		RequestID(discardLogger()),
		CORS(cfg),
	)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Allow-Origin on a 500 = %q, want the origin; check middleware order", got)
	}
}

func TestChainAppliesOutermostFirst(t *testing.T) {
	var order []string
	mark := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		order = append(order, "handler")
	}), mark("first"), mark("second"), mark("third"))

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	want := []string{"first", "second", "third", "handler"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("execution order = %v, want %v", order, want)
	}
}

func TestLoggingRecordsStatus(t *testing.T) {
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short and stout"))
	}), RequestID(discardLogger()), Logging())

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d; the recorder must not alter it", rec.Code, http.StatusTeapot)
	}
	if body := rec.Body.String(); body != "short and stout" {
		t.Errorf("body = %q; the recorder must pass it through unchanged", body)
	}
}

func TestTimeoutReturnsProblemJSON(t *testing.T) {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)
	})

	h := Chain(slow, RequestID(discardLogger()), Timeout(20*time.Millisecond))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want the API's error envelope, not TimeoutHandler's plain text", ct)
	}
}
