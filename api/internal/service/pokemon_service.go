package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// PokemonService holds the catalogue's business rules.
//
// All input validation happens at this layer: not in the handler, which only
// translates HTTP, and not in the repository, which only translates SQL. Any
// conditional about what constitutes a valid request belongs here.
type PokemonService struct {
	repo PokemonRepository
	log  *slog.Logger
}

// NewPokemonService wires a repository and logger into the service.
func NewPokemonService(repo PokemonRepository, log *slog.Logger) *PokemonService {
	return &PokemonService{repo: repo, log: log}
}

// List returns a page of the catalogue.
func (s *PokemonService) List(ctx context.Context, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
	var empty model.PaginatedResult[model.PokemonListItem]

	total, err := s.repo.Count(ctx)
	if err != nil {
		return empty, fmt.Errorf("listing pokemon: %w", err)
	}

	// Past the last page the repository would return an empty slice anyway;
	// skipping the query saves a round trip and keeps the response shape
	// identical.
	if p.Offset() >= int32(total) {
		return model.NewPaginatedResult([]model.PokemonListItem{}, total, p.Page, p.Limit), nil
	}

	items, err := s.repo.List(ctx, p.LimitInt32(), p.Offset())
	if err != nil {
		return empty, fmt.Errorf("listing pokemon: %w", err)
	}

	return model.NewPaginatedResult(items, total, p.Page, p.Limit), nil
}

// SearchByName returns the page of Pokemon whose name contains term.
func (s *PokemonService) SearchByName(ctx context.Context, term string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
	var empty model.PaginatedResult[model.PokemonListItem]

	// Normalised here, once, so the repository always receives the same shape
	// and the count and the page agree on what was searched.
	term = strings.ToLower(strings.TrimSpace(term))

	if len(term) < model.MinSearchLength {
		return empty, fmt.Errorf("%w: search term must be at least %d characters",
			model.ErrInvalidQuery, model.MinSearchLength)
	}

	total, err := s.repo.CountByName(ctx, term)
	if err != nil {
		return empty, fmt.Errorf("searching pokemon %q: %w", term, err)
	}
	if total == 0 || p.Offset() >= int32(total) {
		return model.NewPaginatedResult([]model.PokemonListItem{}, total, p.Page, p.Limit), nil
	}

	items, err := s.repo.SearchByName(ctx, term, p.LimitInt32(), p.Offset())
	if err != nil {
		return empty, fmt.Errorf("searching pokemon %q: %w", term, err)
	}

	return model.NewPaginatedResult(items, total, p.Page, p.Limit), nil
}

// ListByType returns the page of Pokemon belonging to a type.
func (s *PokemonService) ListByType(ctx context.Context, typeName string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
	var empty model.PaginatedResult[model.PokemonListItem]

	typeName = strings.ToLower(strings.TrimSpace(typeName))
	if typeName == "" {
		return empty, fmt.Errorf("%w: type must not be empty", model.ErrInvalidQuery)
	}

	total, err := s.repo.CountByType(ctx, typeName)
	if err != nil {
		return empty, fmt.Errorf("listing pokemon of type %q: %w", typeName, err)
	}
	if total == 0 {
		// An unknown type and a type with no members are different failures:
		// the first is a client error, the second an empty page. Only the
		// former should reach the client as 404, which the type service
		// decides — here an empty page is the honest answer.
		return model.NewPaginatedResult([]model.PokemonListItem{}, 0, p.Page, p.Limit), nil
	}
	if p.Offset() >= int32(total) {
		return model.NewPaginatedResult([]model.PokemonListItem{}, total, p.Page, p.Limit), nil
	}

	items, err := s.repo.ListByType(ctx, typeName, p.LimitInt32(), p.Offset())
	if err != nil {
		return empty, fmt.Errorf("listing pokemon of type %q: %w", typeName, err)
	}

	return model.NewPaginatedResult(items, total, p.Page, p.Limit), nil
}

// GetByID returns the full detail for one Pokemon.
func (s *PokemonService) GetByID(ctx context.Context, id int32) (model.Pokemon, error) {
	if id < 1 {
		return model.Pokemon{}, fmt.Errorf("%w: id must be positive, got %d", model.ErrInvalidID, id)
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Pokemon{}, fmt.Errorf("getting pokemon %d: %w", id, err)
	}
	return p, nil
}

// GetSpecies returns the species a Pokemon belongs to.
func (s *PokemonService) GetSpecies(ctx context.Context, pokemonID int32) (model.Species, error) {
	if pokemonID < 1 {
		return model.Species{}, fmt.Errorf("%w: id must be positive, got %d", model.ErrInvalidID, pokemonID)
	}

	sp, err := s.repo.GetSpeciesByPokemonID(ctx, pokemonID)
	if err != nil {
		return model.Species{}, fmt.Errorf("getting species for pokemon %d: %w", pokemonID, err)
	}
	return sp, nil
}
