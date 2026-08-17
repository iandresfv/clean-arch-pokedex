package handler

import (
	"context"
	"net/http"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// PokemonServicePort is the slice of the service this handler needs.
//
// Declaring it here rather than accepting *service.PokemonService keeps the
// handler testable with a small fake and documents exactly which operations the
// HTTP layer depends on.
type PokemonServicePort interface {
	List(ctx context.Context, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error)
	SearchByName(ctx context.Context, term string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error)
	ListByType(ctx context.Context, typeName string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error)
	GetByID(ctx context.Context, id int32) (model.Pokemon, error)
	GetSpecies(ctx context.Context, pokemonID int32) (model.Species, error)
}

// PokemonHandler serves the catalogue endpoints.
type PokemonHandler struct {
	svc PokemonServicePort
}

// NewPokemonHandler wires a service into the handler.
func NewPokemonHandler(svc PokemonServicePort) *PokemonHandler {
	return &PokemonHandler{svc: svc}
}

// List serves GET /api/v1/pokemon.
func (h *PokemonHandler) List(w http.ResponseWriter, r *http.Request) {
	if err := rejectUnknownParams(r, "page", "limit", "type"); err != nil {
		handleServiceError(w, r, err)
		return
	}

	p, err := parsePagination(r)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	// A type filter is a variation of the listing rather than its own endpoint:
	// same shape, same pagination, one optional predicate.
	var result model.PaginatedResult[model.PokemonListItem]
	if typeName := r.URL.Query().Get("type"); typeName != "" {
		result, err = h.svc.ListByType(r.Context(), typeName, p)
	} else {
		result, err = h.svc.List(r.Context(), p)
	}
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, result)
}

// Search serves GET /api/v1/pokemon/search.
func (h *PokemonHandler) Search(w http.ResponseWriter, r *http.Request) {
	if err := rejectUnknownParams(r, "q", "page", "limit"); err != nil {
		handleServiceError(w, r, err)
		return
	}

	p, err := parsePagination(r)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	result, err := h.svc.SearchByName(r.Context(), r.URL.Query().Get("q"), p)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, result)
}

// GetByID serves GET /api/v1/pokemon/{id}.
func (h *PokemonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	if err := rejectUnknownParams(r); err != nil {
		handleServiceError(w, r, err)
		return
	}

	id, err := pathID(r, "id")
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, p)
}

// GetSpecies serves GET /api/v1/pokemon/{id}/species.
func (h *PokemonHandler) GetSpecies(w http.ResponseWriter, r *http.Request) {
	if err := rejectUnknownParams(r); err != nil {
		handleServiceError(w, r, err)
		return
	}

	id, err := pathID(r, "id")
	if err != nil {
		handleServiceError(w, r, err)
		return
	}

	s, err := h.svc.GetSpecies(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err)
		return
	}
	writeJSON(w, r, http.StatusOK, s)
}
