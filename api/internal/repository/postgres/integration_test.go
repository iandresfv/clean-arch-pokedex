package postgres

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// These tests run against a real PostgreSQL instance.
//
// A mocked database proves the Go code calls the repository correctly; it
// cannot prove the SQL is valid, that a constraint fires, that array_agg
// preserves slot order, or that an index is used. Those need the real engine.
//
// The instance is addressed through TEST_DATABASE_URL rather than started by
// testcontainers: the project already runs PostgreSQL in Compose and CI
// provides a service container, so pulling in a container-orchestration
// dependency would add a large dependency tree to replace one environment
// variable.

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/pokedex?sslmode=disable"
	}

	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Skipf("no test database available (%v)", err)
	}
	if err := pool.Ping(t.Context()); err != nil {
		pool.Close()
		t.Skipf("test database unreachable at %s (%v)", dsn, err)
	}
	t.Cleanup(pool.Close)

	// The suite reads seeded data rather than creating its own, so a missing
	// dataset is a skip, not a failure: the migrations may be applied without
	// the seeder having run.
	var count int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM pokemon").Scan(&count); err != nil {
		t.Skipf("schema not migrated (%v)", err)
	}
	if count == 0 {
		t.Skip("database is empty; run `make seed` before the integration suite")
	}

	return pool
}

func TestGetByIDReturnsCompleteDetail(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))

	got, err := repo.GetByID(t.Context(), 1)
	if err != nil {
		t.Fatalf("GetByID(1) error = %v", err)
	}

	if got.Name != "bulbasaur" {
		t.Errorf("Name = %q, want bulbasaur", got.Name)
	}
	// Slot order is meaningful: the UI renders the primary type first, and
	// array_agg must preserve it.
	want := []string{"grass", "poison"}
	if len(got.Types) != 2 || got.Types[0] != want[0] || got.Types[1] != want[1] {
		t.Errorf("Types = %v, want %v (slot order must be preserved)", got.Types, want)
	}
	if got.Stats.HP == 0 || got.Stats.Speed == 0 {
		t.Errorf("Stats look unpopulated: %+v", got.Stats)
	}
	if got.Species == nil {
		t.Fatal("Species is nil; the detail query must join it")
	}
	if got.Species.Generation != 1 {
		t.Errorf("Species.Generation = %d, want 1", got.Species.Generation)
	}
}

func TestGetByIDTranslatesNotFound(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))

	_, err := repo.GetByID(t.Context(), 999999)

	// The driver's pgx.ErrNoRows must never escape this layer, or the service
	// would have to import the database driver to interpret it.
	if !errors.Is(err, model.ErrPokemonNotFound) {
		t.Fatalf("error = %v, want it to wrap model.ErrPokemonNotFound", err)
	}
	if strings.Contains(err.Error(), "no rows") {
		t.Errorf("driver error leaked through: %v", err)
	}
}

func TestListPaginationBoundaries(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))
	ctx := t.Context()

	total, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}

	tests := []struct {
		name      string
		limit     int32
		offset    int32
		wantCount int
	}{
		{name: "first page", limit: 20, offset: 0, wantCount: 20},
		{name: "single item", limit: 1, offset: 0, wantCount: 1},
		{name: "past the end", limit: 20, offset: int32(total) + 100, wantCount: 0},
		{name: "last partial page", limit: 20, offset: int32(total) - 5, wantCount: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := repo.List(ctx, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(items) != tt.wantCount {
				t.Errorf("got %d items, want %d", len(items), tt.wantCount)
			}
			// An empty result must be an empty slice, never nil: nil marshals
			// to JSON null and breaks clients iterating the page.
			if items == nil {
				t.Error("List returned nil; want an empty slice")
			}
		})
	}
}

func TestSearchByName(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))
	ctx := t.Context()

	tests := []struct {
		name        string
		term        string
		wantMinimum int
		wantContain string
	}{
		{name: "substring in the middle", term: "chu", wantMinimum: 1, wantContain: "pikachu"},
		{name: "prefix", term: "char", wantMinimum: 3, wantContain: "charizard"},
		{name: "case insensitive", term: "PIKA", wantMinimum: 1, wantContain: "pikachu"},
		{name: "no matches", term: "zzzzzz", wantMinimum: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := repo.SearchByName(ctx, tt.term, 50, 0)
			if err != nil {
				t.Fatalf("SearchByName(%q) error = %v", tt.term, err)
			}
			if len(items) < tt.wantMinimum {
				t.Errorf("got %d results for %q, want at least %d", len(items), tt.term, tt.wantMinimum)
			}
			if tt.wantContain == "" {
				return
			}
			found := false
			for _, it := range items {
				if it.Name == tt.wantContain {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("results for %q do not contain %q", tt.term, tt.wantContain)
			}
		})
	}
}

// TestSearchEscapesLikeMetacharacters guards a real injection-adjacent bug:
// unescaped "%" makes the pattern match everything, so a search for "50%"
// would return the entire catalogue instead of nothing.
func TestSearchEscapesLikeMetacharacters(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))

	for _, term := range []string{"%", "%%", "_", "pi_a"} {
		t.Run(term, func(t *testing.T) {
			items, err := repo.SearchByName(t.Context(), term, 50, 0)
			if err != nil {
				t.Fatalf("SearchByName(%q) error = %v", term, err)
			}
			if len(items) > 0 {
				t.Errorf("term %q matched %d rows; metacharacters must be escaped, not interpreted",
					term, len(items))
			}
		})
	}
}

