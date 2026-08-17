package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Query plan regression tests.
//
// Indexes fail silently. Dropping one in a migration, or rewriting
// "WHERE lower(name) LIKE $1" as "WHERE name ILIKE $1", leaves every functional
// test green while the query quietly degrades to a full scan. These are the
// only tests in the suite that catch that class of regression.

// explain returns the textual plan for a query.
func explain(t *testing.T, pool *pgxpool.Pool, query string, args ...any) string {
	t.Helper()

	rows, err := pool.Query(t.Context(), "EXPLAIN (ANALYZE, BUFFERS) "+query, args...)
	if err != nil {
		t.Fatalf("EXPLAIN failed: %v", err)
	}
	defer rows.Close()

	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scanning plan: %v", err)
		}
		plan.WriteString(line)
		plan.WriteString("\n")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading plan: %v", err)
	}
	return plan.String()
}

func TestPrimaryKeyLookupUsesIndex(t *testing.T) {
	pool := testPool(t)

	plan := explain(t, pool, "SELECT * FROM pokemon WHERE id = $1", 25)

	if !strings.Contains(plan, "Index Scan using pokemon_pkey") {
		t.Errorf("detail lookup is not using the primary key index:\n%s", plan)
	}
}

// TestOrderedListAvoidsSort is the subtle one. The primary key index is already
// ordered, so Postgres can walk it and stop after LIMIT rows. A Sort node in
// this plan means it read and sorted the entire table to return twenty rows.
func TestOrderedListAvoidsSort(t *testing.T) {
	pool := testPool(t)

	plan := explain(t, pool, "SELECT id, name FROM pokemon ORDER BY id LIMIT 20")

	if strings.Contains(plan, "Sort") {
		t.Errorf("ordered listing introduced a Sort node; the ordered index should make it unnecessary:\n%s", plan)
	}
	if !strings.Contains(plan, "Index Scan using pokemon_pkey") {
		t.Errorf("ordered listing is not walking the primary key index:\n%s", plan)
	}
}

func TestTypeFilterUsesForeignKeyIndex(t *testing.T) {
	pool := testPool(t)

	plan := explain(t, pool, `
		SELECT p.id FROM pokemon p
		JOIN pokemon_type pt ON pt.pokemon_id = p.id
		JOIN type t ON t.id = pt.type_id
		WHERE t.name = $1 LIMIT 20`, "fire")

	// PostgreSQL creates indexes for PRIMARY KEY and UNIQUE but never for a
	// FOREIGN KEY, so this one is declared by hand; without it the join scans
	// all of pokemon_type.
	if strings.Contains(plan, "Seq Scan on pokemon_type") {
		t.Errorf("type filter is scanning pokemon_type; pokemon_type_type_id_idx is missing or unused:\n%s", plan)
	}
}

// TestTrigramIndexIsUsableForSubstringSearch documents a measured result that
// contradicts the obvious expectation.
//
// The GIN trigram index is correct and usable, but at this table size (~1300
// rows, 728 kB) the planner prefers a sequential scan: the whole table is in
// shared_buffers, and scanning it costs less than walking the GIN index. That
// is the planner doing its job, not a broken index.
//
// So this test asserts what actually matters and stays true as the table grows:
// the index exists, it is usable, and it returns the same rows the scan does.
// Asserting "no Seq Scan" here would fail today and would have to be deleted,
// which is worse than asserting something honest.
func TestTrigramIndexIsUsableForSubstringSearch(t *testing.T) {
	pool := testPool(t)
	ctx := t.Context()

	const query = "SELECT id FROM pokemon WHERE lower(name) LIKE $1 ORDER BY id"
	const pattern = "%char%"

	// A dedicated connection: enable_seqscan is a session setting, and applying
	// it to a pooled connection would leak into unrelated queries.
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquiring connection: %v", err)
	}
	defer conn.Release()

	if _, err = conn.Exec(ctx, "SET enable_seqscan = off"); err != nil {
		t.Fatalf("disabling seqscan: %v", err)
	}

	var forced strings.Builder
	rows, err := conn.Query(ctx, "EXPLAIN "+query, pattern)
	if err != nil {
		t.Fatalf("EXPLAIN failed: %v", err)
	}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scanning plan: %v", err)
		}
		forced.WriteString(line + "\n")
	}
	rows.Close()

	if !strings.Contains(forced.String(), "pokemon_name_trgm_idx") {
		t.Fatalf("the trigram index cannot serve a substring search at all; "+
			"check that the predicate spells lower(name) exactly as the index does:\n%s", forced.String())
	}

	// The index must also be correct, not merely present: forcing it must
	// produce the same rows as the plan the planner chooses on its own.
	indexed := collectIDs(ctx, t, conn, query, pattern)

	if _, err := conn.Exec(ctx, "SET enable_seqscan = on"); err != nil {
		t.Fatalf("restoring seqscan: %v", err)
	}
	scanned := collectIDs(ctx, t, conn, query, pattern)

	if len(indexed) == 0 {
		t.Fatal("substring search returned nothing; the dataset looks wrong")
	}
	if len(indexed) != len(scanned) {
		t.Fatalf("index returned %d rows, sequential scan returned %d; they must agree",
			len(indexed), len(scanned))
	}
	for i := range indexed {
		if indexed[i] != scanned[i] {
			t.Fatalf("row %d differs: index %d, scan %d", i, indexed[i], scanned[i])
		}
	}
}

func collectIDs(ctx context.Context, t *testing.T, conn *pgxpool.Conn, query, pattern string) []int32 {
	t.Helper()

	rows, err := conn.Query(ctx, query, pattern)
	if err != nil {
		t.Fatalf("querying: %v", err)
	}
	defer rows.Close()

	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scanning: %v", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading rows: %v", err)
	}
	return ids
}
