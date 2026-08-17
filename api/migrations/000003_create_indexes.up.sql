-- Search and foreign-key indexes.
--
-- Kept in their own migration because indexes are a performance decision with
-- their own rationale and their own rollback, which makes them visible as such
-- in the schema history.

-- pg_trgm provides trigram matching: it decomposes text into three-character
-- pieces and indexes those.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Substring search cannot use a B-tree. A B-tree is ordered, so finding a value
-- requires knowing how the string starts: LIKE 'pika%' can seek, LIKE '%kach%'
-- cannot and degrades to a sequential scan.
--
-- A GIN (Generalized Inverted Index) over trigrams inverts the problem:
--   pikachu -> "  p", " pi", "pik", "ika", "kac", "ach", "chu", "hu "
-- so '%kach%' becomes a lookup for rows containing both "kac" and "ach".
-- Postgres re-checks the surviving candidates against the real pattern, since
-- the index can yield false positives but never false negatives.
--
-- The index is on lower(name), which makes it a functional index: the query
-- must apply the identical expression or the planner will not use it.
-- Known limitation: patterns shorter than three characters contain no complete
-- trigram, so they fall back to a sequential scan. The service enforces a
-- two-character minimum and at this table size the fallback is sub-millisecond.
CREATE INDEX pokemon_name_trgm_idx
    ON pokemon USING gin (lower(name) gin_trgm_ops);

-- PostgreSQL creates indexes automatically for PRIMARY KEY and UNIQUE, but
-- never for a FOREIGN KEY. Without these, filtering by type scans all of
-- pokemon_type, and cascading deletes scan the child tables.
--
-- Note that pokemon_type's primary key (pokemon_id, slot) already serves
-- lookups by pokemon_id: a B-tree on (a, b) answers queries on a alone, but not
-- on b alone. That leading-column rule is why only type_id needs its own index.
CREATE INDEX pokemon_type_type_id_idx ON pokemon_type (type_id);
CREATE INDEX pokemon_species_id_idx   ON pokemon (species_id);

-- The list endpoint orders by pokedex_order; without this the planner adds a
-- Sort node instead of walking an ordered index.
CREATE INDEX pokemon_pokedex_order_idx ON pokemon (pokedex_order);

-- Tracks which seeder run produced the current dataset. Cache keys are
-- namespaced by this value, which turns invalidation into a version bump
-- instead of a key scan.
CREATE TABLE dataset_version (
    id         smallint    PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    version    bigint      NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO dataset_version (id, version) VALUES (1, 1);
