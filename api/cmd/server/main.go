// Command server runs the Pokedex HTTP API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/config"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/handler"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/repository/postgres"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/router"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/service"
)

// version is injected at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	// All work happens in run so that deferred cleanup executes. Calling
	// os.Exit anywhere else would skip every pending defer.
	if err := run(); err != nil {
		slog.Error("server terminated", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := newLogger(cfg)
	slog.SetDefault(logger)

	// NotifyContext cancels ctx on SIGINT/SIGTERM. SIGTERM is what a container
	// runtime sends before killing the process.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Composition root: every dependency is constructed once, here, and passed
	// explicitly. This is the Go equivalent of the client's createContainer —
	// a DI framework would trade these few lines for reflection at startup.
	pool, err := postgres.NewPool(ctx, cfg.Database, logger)
	if err != nil {
		return err
	}
	defer pool.Close()

	pokemonRepo := postgres.NewPokemonRepository(pool)
	typeRepo := postgres.NewTypeRepository(pool)

	pokemonSvc := service.NewPokemonService(pokemonRepo, logger)
	typeSvc := service.NewTypeService(typeRepo, logger)

	mux := router.New(router.Handlers{
		Pokemon: handler.NewPokemonHandler(pokemonSvc),
		Type:    handler.NewTypeHandler(typeSvc),
		Health:  handler.NewHealthHandler(pokemonRepo, version),
	})

	srv := &http.Server{
		Addr:              cfg.Server.Addr(),
		Handler:           mux,
		ReadTimeout:       cfg.Server.ReadTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
		// net/http writes its internal errors to a *log.Logger. Routing it
		// through slog keeps every line in the structured stream.
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	return serve(ctx, srv, cfg, logger)
}

// serve runs srv until ctx is cancelled or the listener fails, then shuts it
// down gracefully.
func serve(ctx context.Context, srv *http.Server, cfg *config.Config, logger *slog.Logger) error {
	// Buffered so the goroutine can exit even when nobody reads the value,
	// which happens on the normal shutdown path.
	listenErr := make(chan error, 1)

	go func() {
		logger.Info("server starting",
			"addr", srv.Addr,
			"env", cfg.Env,
			"version", version,
			"tls", cfg.TLS.Enabled,
		)

		var err error
		if cfg.TLS.Enabled {
			err = srv.ListenAndServeTLS(cfg.TLS.CertPath, cfg.TLS.KeyPath)
		} else {
			err = srv.ListenAndServe()
		}
		// ErrServerClosed is the expected result of a graceful Shutdown.
		// errors.Is is required because the error may arrive wrapped.
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- fmt.Errorf("listening on %s: %w", srv.Addr, err)
			return
		}
		listenErr <- nil
	}()

	select {
	case err := <-listenErr:
		// The listener failed before any signal arrived, typically because the
		// port is already bound.
		return err
	case <-ctx.Done():
		// Duration formats as raw nanoseconds under slog, so render it here.
		logger.Info("shutdown signal received", "grace_period", cfg.Server.ShutdownTimeout.String())
	}

	// A fresh context: ctx is already cancelled, so reusing it would abort the
	// shutdown immediately.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// nolint:contextcheck // Detaching from ctx is the point: ctx is already
	// cancelled by the signal, so inheriting it would abort the drain instantly.
	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Deadline exceeded means requests were still in flight. Close forces
		// the remaining connections shut so the process can exit.
		_ = srv.Close()
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if err := <-listenErr; err != nil {
		return err
	}

	logger.Info("server stopped")
	return nil
}

func newLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.Log.Level}

	var handler slog.Handler
	if cfg.Log.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With("service", "pokedex-api")
}
