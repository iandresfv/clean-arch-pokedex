package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[],"total":0}`))
	})
}

func TestSecurityHeaders(t *testing.T) {
	tests := []struct {
		name       string
		tlsEnabled bool
		wantHSTS   bool
	}{
		{name: "plain http", tlsEnabled: false, wantHSTS: false},
		{name: "tls enabled", tlsEnabled: true, wantHSTS: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			SecurityHeaders(tt.tlsEnabled)(okHandler()).
				ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil))

			for header, want := range map[string]string{
				"X-Content-Type-Options": "nosniff",
				"X-Frame-Options":        "DENY",
				"Referrer-Policy":        "strict-origin-when-cross-origin",
			} {
				if got := rec.Header().Get(header); got != want {
					t.Errorf("%s = %q, want %q", header, got, want)
				}
			}

			// Sending HSTS over plain HTTP is ignored by browsers and reads as
			// a false positive in an audit.
			hsts := rec.Header().Get("Strict-Transport-Security")
			if tt.wantHSTS && hsts == "" {
				t.Error("HSTS missing while TLS is enabled")
			}
			if !tt.wantHSTS && hsts != "" {
				t.Errorf("HSTS = %q sent over plain HTTP", hsts)
			}
		})
	}
}

func TestHTTPCacheReturnsNotModified(t *testing.T) {
	version := NewDatasetVersion(1)
	h := Chain(okHandler(), RequestID(discardLogger()), HTTPCache(5*time.Minute, version))

	// First request: full body plus a validator.
	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil))

	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag on the first response")
	}
	if cc := first.Header().Get("Cache-Control"); cc == "" {
		t.Error("no Cache-Control on the first response")
	}
	if first.Body.Len() == 0 {
		t.Error("first response has no body")
	}

	// Second request presenting the validator: the body must not be re-sent.
	second := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil)
	req.Header.Set("If-None-Match", etag)
	h.ServeHTTP(second, req)

	if second.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", second.Code)
	}
	if second.Body.Len() != 0 {
		t.Errorf("304 carried a body of %d bytes; it must be empty", second.Body.Len())
	}
}

// TestHTTPCacheInvalidatesOnDatasetVersionBump proves the invalidation
// strategy: because the version is folded into the validator, one increment
// orphans every cached entry at once, with no key scan.
func TestHTTPCacheInvalidatesOnDatasetVersionBump(t *testing.T) {
	version := NewDatasetVersion(1)
	h := Chain(okHandler(), RequestID(discardLogger()), HTTPCache(5*time.Minute, version))

	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil))
	oldETag := first.Header().Get("ETag")

	version.Store(2)

	second := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil)
	req.Header.Set("If-None-Match", oldETag)
	h.ServeHTTP(second, req)

	if second.Code == http.StatusNotModified {
		t.Error("a stale validator was accepted after the dataset version changed")
	}
	if second.Header().Get("ETag") == oldETag {
		t.Error("the ETag did not change with the dataset version")
	}
}

func TestHTTPCacheSkipsErrorsAndProbes(t *testing.T) {
	notFound := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404}`))
	})

	rec := httptest.NewRecorder()
	Chain(notFound, RequestID(discardLogger()), HTTPCache(5*time.Minute, NewDatasetVersion(1))).
		ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon/9999", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
	// Caching a 404 would outlive the seeding run that creates the row.
	if rec.Header().Get("ETag") != "" {
		t.Error("an error response was given a cache validator")
	}

	probe := httptest.NewRecorder()
	Chain(okHandler(), RequestID(discardLogger()), HTTPCache(5*time.Minute, NewDatasetVersion(1))).
		ServeHTTP(probe, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil))

	if probe.Header().Get("Cache-Control") != "" {
		t.Error("a probe response was made cacheable; an orchestrator would act on a stale answer")
	}
}

func TestRateLimitAllowsBurstThenThrottles(t *testing.T) {
	store := NewMemoryRateLimitStore(60, 5)
	cfg := config.RateLimit{Enabled: true, RequestsPerMin: 60, Burst: 5}
	h := Chain(okHandler(), RequestID(discardLogger()),
		RateLimit(store, NewClientIPResolver(nil), cfg))

	send := func() int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil)
		req.RemoteAddr = "203.0.113.10:54321"
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	// The bucket starts full, so the first burst passes.
	for i := range 5 {
		if code := send(); code != http.StatusOK {
			t.Fatalf("request %d was throttled at %d; the burst allowance is 5", i+1, code)
		}
	}

	if code := send(); code != http.StatusTooManyRequests {
		t.Errorf("request 6 returned %d, want 429", code)
	}
}

