-- Core catalogue schema: species and their concrete Pokemon forms.

-- Refreshes updated_at on every UPDATE. Defined once and attached per table so
-- the timestamp cannot drift when a row is modified outside the application.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- In PokeAPI's model a species is the conceptual creature and a Pokemon is one
-- concrete form of it (Charizard has one species and three forms). The
-- relationship is therefore one species to many Pokemon, which is why the
-- foreign key lives on pokemon rather than here.
CREATE TABLE species (
    id           integer     PRIMARY KEY,
    name         text        NOT NULL UNIQUE,
    generation   smallint    NOT NULL CHECK (generation BETWEEN 1 AND 9),
    flavor_text  text        NOT NULL DEFAULT '',
    habitat      text,
    color        text,
    shape        text,
    is_legendary boolean     NOT NULL DEFAULT false,
    is_mythical  boolean     NOT NULL DEFAULT false,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE pokemon (
    id                      integer     PRIMARY KEY,
    name                    text        NOT NULL UNIQUE,
    pokedex_order           integer     NOT NULL,
    species_id              integer     NOT NULL REFERENCES species(id) ON DELETE CASCADE,
    base_experience         integer     CHECK (base_experience IS NULL OR base_experience >= 0),

    -- Units are part of the column name: PokeAPI reports height in decimetres
    -- and weight in hectograms, and formatting stays in the client.
    height_dm               integer     NOT NULL CHECK (height_dm >= 0),
    weight_hg               integer     NOT NULL CHECK (weight_hg >= 0),

    -- The domain defines exactly six fixed stats, so they are six columns
    -- rather than rows in a key/value table: one row read, no join, no pivot.
    stat_hp                 smallint    NOT NULL CHECK (stat_hp              BETWEEN 1 AND 255),
    stat_attack             smallint    NOT NULL CHECK (stat_attack          BETWEEN 1 AND 255),
    stat_defense            smallint    NOT NULL CHECK (stat_defense         BETWEEN 1 AND 255),
    stat_special_attack     smallint    NOT NULL CHECK (stat_special_attack  BETWEEN 1 AND 255),
    stat_special_defense    smallint    NOT NULL CHECK (stat_special_defense BETWEEN 1 AND 255),
    stat_speed              smallint    NOT NULL CHECK (stat_speed           BETWEEN 1 AND 255),

    sprite_front_default    text,
    sprite_front_shiny      text,
    sprite_back_default     text,
    sprite_back_shiny       text,
    sprite_official_artwork text,

    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now()
);

CREATE TRIGGER species_set_updated_at
    BEFORE UPDATE ON species
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER pokemon_set_updated_at
    BEFORE UPDATE ON pokemon
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
