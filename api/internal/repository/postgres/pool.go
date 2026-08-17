package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
)

// NewPool creates the connection pool and blocks until the database answers or
// the context is cancelled.
//
// Opening a TCP connection and authenticating costs milliseconds, so a pool
// establishes a set of connections once and lends them to requests. Without
// one, every request pays that cost and a traffic spike opens connections until
// the server refuses them.
//
// Sizing is a deployment-wide concern: replicas * MaxConns plus migration jobs
// and admin sessions must stay under the server's max_connections, which
// defaults to 100.
func NewPool(ctx context.Context, cfg config.Database, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing database DSN: %w", err)
	}

	poolCfg.MaxConns = int32(cfg.MaxConns)
	poolCfg.MinConns = int32(cfg.MinConns)
	// Recycling connections lets a rolling restart of the database drain
	// cleanly and stops any single connection accumulating server-side state
	// indefinitely.
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	// A server-side ceiling on query duration. Without it a pathological query
	// holds its connection until the client goes away, and the pool drains.
	poolCfg.ConnConfig.RuntimeParams["statement_timeout"] =
		fmt.Sprintf("%d", cfg.StatementTimeout.Milliseconds())
	poolCfg.ConnConfig.RuntimeParams["application_name"] = "pokedex-api"

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := waitForDatabase(ctx, pool, cfg.ConnectTimeout, logger); err != nil {
		pool.Close()
		return nil, err
	}

	logger.Info("database pool ready",
		"host", cfg.Host,
		"database", cfg.Name,
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
	)
	return pool, nil
}

// waitForDatabase pings with exponential backoff.
//
// Both Compose and Kubernetes start the API before PostgreSQL has finished
// initialising, so the first attempt on a cold environment always fails.
// Backing off rather than retrying at a fixed interval keeps a recovering
// database from being knocked over by the retry traffic itself.
func waitForDatabase(ctx context.Context, pool *pgxpool.Pool, budget time.Duration, logger *slog.Logger) error {
	deadline := time.Now().Add(budget)
	delay := 100 * time.Millisecond
	const maxDelay = 3 * time.Second

	for attempt := 1; ; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err := pool.Ping(pingCtx)
		cancel()

		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("database unreachable after %s: %w", budget, err)
		}

		logger.Warn("database not ready, retrying",
			"attempt", attempt,
			"retry_in", delay.String(),
			"error", err.Error(),
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		if delay *= 2; delay > maxDelay {
			delay = maxDelay
		}
	}
}
