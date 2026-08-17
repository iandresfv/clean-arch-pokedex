// Command seeder populates the database from PokeAPI.
//
// Run with: go run ./cmd/seeder/ --limit=1302 --concurrency=8
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/repository/postgres"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/seeder"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seeding failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		limit       = flag.Int("limit", 1302, "how many Pokemon to seed (the full national dex is 1302)")
		concurrency = flag.Int("concurrency", 8, "parallel upstream requests")
		dryRun      = flag.Bool("dry-run", false, "fetch and validate without writing")
		baseURL     = flag.String("base-url", seeder.DefaultBaseURL, "PokeAPI base URL")
	)
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Log.Level})).
		With("service", "pokedex-seeder")
	slog.SetDefault(logger)

	// Ctrl-C leaves the database untouched: the write happens in a single
	// transaction that is only committed at the end.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.Database, logger)
	if err != nil {
		return err
	}
	defer pool.Close()

	s := seeder.New(seeder.NewClient(*baseURL, logger), pool, logger)

	logger.Info("seeding started",
		"limit", *limit, "concurrency", *concurrency, "dry_run", *dryRun, "source", *baseURL)

	result, err := s.Run(ctx, seeder.Options{
		Limit:       *limit,
		Concurrency: *concurrency,
		DryRun:      *dryRun,
	})
	if err != nil {
		return err
	}

	logger.Info("seeding complete",
		"fetched", result.Fetched,
		"inserted", result.Inserted,
		"skipped", result.Skipped,
		"dataset_version", result.Version,
		"duration", result.Duration.Round(time.Millisecond).String(),
	)
	return nil
}
