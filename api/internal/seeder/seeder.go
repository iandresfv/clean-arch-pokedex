package seeder

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/repository/postgres"
)

// Options configures one seeding run.
type Options struct {
	Limit       int
	Concurrency int
	DryRun      bool
}

// Seeder loads the catalogue from PokeAPI into PostgreSQL.
type Seeder struct {
	client *Client
	pool   *pgxpool.Pool
	log    *slog.Logger
}

// New builds a Seeder.
func New(client *Client, pool *pgxpool.Pool, log *slog.Logger) *Seeder {
	return &Seeder{client: client, pool: pool, log: log}
}

// record is one fully fetched Pokemon with its species.
type record struct {
	pokemon *PokemonResponse
	species *SpeciesResponse
}

// Result summarises a run.
type Result struct {
	Fetched  int
	Inserted int
	Skipped  int
	Duration time.Duration
	Version  int64
}

// Run fetches and stores the catalogue.
func (s *Seeder) Run(ctx context.Context, opts Options) (Result, error) {
	start := time.Now()

	if opts.Concurrency < 1 {
		opts.Concurrency = 8
	}

	entries, err := s.client.ListPokemon(ctx, opts.Limit)
	if err != nil {
		return Result{}, err
	}
	s.log.Info("index fetched", "count", len(entries), "concurrency", opts.Concurrency)

	records, skipped := s.fetchAll(ctx, entries, opts.Concurrency)
	if ctx.Err() != nil {
		return Result{}, ctx.Err()
	}

	result := Result{Fetched: len(records), Skipped: skipped, Duration: time.Since(start)}

	if opts.DryRun {
		s.log.Info("dry run: nothing written", "would_insert", len(records))
		result.Duration = time.Since(start)
		return result, nil
	}

	version, err := s.persist(ctx, records)
	if err != nil {
		return Result{}, err
	}

	result.Inserted = len(records)
	result.Version = version
	result.Duration = time.Since(start)
	return result, nil
}

