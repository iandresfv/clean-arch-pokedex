package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// TypeService exposes the elemental type reference data.
//
// Note that the client keeps its own copy of the effectiveness chart in a
// domain service. That duplication is deliberate: the client's chart is a pure
// domain rule with full test coverage, while this endpoint makes the database
// the system of record and serves consumers that are not the browser. A test in
// the seeding phase asserts the two representations agree.
type TypeService struct {
	repo TypeRepository
	log  *slog.Logger
}

// NewTypeService wires a repository and logger into the service.
func NewTypeService(repo TypeRepository, log *slog.Logger) *TypeService {
	return &TypeService{repo: repo, log: log}
}

// List returns all eighteen types.
func (s *TypeService) List(ctx context.Context) ([]model.Type, error) {
	types, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing types: %w", err)
	}
	return types, nil
}

// GetMatchups returns the defensive profile of a type.
func (s *TypeService) GetMatchups(ctx context.Context, name string) (model.Matchups, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return model.Matchups{}, fmt.Errorf("%w: type name must not be empty", model.ErrInvalidQuery)
	}

	m, err := s.repo.GetMatchups(ctx, name)
	if err != nil {
		return model.Matchups{}, fmt.Errorf("getting matchups for %q: %w", name, err)
	}
	return m, nil
}
