// Package service holds business logic and orchestration.
//
// The repository interfaces are declared here rather than next to their
// implementations. That is the Go convention: the consumer owns the
// abstraction, so internal/repository/postgres satisfies these interfaces
// without importing this package, and swapping the storage engine never touches
// the service.
package service

import (
	"context"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// PokemonRepository is the persistence contract for the catalogue.
//
// Search takes a plain term, not a SQL pattern: building the LIKE expression
// and escaping its metacharacters is knowledge that belongs to the SQL layer.
type PokemonRepository interface {
	List(ctx context.Context, limit, offset int32) ([]model.PokemonListItem, error)
	SearchByName(ctx context.Context, term string, limit, offset int32) ([]model.PokemonListItem, error)
	ListByType(ctx context.Context, typeName string, limit, offset int32) ([]model.PokemonListItem, error)
	GetByID(ctx context.Context, id int32) (model.Pokemon, error)
	Count(ctx context.Context) (int64, error)
	CountByName(ctx context.Context, term string) (int64, error)
	CountByType(ctx context.Context, typeName string) (int64, error)
	GetSpeciesByPokemonID(ctx context.Context, pokemonID int32) (model.Species, error)
}

// TypeRepository is the persistence contract for elemental types.
type TypeRepository interface {
	List(ctx context.Context) ([]model.Type, error)
	GetMatchups(ctx context.Context, name string) (model.Matchups, error)
}

// HealthRepository reports whether the storage backend is reachable. It backs
// the readiness probe, which answers whether this instance should receive
// traffic.
type HealthRepository interface {
	Ping(ctx context.Context) error
	DatasetVersion(ctx context.Context) (int64, error)
}