func TestRateLimitIsPerClient(t *testing.T) {
	store := NewMemoryRateLimitStore(60, 2)
	cfg := config.RateLimit{Enabled: true, RequestsPerMin: 60, Burst: 2}
	h := Chain(okHandler(), RequestID(discardLogger()),
		RateLimit(store, NewClientIPResolver(nil), cfg))

	send := func(addr string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil)
		req.RemoteAddr = addr
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	send("203.0.113.10:1")
	send("203.0.113.10:2")
	if code := send("203.0.113.10:3"); code != http.StatusTooManyRequests {
		t.Fatalf("the first client was not throttled: %d", code)
	}
	// A different address must have its own allowance, or one heavy client
	// would throttle everyone.
	if code := send("203.0.113.99:1"); code != http.StatusOK {
		t.Errorf("a second client was throttled by the first: %d", code)
	}
}

func TestRateLimitExemptsProbes(t *testing.T) {
	store := NewMemoryRateLimitStore(60, 1)
	cfg := config.RateLimit{Enabled: true, RequestsPerMin: 60, Burst: 1}
	h := Chain(okHandler(), RequestID(discardLogger()),
		RateLimit(store, NewClientIPResolver(nil), cfg))

	// Throttling probes makes an orchestrator restart a healthy instance.
	for i := range 10 {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
		req.RemoteAddr = "203.0.113.10:1"
		h.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			t.Fatalf("probe %d was throttled", i+1)
		}
	}
}

// TestRateLimitRefillsOverTime uses synctest, which runs the test in a bubble
// with a virtual clock: Sleep advances time instantly once every goroutine is
// blocked. Without it this test would have to sleep for real, and a rate
// limiter's behaviour is defined in minutes.
func TestRateLimitRefillsOverTime(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := NewMemoryRateLimitStore(60, 1) // one token per second
		cfg := config.RateLimit{Enabled: true, RequestsPerMin: 60, Burst: 1}
		h := Chain(okHandler(), RequestID(discardLogger()),
			RateLimit(store, NewClientIPResolver(nil), cfg))

		send := func() int {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon", nil)
			req.RemoteAddr = "203.0.113.10:1"
			h.ServeHTTP(rec, req)
			return rec.Code
		}

		if code := send(); code != http.StatusOK {
			t.Fatalf("first request returned %d", code)
		}
		if code := send(); code != http.StatusTooManyRequests {
			t.Fatalf("second request returned %d, want 429", code)
		}

		// Virtual time: instant in wall-clock terms.
		time.Sleep(2 * time.Second)

		if code := send(); code != http.StatusOK {
			t.Errorf("the bucket did not refill after two seconds: %d", code)
		}
	})
}

// TestRateLimitCleanupEvictsStaleBuckets guards against an unbounded memory
// leak: without eviction the map grows with every distinct address ever seen.
func TestRateLimitCleanupEvictsStaleBuckets(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		store := NewMemoryRateLimitStore(60, 5)
		store.ttl = 5 * time.Minute

		for _, addr := range []string{"203.0.113.1", "203.0.113.2", "203.0.113.3"} {
			store.Allow(t.Context(), addr)
		}
		if got := store.Size(); got != 3 {
			t.Fatalf("tracked %d clients, want 3", got)
		}

		go store.Cleanup(t.Context(), time.Minute)

		time.Sleep(7 * time.Minute)
		synctest.Wait()

		if got := store.Size(); got != 0 {
			t.Errorf("%d buckets survived past the TTL; the map would grow without bound", got)
		}
	})
}

func TestClientIPResolver(t *testing.T) {
	tests := []struct {
		name       string
		trusted    []string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{
			name: "no proxy configured, header ignored",
			// Believing the header unconditionally would let any caller bypass
			// the limiter by inventing an address.
			remoteAddr: "203.0.113.10:1234",
			forwarded:  "198.51.100.1",
			want:       "203.0.113.10",
		},
		{
			name:       "trusted proxy, header believed",
			trusted:    []string{"10.0.0.0/8"},
			remoteAddr: "10.0.0.5:1234",
			forwarded:  "198.51.100.1",
			want:       "198.51.100.1",
		},
		{
			name:       "chain of proxies, first untrusted from the right wins",
			trusted:    []string{"10.0.0.0/8"},
			remoteAddr: "10.0.0.5:1234",
			forwarded:  "198.51.100.1, 10.0.0.9, 10.0.0.8",
			want:       "198.51.100.1",
		},
		{
			name:       "trusted proxy without the header",
			trusted:    []string{"10.0.0.0/8"},
			remoteAddr: "10.0.0.5:1234",
			want:       "10.0.0.5",
		},
		{
			name:       "malformed header entries are skipped",
			trusted:    []string{"10.0.0.0/8"},
			remoteAddr: "10.0.0.5:1234",
			forwarded:  "not-an-ip, 198.51.100.1",
			want:       "198.51.100.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}

			if got := NewClientIPResolver(tt.trusted).Resolve(req); got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
