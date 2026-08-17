package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// Pinger reports whether a downstream dependency is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler serves the liveness and readiness probes.
//
// The two are deliberately different checks. Liveness answers "should this
// process be restarted?" and must not touch the database: if PostgreSQL is
// down, restarting the API fixes nothing and a restart loop makes recovery
// harder. Readiness answers "should this instance receive traffic?" and does
// check the database, so an instance that cannot serve is removed from the load
// balancer while staying alive to recover.
type HealthHandler struct {
	db      Pinger
	version string
}

// NewHealthHandler wires the database probe and build version.
func NewHealthHandler(db Pinger, version string) *HealthHandler {
	return &HealthHandler{db: db, version: version}
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}

type readyResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks"`
	Version  string            `json:"version,omitempty"`
	Duration string            `json:"duration"`
}

// Health serves GET /health. It reports on the process alone.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, healthResponse{Status: "ok", Version: h.version})
}

// Ready serves GET /ready. It verifies every dependency required to serve
// traffic and returns 503 when any of them is unreachable.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Bounded independently of the request: a probe that hangs as long as the
	// handler timeout allows would keep the kubelet waiting and delay the
	// instance being pulled from rotation.
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}
	status := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		reqctx.Logger(r.Context()).Warn("readiness check failed", "dependency", "postgres", "error", err)
		checks["postgres"] = "unavailable"
		status = http.StatusServiceUnavailable
	} else {
		checks["postgres"] = "ok"
	}

	body := readyResponse{
		Status:   "ready",
		Checks:   checks,
		Version:  h.version,
		Duration: time.Since(start).String(),
	}
	if status != http.StatusOK {
		body.Status = "not ready"
	}

	writeJSON(w, r, status, body)
}
