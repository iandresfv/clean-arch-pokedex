-- List queries return the complete row, not a slim projection.
--
-- The client's PokemonRepository port returns domain entities, and a domain
-- Pokemon cannot be constructed without its stats, measurements and order. A
-- lighter projection would force the client to fetch each row's detail
-- separately, reintroducing exactly the N+1 this API exists to remove. One page
-- costs roughly 30 kB instead of 8 kB, in exchange for one request instead of
-- twenty-one.
--
-- Types are aggregated in SQL with array_agg so that one Pokemon is always one
-- row. A plain join to pokemon_type multiplies rows (a dual-type Pokemon
-- returns two), forcing the caller to collapse duplicates and shipping a second
-- copy of every sprite URL over the wire.
--
-- COALESCE guarantees an empty array rather than NULL, so the Go side never has
-- to distinguish "no types" from "null slice".

-- name: ListPokemon :many
SELECT
    p.id,
    p.name,
    p.pokedex_order,
    p.base_experience,
    p.height_dm,
    p.weight_hg,
    p.stat_hp,
    p.stat_attack,
    p.stat_defense,
    p.stat_special_attack,
    p.stat_special_defense,
    p.stat_speed,
    p.sprite_front_default,
    p.sprite_front_shiny,
    p.sprite_back_default,
    p.sprite_back_shiny,
    p.sprite_official_artwork,
    COALESCE((
        SELECT array_agg(t.name ORDER BY pt.slot)
        FROM pokemon_type pt
        JOIN type t ON t.id = pt.type_id
        WHERE pt.pokemon_id = p.id
    ), '{}')::text[] AS type_names
FROM pokemon p
ORDER BY p.id
LIMIT $1 OFFSET $2;

-- name: SearchPokemonByName :many
-- The predicate must spell lower(p.name) exactly as the functional index
-- (pokemon_name_trgm_idx) does, or the planner will not use it. The caller
-- supplies the pattern already lowercased and wrapped in wildcards.
SELECT
    p.id,
    p.name,
    p.pokedex_order,
    p.base_experience,
    p.height_dm,
    p.weight_hg,
    p.stat_hp,
    p.stat_attack,
    p.stat_defense,
    p.stat_special_attack,
    p.stat_special_defense,
    p.stat_speed,
    p.sprite_front_default,
    p.sprite_front_shiny,
    p.sprite_back_default,
    p.sprite_back_shiny,
    p.sprite_official_artwork,
    COALESCE((
        SELECT array_agg(t.name ORDER BY pt.slot)
        FROM pokemon_type pt
        JOIN type t ON t.id = pt.type_id
        WHERE pt.pokemon_id = p.id
    ), '{}')::text[] AS type_names
FROM pokemon p
WHERE lower(p.name) LIKE $1
ORDER BY p.id
LIMIT $2 OFFSET $3;

-- name: ListPokemonByType :many
SELECT
    p.id,
    p.name,
    p.pokedex_order,
    p.base_experience,
    p.height_dm,
    p.weight_hg,
    p.stat_hp,
    p.stat_attack,
    p.stat_defense,
    p.stat_special_attack,
    p.stat_special_defense,
    p.stat_speed,
    p.sprite_front_default,
    p.sprite_front_shiny,
    p.sprite_back_default,
    p.sprite_back_shiny,
    p.sprite_official_artwork,
    COALESCE((
        SELECT array_agg(t2.name ORDER BY pt2.slot)
        FROM pokemon_type pt2
        JOIN type t2 ON t2.id = pt2.type_id
        WHERE pt2.pokemon_id = p.id
    ), '{}')::text[] AS type_names
FROM pokemon p
JOIN pokemon_type pt ON pt.pokemon_id = p.id
JOIN type t ON t.id = pt.type_id
WHERE t.name = $1
ORDER BY p.id
LIMIT $2 OFFSET $3;