// fetchAll retrieves every entry through a bounded worker pool.
//
// Concurrency is capped deliberately. Spawning one goroutine per Pokemon would
// open ~1300 simultaneous connections to PokeAPI, which gets the client rate
// limited or banned — and would be indistinguishable from an attack.
func (s *Seeder) fetchAll(ctx context.Context, entries []ListEntry, workers int) ([]record, int) {
	jobs := make(chan ListEntry)
	results := make(chan *record)

	var wg sync.WaitGroup
	for range workers {
		// WaitGroup.Go (Go 1.25) pairs Add and Done, removing the classic bug
		// where an early return skips the deferred Done.
		wg.Go(func() {
			for entry := range jobs {
				rec, err := s.fetchOne(ctx, entry)
				if err != nil {
					// A single missing or malformed resource must not abort a
					// run of a thousand; it is counted and reported instead.
					s.log.Warn("skipping pokemon", "name", entry.Name, "error", err)
					results <- nil
					continue
				}
				results <- rec
			}
		})
	}

	go func() {
		defer close(jobs)
		for _, e := range entries {
			select {
			case <-ctx.Done():
				return
			case jobs <- e:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	records := make([]record, 0, len(entries))
	skipped := 0
	done := 0
	for rec := range results {
		done++
		if rec == nil {
			skipped++
		} else {
			records = append(records, *rec)
		}
		if done%100 == 0 {
			s.log.Info("fetch progress", "done", done, "total", len(entries))
		}
	}
	return records, skipped
}

func (s *Seeder) fetchOne(ctx context.Context, entry ListEntry) (*record, error) {
	p, err := s.client.GetPokemon(ctx, entry.URL)
	if err != nil {
		return nil, err
	}

	sp, err := s.client.GetSpecies(ctx, p.Species.URL)
	if err != nil {
		return nil, err
	}

	return &record{pokemon: p, species: sp}, nil
}

// persist writes every record inside one transaction.
//
// All-or-nothing matters here: a failure halfway through a partial write would
// leave the catalogue in a state no page of the UI reflects, and the next run
// would have no way to tell which rows were already correct.
//
// pgx.Batch pipelines the statements into few round trips. CopyFrom would be
// faster still, but it cannot express ON CONFLICT, and idempotency is worth more
// than the difference at this row count: the seeder must be re-runnable.
func (s *Seeder) persist(ctx context.Context, records []record) (int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	// Rollback after a successful Commit is a no-op, so this is safe as an
	// unconditional cleanup path.
	defer func() { _ = tx.Rollback(ctx) }()

	q := postgres.New(tx)

	// Species first: pokemon carries a foreign key to it.
	if err = s.writeSpecies(ctx, q, records); err != nil {
		return 0, err
	}
	if err = s.writePokemon(ctx, q, records); err != nil {
		return 0, err
	}
	if err = s.writeTypes(ctx, tx, records); err != nil {
		return 0, err
	}

	version, err := q.BumpDatasetVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("bumping dataset version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing seed transaction: %w", err)
	}

	// ANALYZE runs outside the transaction because it is maintenance, not part
	// of the atomic write. Until the planner's statistics are refreshed they
	// still describe an empty table, so it picks sequential scans for
	// everything and a freshly seeded database is slower than an idle one.
	if err := postgres.New(s.pool).AnalyzePokemon(ctx); err != nil {
		s.log.Warn("ANALYZE failed; query plans may be poor until autovacuum runs", "error", err)
	}

	return version, nil
}

func (s *Seeder) writeSpecies(ctx context.Context, q *postgres.Queries, records []record) error {
	seen := make(map[int32]struct{}, len(records))

	for _, rec := range records {
		sp := rec.species
		id := int32(sp.ID)
		// Several forms share one species; inserting it twice in the same
		// transaction is wasted work.
		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}

		if err := q.UpsertSpecies(ctx, postgres.UpsertSpeciesParams{
			ID:          id,
			Name:        sp.Name,
			Generation:  generationNumber(sp.Generation.Name),
			FlavorText:  englishFlavorText(sp),
			Habitat:     namedOrNil(sp.Habitat),
			Color:       namedOrNil(sp.Color),
			Shape:       namedOrNil(sp.Shape),
			IsLegendary: sp.IsLegendary,
			IsMythical:  sp.IsMythical,
		}); err != nil {
			return fmt.Errorf("upserting species %s: %w", sp.Name, err)
		}
	}
	return nil
}

func (s *Seeder) writePokemon(ctx context.Context, q *postgres.Queries, records []record) error {
	for _, rec := range records {
		p := rec.pokemon
		stats := statsByName(p)

		if err := q.UpsertPokemon(ctx, postgres.UpsertPokemonParams{
			ID:                    int32(p.ID),
			Name:                  p.Name,
			PokedexOrder:          int32(p.Order),
			SpeciesID:             int32(rec.species.ID),
			BaseExperience:        int32Ptr(p.BaseExperience),
			HeightDm:              int32(p.Height),
			WeightHg:              int32(p.Weight),
			StatHp:                stats["hp"],
			StatAttack:            stats["attack"],
			StatDefense:           stats["defense"],
			StatSpecialAttack:     stats["special-attack"],
			StatSpecialDefense:    stats["special-defense"],
			StatSpeed:             stats["speed"],
			SpriteFrontDefault:    p.Sprites.FrontDefault,
			SpriteFrontShiny:      p.Sprites.FrontShiny,
			SpriteBackDefault:     p.Sprites.BackDefault,
			SpriteBackShiny:       p.Sprites.BackShiny,
			SpriteOfficialArtwork: p.Sprites.Other.OfficialArtwork.FrontDefault,
		}); err != nil {
			return fmt.Errorf("upserting pokemon %s: %w", p.Name, err)
		}
	}
	return nil
}

// writeTypes replaces each Pokemon's type rows.
//
// Delete-then-insert rather than upsert: a Pokemon that loses a type between
// runs would otherwise keep a stale row, because an upsert can only correct
// slots that still exist.
func (s *Seeder) writeTypes(ctx context.Context, tx pgx.Tx, records []record) error {
	q := postgres.New(tx)

	for _, rec := range records {
		p := rec.pokemon
		if err := q.DeletePokemonTypes(ctx, int32(p.ID)); err != nil {
			return fmt.Errorf("clearing types for %s: %w", p.Name, err)
		}

		for _, t := range p.Types {
			typeID, ok := typeIDByName[t.Type.Name]
			if !ok {
				s.log.Warn("unknown type reported by upstream", "pokemon", p.Name, "type", t.Type.Name)
				continue
			}
			// The schema constrains slot to 1 or 2; anything else is upstream
			// data this schema does not model.
			if t.Slot != 1 && t.Slot != 2 {
				continue
			}
			if err := q.InsertPokemonType(ctx, postgres.InsertPokemonTypeParams{
				PokemonID: int32(p.ID),
				TypeID:    typeID,
				Slot:      int16(t.Slot),
			}); err != nil {
				return fmt.Errorf("inserting type %s for %s: %w", t.Type.Name, p.Name, err)
			}
		}
	}
	return nil
}

// typeIDByName mirrors the ids seeded by migration 000002, which are PokeAPI's
// canonical type ids.
var typeIDByName = map[string]int16{
	"normal": 1, "fighting": 2, "flying": 3, "poison": 4, "ground": 5,
	"rock": 6, "bug": 7, "ghost": 8, "steel": 9, "fire": 10,
	"water": 11, "grass": 12, "electric": 13, "psychic": 14, "ice": 15,
	"dragon": 16, "dark": 17, "fairy": 18,
}

// generationRoman maps PokeAPI's "generation-vii" form onto a number.
var generationRoman = map[string]int16{
	"i": 1, "ii": 2, "iii": 3, "iv": 4, "v": 5,
	"vi": 6, "vii": 7, "viii": 8, "ix": 9,
}

func generationNumber(name string) int16 {
	roman := strings.TrimPrefix(name, "generation-")
	if n, ok := generationRoman[roman]; ok {
		return n
	}
	// The schema constrains generation to 1..9, so an unrecognised value must
	// still land in range rather than fail the whole transaction.
	return 1
}

// englishFlavorText picks the first English entry and normalises the control
// characters PokeAPI embeds for in-game line breaks.
func englishFlavorText(sp *SpeciesResponse) string {
	for _, e := range sp.FlavorTextEntries {
		if e.Language.Name == "en" {
			r := strings.NewReplacer("\n", " ", "\f", " ", "\u00ad", "", "\r", " ")
			return strings.Join(strings.Fields(r.Replace(e.FlavorText)), " ")
		}
	}
	return ""
}

func statsByName(p *PokemonResponse) map[string]int16 {
	out := make(map[string]int16, len(p.Stats))
	for _, s := range p.Stats {
		// The schema constrains stats to 1..255. Clamping keeps one anomalous
		// upstream value from failing the entire transaction.
		out[s.Stat.Name] = int16(min(max(s.BaseStat, 1), 255))
	}
	return out
}

func namedOrNil(n *NamedResource) *string {
	if n == nil || n.Name == "" {
		return nil
	}
	return &n.Name
}

func int32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	c := int32(*v)
	return &c
}
