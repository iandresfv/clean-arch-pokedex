DROP TABLE IF EXISTS dataset_version;

DROP INDEX IF EXISTS pokemon_pokedex_order_idx;
DROP INDEX IF EXISTS pokemon_species_id_idx;
DROP INDEX IF EXISTS pokemon_type_type_id_idx;
DROP INDEX IF EXISTS pokemon_name_trgm_idx;

-- The extension is left in place: other schemas in the same database may
-- depend on it, and dropping it is not this migration's decision to make.
