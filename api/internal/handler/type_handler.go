package handler

import (
	"context"
	"net/http"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// TypeServicePort is the slice of the type service this handler needs.
type TypeServicePort interface {
	List(ctx context.Context) ([]model.Type, error)
	GetMatchups(ctx context.Context, name string) (model.Matchups, error)
}

// TypeHandler serves the elemental type endpoints.
type TypeHandler struct {
	svc TypeServicePort
}

// NewTypeHandler wires a service into the handler.
func NewTypeHandler(svc TypeServicePort) *TypeHandler {
	return &TypeHandler{svc: svc}
}

// List serves GET /api/v1/types.
func (h *TypeHandler) List(w http.ResponseWriter, r *http.Request) {
	if err := rejectUnknownParams(r); err != nil {
		handleServiceError(w, r, err)
		return
	}

	types, err := h.svc.List(r.Context())
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, types)
}

// GetMatchups serves GET /api/v1/types/{name}/matchups.
func (h *TypeHandler) GetMatchups(w http.ResponseWriter, r *http.Request) {
	if err := rejectUnknownParams(r); err != nil {
		handleServiceError(w, r, err)
		return
	}

	m, err := h.svc.GetMatchups(r.Context(), r.PathValue("name"))
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, m)
}
