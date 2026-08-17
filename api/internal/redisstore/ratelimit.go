// Package redisstore holds the Redis-backed implementations of state that must
// be shared between replicas.
//
// Redis is not here to make Postgres reads faster: the catalogue is ~1300
// immutable rows that PostgreSQL serves from shared_buffers in well under a
// millisecond, and putting a network hop in front of that is plausibly a net
// loss. It is here because process-local state becomes incorrect the moment
// there is more than one process — see RateLimiter.
package redisstore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
)

// Client wraps the Redis connection.
type Client struct {
	rdb *redis.Client
	log *slog.Logger
}

// New connects to Redis and verifies it answers.
func New(ctx context.Context, cfg config.Redis, log *slog.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		// Bounded so a hung Redis degrades one request rather than occupying a
		// handler until its own timeout fires.
		DialTimeout:  3 * time.Second,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		PoolSize:     10,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("connecting to redis at %s: %w", cfg.Addr, err)
	}

	log.Info("redis connected", "addr", cfg.Addr, "db", cfg.DB)
	return &Client{rdb: rdb, log: log}, nil
}

// Close releases the connection pool.
func (c *Client) Close() error { return c.rdb.Close() }

// Ping reports whether Redis is reachable.
func (c *Client) Ping(ctx context.Context) error { return c.rdb.Ping(ctx).Err() }

// tokenBucketScript implements the whole read-modify-write in one atomic step.
//
// Atomicity is the reason this is a script rather than Go code issuing GET and
// SET. Two replicas interleaving a read and a write both observe the same
// starting value and both allow the request, so the counter loses increments
// under exactly the concurrent load a rate limiter exists to control. Redis
// executes a script to completion without interleaving anything else, which
// removes the race by construction.
//
// KEYS[1]  bucket key
// ARGV[1]  capacity (burst)
// ARGV[2]  refill rate, tokens per second
// ARGV[3]  current time in milliseconds
// ARGV[4]  key TTL in seconds
//
// Returns { allowed, retry_after_ms }.
var tokenBucketScript = redis.NewScript(`
local key      = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate     = tonumber(ARGV[2])
local now_ms   = tonumber(ARGV[3])
local ttl      = tonumber(ARGV[4])

local bucket = redis.call('HMGET', key, 'tokens', 'ts')
local tokens = tonumber(bucket[1])
local ts     = tonumber(bucket[2])

if tokens == nil then
  tokens = capacity
  ts = now_ms
end

-- Refill for the time elapsed since this client was last seen, capped at the
-- bucket's capacity so an idle client cannot accumulate an unbounded burst.
local elapsed_s = math.max(0, (now_ms - ts) / 1000)
tokens = math.min(capacity, tokens + elapsed_s * rate)

local allowed = 0
local retry_after_ms = 0

if tokens >= 1 then
  allowed = 1
  tokens = tokens - 1
else
  retry_after_ms = math.ceil(((1 - tokens) / rate) * 1000)
end

redis.call('HSET', key, 'tokens', tokens, 'ts', now_ms)
-- Every write refreshes the TTL, so idle clients expire on their own and the
-- keyspace cannot grow without bound.
redis.call('EXPIRE', key, ttl)

return { allowed, retry_after_ms }
`)

// RateLimiter is a token bucket whose state lives in Redis.
//
// The in-process limiter is correct with exactly one instance. With more, a
// Service spreads a client's connections across pods, each pod counts only what
// it happens to receive, and the effective limit becomes N times the configured
// one — a number that also moves on its own whenever the autoscaler acts.
// Moving the counters here is a correctness fix for a reproducible failure, not
// a performance optimisation.
type RateLimiter struct {
	client   *Client
	capacity float64
	// refill is tokens per second.
	refill float64
	ttl    time.Duration
	log    *slog.Logger
}

// NewRateLimiter builds a Redis-backed limiter.
func NewRateLimiter(client *Client, requestsPerMinute, burst int, log *slog.Logger) *RateLimiter {
	return &RateLimiter{
		client:   client,
		capacity: float64(burst),
		refill:   float64(requestsPerMinute) / 60.0,
		ttl:      10 * time.Minute,
		log:      log,
	}
}

// Allow implements middleware.RateLimitStore.
//
// It fails open: when Redis is unavailable the request is allowed and the
// failure is logged. Rejecting all traffic because a rate limiter's backing
// store is down converts a degraded dependency into a full outage, which is a
// far worse failure than briefly serving above the configured rate.
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	result, err := tokenBucketScript.Run(ctx, r.client.rdb,
		[]string{"ratelimit:" + key},
		r.capacity,
		r.refill,
		time.Now().UnixMilli(),
		int(r.ttl.Seconds()),
	).Int64Slice()

	if err != nil {
		r.log.Error("rate limiter store unavailable, failing open",
			"error", err, "client", key)
		return true, 0
	}
	if len(result) != 2 {
		r.log.Error("unexpected rate limiter response, failing open", "values", len(result))
		return true, 0
	}

	if result[0] == 1 {
		return true, 0
	}
	return false, time.Duration(result[1]) * time.Millisecond
}