func TestCountByNameAgreesWithSearch(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))
	ctx := t.Context()

	// The count and the page must apply identical normalisation, or the UI
	// reports a total that its own pages can never reach.
	const term = "char"

	total, err := repo.CountByName(ctx, term)
	if err != nil {
		t.Fatalf("CountByName() error = %v", err)
	}
	items, err := repo.SearchByName(ctx, term, 100, 0)
	if err != nil {
		t.Fatalf("SearchByName() error = %v", err)
	}

	if int(total) != len(items) {
		t.Errorf("CountByName = %d but SearchByName returned %d", total, len(items))
	}
}

func TestGetSpeciesByPokemonID(t *testing.T) {
	repo := NewPokemonRepository(testPool(t))

	got, err := repo.GetSpeciesByPokemonID(t.Context(), 25)
	if err != nil {
		t.Fatalf("GetSpeciesByPokemonID(25) error = %v", err)
	}
	if got.Name != "pikachu" {
		t.Errorf("Name = %q, want pikachu", got.Name)
	}

	_, err = repo.GetSpeciesByPokemonID(t.Context(), 999999)
	if !errors.Is(err, model.ErrSpeciesNotFound) {
		t.Errorf("error for a missing pokemon = %v, want ErrSpeciesNotFound", err)
	}
}

func TestTypeMatchups(t *testing.T) {
	repo := NewTypeRepository(testPool(t))
	ctx := t.Context()

	types, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(types) != 18 {
		t.Errorf("got %d types, want 18", len(types))
	}

	m, err := repo.GetMatchups(ctx, "fire")
	if err != nil {
		t.Fatalf("GetMatchups(fire) error = %v", err)
	}
	assertContains(t, "fire weaknesses", m.Weaknesses, []string{"ground", "rock", "water"})
	if len(m.Immunities) != 0 {
		t.Errorf("fire immunities = %v, want none", m.Immunities)
	}

	// Ghost is one of the eight immunity cases in the chart.
	ghost, err := repo.GetMatchups(ctx, "ghost")
	if err != nil {
		t.Fatalf("GetMatchups(ghost) error = %v", err)
	}
	assertContains(t, "ghost immunities", ghost.Immunities, []string{"normal", "fighting"})

	// An unknown type must be a 404, not an empty matchup set that looks like
	// a type with no interactions.
	if _, err := repo.GetMatchups(ctx, "fier"); !errors.Is(err, model.ErrTypeNotFound) {
		t.Errorf("error for an unknown type = %v, want ErrTypeNotFound", err)
	}
}

func assertContains(t *testing.T, label string, got, want []string) {
	t.Helper()
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s = %v, missing %q", label, got, w)
		}
	}
}

// TestUpsertIsIdempotent proves the seeder can be re-run: a second write of the
// same row updates instead of failing on the primary key.
func TestUpsertIsIdempotent(t *testing.T) {
	pool := testPool(t)
	ctx := t.Context()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	// Rolled back unconditionally: this test must not mutate the shared dataset.
	defer func() { _ = tx.Rollback(ctx) }()

	q := New(tx)
	params := UpsertSpeciesParams{
		ID: 999001, Name: "test-species", Generation: 1,
		FlavorText: "first", IsLegendary: false, IsMythical: false,
	}

	if err := q.UpsertSpecies(ctx, params); err != nil {
		t.Fatalf("first upsert failed: %v", err)
	}
	params.FlavorText = "second"
	if err := q.UpsertSpecies(ctx, params); err != nil {
		t.Fatalf("second upsert failed; ON CONFLICT DO UPDATE is not in effect: %v", err)
	}

	var flavor string
	if err := tx.QueryRow(ctx, "SELECT flavor_text FROM species WHERE id = $1", 999001).Scan(&flavor); err != nil {
		t.Fatalf("reading back: %v", err)
	}
	if flavor != "second" {
		t.Errorf("flavor_text = %q, want %q; the upsert must overwrite", flavor, "second")
	}
}

// TestCheckConstraintsRejectInvalidData proves the database is the last line of
// defence: service validation protects against bad API input, constraints
// protect against a buggy seeder or a manual UPDATE.
func TestCheckConstraintsRejectInvalidData(t *testing.T) {
	pool := testPool(t)
	ctx := t.Context()

	tests := []struct {
		name string
		sql  string
		args []any
	}{
		{
			name: "generation out of range",
			sql:  "INSERT INTO species (id, name, generation) VALUES (999002, 'bad-gen', $1)",
			args: []any{99},
		},
		{
			name: "slot outside 1 or 2",
			sql:  "INSERT INTO pokemon_type (pokemon_id, type_id, slot) VALUES (1, 1, $1)",
			args: []any{5},
		},
		{
			name: "effectiveness multiplier not on the allowed scale",
			sql:  "INSERT INTO type_effectiveness (attacking_type_id, defending_type_id, multiplier) VALUES (1, 2, $1)",
			args: []any{3.7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatalf("Begin() error = %v", err)
			}
			defer func() { _ = tx.Rollback(ctx) }()

			if _, err := tx.Exec(ctx, tt.sql, tt.args...); err == nil {
				t.Error("insert succeeded; the CHECK constraint is missing or wrong")
			}
		})
	}
}
