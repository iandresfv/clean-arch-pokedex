package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/httperr"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// RateLimitStore holds per-client request counters.
//
// The interface exists because the in-memory implementation is correct only
// with a single instance: a Kubernetes Service spreads a client's connections
// across replicas, so each pod sees a fraction of the traffic and the effective
// limit becomes N times the configured one. The Redis implementation moves the
// counters somewhere every replica can see them.
type RateLimitStore interface {
	// Allow reports whether the client may proceed, and how long to wait if
	// not.
	Allow(ctx context.Context, key string) (allowed bool, retryAfter time.Duration)
}

// tokenBucket is one client's allowance.
//
// A token bucket refills at a constant rate up to a fixed capacity: short
// bursts are absorbed up to the capacity while the long-run average stays at
// the refill rate. Tokens are stored as a float so a partial refill is not lost
// to truncation between requests.
type tokenBucket struct {
	tokens   float64
	lastSeen time.Time
}

// MemoryRateLimitStore keeps buckets in process memory.
type MemoryRateLimitStore struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	capacity float64
	// refill is tokens per second.
	refill float64
	ttl    time.Duration
}

// NewMemoryRateLimitStore builds an in-process limiter.
func NewMemoryRateLimitStore(requestsPerMinute, burst int) *MemoryRateLimitStore {
	return &MemoryRateLimitStore{
		buckets:  make(map[string]*tokenBucket),
		capacity: float64(burst),
		refill:   float64(requestsPerMinute) / 60.0,
		ttl:      10 * time.Minute,
	}
}

// Allow implements RateLimitStore.
func (s *MemoryRateLimitStore) Allow(_ context.Context, key string) (bool, time.Duration) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.buckets[key]
	if !ok {
		s.buckets[key] = &tokenBucket{tokens: s.capacity - 1, lastSeen: now}
		return true, 0
	}

	b.tokens = min(s.capacity, b.tokens+now.Sub(b.lastSeen).Seconds()*s.refill)
	b.lastSeen = now

	if b.tokens < 1 {
		// Time until one whole token is available again.
		return false, time.Duration((1-b.tokens)/s.refill*float64(time.Second)) + time.Second
	}

	b.tokens--
	return true, 0
}

// Cleanup evicts buckets untouched for longer than the TTL, until ctx ends.
//
// Without it the map grows with every distinct client address seen, which for a
// public endpoint is an unbounded memory leak.
func (s *MemoryRateLimitStore) Cleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.mu.Lock()
			for key, b := range s.buckets {
				if now.Sub(b.lastSeen) > s.ttl {
					delete(s.buckets, key)
				}
			}
			s.mu.Unlock()
		}
	}
}

// Size reports how many buckets are tracked. Used by the tests that verify
// eviction actually happens.
func (s *MemoryRateLimitStore) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.buckets)
}

// ClientIPResolver determines which address a request is attributed to.
type ClientIPResolver struct {
	trusted []*net.IPNet
}

// NewClientIPResolver parses the trusted proxy CIDR list.
func NewClientIPResolver(cidrs []string) *ClientIPResolver {
	r := &ClientIPResolver{}
	for _, c := range cidrs {
		if _, network, err := net.ParseCIDR(c); err == nil {
			r.trusted = append(r.trusted, network)
		}
	}
	return r
}

// Resolve returns the address the rate limiter should key on.
//
// Behind a load balancer, RemoteAddr is the proxy, so every user would share
// one bucket and a single heavy client would throttle everyone. X-Forwarded-For
// fixes that — but the header is client-supplied and trivially spoofed, so
// trusting it unconditionally lets any caller bypass the limiter by inventing
// an address. It is therefore believed only when the immediate peer is a
// configured proxy.
func (r *ClientIPResolver) Resolve(req *http.Request) string {
	peer := peerIP(req.RemoteAddr)

	if !r.isTrusted(peer) {
		return peer
	}

	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded == "" {
		return peer
	}

	// The header is a chain: client, proxy1, proxy2. Walking from the right,
	// the first address that is not itself a trusted proxy is the closest thing
	// to the real client that cannot have been forged by it.
	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		if net.ParseIP(candidate) == nil {
			continue
		}
		if !r.isTrusted(candidate) {
			return candidate
		}
	}
	return peer
}

func (r *ClientIPResolver) isTrusted(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, network := range r.trusted {
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}

func peerIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// RateLimit throttles requests per client address.
func RateLimit(store RateLimitStore, resolver *ClientIPResolver, cfg config.RateLimit) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Throttling a liveness or readiness probe makes an orchestrator
			// restart a perfectly healthy instance.
			if !cfg.Enabled || isProbePath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			key := resolver.Resolve(r)
			allowed, retryAfter := store.Allow(r.Context(), key)

			if !allowed {
				seconds := max(int(retryAfter.Seconds()), 1)
				w.Header().Set("Retry-After", strconv.Itoa(seconds))
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerMin))

				reqctx.Logger(r.Context()).Warn("rate limit exceeded",
					"client", key, "path", r.URL.Path, "retry_after_s", seconds)

				httperr.Write(w, r, http.StatusTooManyRequests, "Too Many Requests",
					"rate limit exceeded; retry after "+strconv.Itoa(seconds)+"s")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
