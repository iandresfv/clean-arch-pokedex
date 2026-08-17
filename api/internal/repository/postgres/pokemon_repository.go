package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// PokemonRepository reads the catalogue from PostgreSQL. It satisfies
// service.PokemonRepository without importing that package.
type PokemonRepository struct {
	q    *Queries
	pool *pgxpool.Pool
}

// NewPokemonRepository wires the generated queries onto a pool.
func NewPokemonRepository(pool *pgxpool.Pool) *PokemonRepository {
	return &PokemonRepository{q: New(pool), pool: pool}
}

// likeEscaper neutralises the metacharacters LIKE would otherwise interpret.
// Without it, searching for "50%" matches everything starting with "50", and a
// term of "_" matches every single-character name.
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// searchPattern builds the substring pattern. The term is lowercased here so
// the predicate matches the functional index on lower(name) exactly; any
// mismatch silently degrades the query to a sequential scan.
func searchPattern(term string) string {
	return "%" + likeEscaper.Replace(strings.ToLower(strings.TrimSpace(term))) + "%"
}

// List implements service.PokemonRepository.
func (r *PokemonRepository) List(ctx context.Context, limit, offset int32) ([]model.PokemonListItem, error) {
	rows, err := r.q.ListPokemon(ctx, ListPokemonParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("listing pokemon: %w", err)
	}

	items := make([]model.PokemonListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.PokemonListItem{
			ID:        row.ID,
			Name:      row.Name,
			Types:     row.TypeNames,
			SpriteURL: preferredSprite(row.SpriteOfficialArtwork, row.SpriteFrontDefault),
		})
	}
	return items, nil
}

// SearchByName implements service.PokemonRepository.
func (r *PokemonRepository) SearchByName(ctx context.Context, term string, limit, offset int32) ([]model.PokemonListItem, error) {
	rows, err := r.q.SearchPokemonByName(ctx, SearchPokemonByNameParams{
		Name:   searchPattern(term),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("searching pokemon by name %q: %w", term, err)
	}

	items := make([]model.PokemonListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.PokemonListItem{
			ID:        row.ID,
			Name:      row.Name,
			Types:     row.TypeNames,
			SpriteURL: preferredSprite(row.SpriteOfficialArtwork, row.SpriteFrontDefault),
		})
	}
	return items, nil
}

// ListByType implements service.PokemonRepository.
func (r *PokemonRepository) ListByType(ctx context.Context, typeName string, limit, offset int32) ([]model.PokemonListItem, error) {
	rows, err := r.q.ListPokemonByType(ctx, ListPokemonByTypeParams{
		Name:   strings.ToLower(typeName),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pokemon of type %q: %w", typeName, err)
	}

	items := make([]model.PokemonListItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.PokemonListItem{
			ID:        row.ID,
			Name:      row.Name,
			Types:     row.TypeNames,
			SpriteURL: preferredSprite(row.SpriteOfficialArtwork, row.SpriteFrontDefault),
		})
	}
	return items, nil
}

// GetByID implements service.PokemonRepository.
func (r *PokemonRepository) GetByID(ctx context.Context, id int32) (model.Pokemon, error) {
	row, err := r.q.GetPokemonByID(ctx, id)
	if err != nil {
		// Translating the driver's sentinel here is what keeps the service
		// database-agnostic. If pgx.ErrNoRows leaked upward, the service would
		// import the driver and changing storage would mean rewriting it.
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Pokemon{}, fmt.Errorf("pokemon %d: %w", id, model.ErrPokemonNotFound)
		}
		return model.Pokemon{}, fmt.Errorf("getting pokemon %d: %w", id, err)
	}

	return model.Pokemon{
		ID:             row.ID,
		Name:           row.Name,
		Types:          row.TypeNames,
		HeightDm:       row.HeightDm,
		WeightHg:       row.WeightHg,
		BaseExperience: row.BaseExperience,
		Stats: model.Stats{
			HP:             row.StatHp,
			Attack:         row.StatAttack,
			Defense:        row.StatDefense,
			SpecialAttack:  row.StatSpecialAttack,
			SpecialDefense: row.StatSpecialDefense,
			Speed:          row.StatSpeed,
		},
		Sprites: model.Sprites{
			FrontDefault:    row.SpriteFrontDefault,
			FrontShiny:      row.SpriteFrontShiny,
			BackDefault:     row.SpriteBackDefault,
			BackShiny:       row.SpriteBackShiny,
			OfficialArtwork: row.SpriteOfficialArtwork,
		},
		Species: &model.Species{
			ID:          row.SpeciesID,
			Name:        row.SpeciesName,
			Generation:  row.SpeciesGeneration,
			FlavorText:  row.SpeciesFlavorText,
			Habitat:     row.SpeciesHabitat,
			IsLegendary: row.SpeciesIsLegendary,
			IsMythical:  row.SpeciesIsMythical,
		},
	}, nil
}

// Count implements service.PokemonRepository.
func (r *PokemonRepository) Count(ctx context.Context) (int64, error) {
	total, err := r.q.CountPokemon(ctx)
	if err != nil {
		return 0, fmt.Errorf("counting pokemon: %w", err)
	}
	return total, nil
}

// CountByName implements service.PokemonRepository.
func (r *PokemonRepository) CountByName(ctx context.Context, term string) (int64, error) {
	total, err := r.q.CountPokemonByName(ctx, searchPattern(term))
	if err != nil {
		return 0, fmt.Errorf("counting pokemon matching %q: %w", term, err)
	}
	return total, nil
}

// CountByType implements service.PokemonRepository.
func (r *PokemonRepository) CountByType(ctx context.Context, typeName string) (int64, error) {
	total, err := r.q.CountPokemonByType(ctx, strings.ToLower(typeName))
	if err != nil {
		return 0, fmt.Errorf("counting pokemon of type %q: %w", typeName, err)
	}
	return total, nil
}

// GetSpeciesByPokemonID implements service.PokemonRepository.
func (r *PokemonRepository) GetSpeciesByPokemonID(ctx context.Context, pokemonID int32) (model.Species, error) {
	row, err := r.q.GetSpeciesByPokemonID(ctx, pokemonID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Species{}, fmt.Errorf("species for pokemon %d: %w", pokemonID, model.ErrSpeciesNotFound)
		}
		return model.Species{}, fmt.Errorf("getting species for pokemon %d: %w", pokemonID, err)
	}

	return model.Species{
		ID:          row.ID,
		Name:        row.Name,
		Generation:  row.Generation,
		FlavorText:  row.FlavorText,
		Habitat:     row.Habitat,
		Color:       row.Color,
		Shape:       row.Shape,
		IsLegendary: row.IsLegendary,
		IsMythical:  row.IsMythical,
	}, nil
}

// Ping reports whether the pool can reach the database. It backs /ready.
func (r *PokemonRepository) Ping(ctx context.Context) error {
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	return nil
}

// DatasetVersion returns the counter the seeder bumps, used to namespace cache
// keys so a reseed invalidates every entry at once.
func (r *PokemonRepository) DatasetVersion(ctx context.Context) (int64, error) {
	v, err := r.q.GetDatasetVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("reading dataset version: %w", err)
	}
	return v, nil
}

// preferredSprite picks the highest-quality artwork available, mirroring the
// client's Sprites.getBestQuality so both data sources render identically.
func preferredSprite(candidates ...*string) *string {
	for _, c := range candidates {
		if c != nil && *c != "" {
			return c
		}
	}
	return nil
}
