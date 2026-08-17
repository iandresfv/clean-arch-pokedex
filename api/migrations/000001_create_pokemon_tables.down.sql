-- Reverse dependency order: pokemon references species.
DROP TRIGGER IF EXISTS pokemon_set_updated_at ON pokemon;
DROP TRIGGER IF EXISTS species_set_updated_at ON species;

DROP TABLE IF EXISTS pokemon;
DROP TABLE IF EXISTS species;

DROP FUNCTION IF EXISTS set_updated_at();
