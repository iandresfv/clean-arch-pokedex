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

// TypeRepository reads elemental types and their effectiveness matrix.
type TypeRepository struct {
	q *Queries
}

// NewTypeRepository wires the generated queries onto a pool.
func NewTypeRepository(pool *pgxpool.Pool) *TypeRepository {
	return &TypeRepository{q: New(pool)}
}

// List implements service.TypeRepository.
func (r *TypeRepository) List(ctx context.Context) ([]model.Type, error) {
	rows, err := r.q.ListTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing types: %w", err)
	}

	types := make([]model.Type, 0, len(rows))
	for _, row := range rows {
		types = append(types, model.Type{ID: row.ID, Name: row.Name})
	}
	return types, nil
}

// GetMatchups returns the defensive profile of a type: which attacking types
// hit it for double, half, or no damage.
func (r *TypeRepository) GetMatchups(ctx context.Context, name string) (model.Matchups, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	// Checked first so an unknown type is a 404 rather than an empty matchup
	// set, which would be indistinguishable from a type with no interactions.
	if _, err := r.q.GetTypeByName(ctx, name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Matchups{}, fmt.Errorf("type %q: %w", name, model.ErrTypeNotFound)
		}
		return model.Matchups{}, fmt.Errorf("looking up type %q: %w", name, err)
	}

	rows, err := r.q.GetTypeMatchups(ctx, name)
	if err != nil {
		return model.Matchups{}, fmt.Errorf("getting matchups for type %q: %w", name, err)
	}

	matchups := model.Matchups{
		Type:        name,
		Weaknesses:  []string{},
		Resistances: []string{},
		Immunities:  []string{},
	}
	for _, row := range rows {
		switch row.Multiplier {
		case 2:
			matchups.Weaknesses = append(matchups.Weaknesses, row.AttackingType)
		case 0.5:
			matchups.Resistances = append(matchups.Resistances, row.AttackingType)
		case 0:
			matchups.Immunities = append(matchups.Immunities, row.AttackingType)
		}
	}
	return matchups, nil
}