-- name: GetPokemonByID :one
-- Species is joined here rather than fetched separately: the relationship is
-- many-to-one and the detail response embeds it, so a second round trip would
-- buy nothing.
SELECT
    p.id,
    p.name,
    p.pokedex_order,
    p.base_experience,
    p.height_dm,
    p.weight_hg,
    p.stat_hp,
    p.stat_attack,
    p.stat_defense,
    p.stat_special_attack,
    p.stat_special_defense,
    p.stat_speed,
    p.sprite_front_default,
    p.sprite_front_shiny,
    p.sprite_back_default,
    p.sprite_back_shiny,
    p.sprite_official_artwork,
    COALESCE((
        SELECT array_agg(t.name ORDER BY pt.slot)
        FROM pokemon_type pt
        JOIN type t ON t.id = pt.type_id
        WHERE pt.pokemon_id = p.id
    ), '{}')::text[] AS type_names,
    s.id           AS species_id,
    s.name         AS species_name,
    s.generation   AS species_generation,
    s.flavor_text  AS species_flavor_text,
    s.habitat      AS species_habitat,
    s.is_legendary AS species_is_legendary,
    s.is_mythical  AS species_is_mythical
FROM pokemon p
JOIN species s ON s.id = p.species_id
WHERE p.id = $1;

-- name: CountPokemon :one
-- count(*) reads the whole table: index entries do not record row visibility
-- for the current transaction (MVCC), so the planner cannot answer from an
-- index alone. At this table size that is microseconds. For a large table the
-- estimate in pg_class.reltuples would be the scalable alternative.
SELECT count(*) FROM pokemon;

-- name: CountPokemonByName :one
SELECT count(*) FROM pokemon WHERE lower(name) LIKE $1;

-- name: CountPokemonByType :one
SELECT count(*)
FROM pokemon p
JOIN pokemon_type pt ON pt.pokemon_id = p.id
JOIN type t ON t.id = pt.type_id
WHERE t.name = $1;

-- name: UpsertPokemon :exec
-- Idempotent so the seeder can be re-run: a second run updates rather than
-- failing on the primary key.
INSERT INTO pokemon (
    id, name, pokedex_order, species_id, base_experience,
    height_dm, weight_hg,
    stat_hp, stat_attack, stat_defense,
    stat_special_attack, stat_special_defense, stat_speed,
    sprite_front_default, sprite_front_shiny,
    sprite_back_default, sprite_back_shiny, sprite_official_artwork
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7,
    $8, $9, $10,
    $11, $12, $13,
    $14, $15,
    $16, $17, $18
)
ON CONFLICT (id) DO UPDATE SET
    name                    = EXCLUDED.name,
    pokedex_order           = EXCLUDED.pokedex_order,
    species_id              = EXCLUDED.species_id,
    base_experience         = EXCLUDED.base_experience,
    height_dm               = EXCLUDED.height_dm,
    weight_hg               = EXCLUDED.weight_hg,
    stat_hp                 = EXCLUDED.stat_hp,
    stat_attack             = EXCLUDED.stat_attack,
    stat_defense            = EXCLUDED.stat_defense,
    stat_special_attack     = EXCLUDED.stat_special_attack,
    stat_special_defense    = EXCLUDED.stat_special_defense,
    stat_speed              = EXCLUDED.stat_speed,
    sprite_front_default    = EXCLUDED.sprite_front_default,
    sprite_front_shiny      = EXCLUDED.sprite_front_shiny,
    sprite_back_default     = EXCLUDED.sprite_back_default,
    sprite_back_shiny       = EXCLUDED.sprite_back_shiny,
    sprite_official_artwork = EXCLUDED.sprite_official_artwork;

-- name: DeletePokemonTypes :exec
DELETE FROM pokemon_type WHERE pokemon_id = $1;

-- name: InsertPokemonType :exec
INSERT INTO pokemon_type (pokemon_id, type_id, slot)
VALUES ($1, $2, $3)
ON CONFLICT (pokemon_id, slot) DO UPDATE SET type_id = EXCLUDED.type_id;

-- name: AnalyzePokemon :exec
-- Run after a bulk load. Until ANALYZE refreshes the planner's statistics they
-- still describe an empty table, so it picks sequential scans for everything
-- and a freshly seeded database is measurably slower than an idle one.
ANALYZE pokemon, species, pokemon_type;
