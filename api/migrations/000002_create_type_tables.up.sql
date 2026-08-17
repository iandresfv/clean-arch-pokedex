-- Type reference data and the full effectiveness matrix.
--
-- This is static reference data: it is part of the schema's meaning and changes
-- only when the games change, so it ships in a versioned migration rather than
-- in the network-dependent seeder.

CREATE TABLE type (
    id   smallint PRIMARY KEY,
    name text     NOT NULL UNIQUE
);

-- Ids match PokeAPI's canonical type ids so seeded payloads map directly.
INSERT INTO type (id, name) VALUES
    ( 1, 'normal'),
    ( 2, 'fighting'),
    ( 3, 'flying'),
    ( 4, 'poison'),
    ( 5, 'ground'),
    ( 6, 'rock'),
    ( 7, 'bug'),
    ( 8, 'ghost'),
    ( 9, 'steel'),
    (10, 'fire'),
    (11, 'water'),
    (12, 'grass'),
    (13, 'electric'),
    (14, 'psychic'),
    (15, 'ice'),
    (16, 'dragon'),
    (17, 'dark'),
    (18, 'fairy');

CREATE TABLE pokemon_type (
    pokemon_id integer  NOT NULL REFERENCES pokemon(id) ON DELETE CASCADE,
    type_id    smallint NOT NULL REFERENCES type(id),
    -- Slot order is meaningful: Bulbasaur is grass/poison, not poison/grass.
    slot       smallint NOT NULL CHECK (slot IN (1, 2)),
    PRIMARY KEY (pokemon_id, slot),
    -- Prevents the same type occupying both slots.
    UNIQUE (pokemon_id, type_id)
);

CREATE TABLE type_effectiveness (
    attacking_type_id smallint     NOT NULL REFERENCES type(id),
    defending_type_id smallint     NOT NULL REFERENCES type(id),
    -- numeric, not real: 0.5 has no exact binary representation and a chain of
    -- float multiplications drifts (0.25 -> 0.2499...).
    multiplier        numeric(2,1) NOT NULL CHECK (multiplier IN (0, 0.5, 1, 2)),
    PRIMARY KEY (attacking_type_id, defending_type_id)
);

-- 18 x 18 = 324 rows. Row = attacking type, column = defending type.
INSERT INTO type_effectiveness (attacking_type_id, defending_type_id, multiplier) VALUES
    ( 1,  1, 1.0),  -- normal -> normal
    ( 1,  2, 1.0),  -- normal -> fighting
    ( 1,  3, 1.0),  -- normal -> flying
    ( 1,  4, 1.0),  -- normal -> poison
    ( 1,  5, 1.0),  -- normal -> ground
    ( 1,  6, 0.5),  -- normal -> rock
    ( 1,  7, 1.0),  -- normal -> bug
    ( 1,  8, 0.0),  -- normal -> ghost
    ( 1,  9, 0.5),  -- normal -> steel
    ( 1, 10, 1.0),  -- normal -> fire
    ( 1, 11, 1.0),  -- normal -> water
    ( 1, 12, 1.0),  -- normal -> grass
    ( 1, 13, 1.0),  -- normal -> electric
    ( 1, 14, 1.0),  -- normal -> psychic
    ( 1, 15, 1.0),  -- normal -> ice
    ( 1, 16, 1.0),  -- normal -> dragon
    ( 1, 17, 1.0),  -- normal -> dark
    ( 1, 18, 1.0),  -- normal -> fairy
    ( 2,  1, 2.0),  -- fighting -> normal
    ( 2,  2, 1.0),  -- fighting -> fighting
    ( 2,  3, 0.5),  -- fighting -> flying
    ( 2,  4, 0.5),  -- fighting -> poison
    ( 2,  5, 1.0),  -- fighting -> ground
    ( 2,  6, 2.0),  -- fighting -> rock
    ( 2,  7, 0.5),  -- fighting -> bug
    ( 2,  8, 0.0),  -- fighting -> ghost
    ( 2,  9, 2.0),  -- fighting -> steel
    ( 2, 10, 1.0),  -- fighting -> fire
    ( 2, 11, 1.0),  -- fighting -> water
    ( 2, 12, 1.0),  -- fighting -> grass
    ( 2, 13, 1.0),  -- fighting -> electric
    ( 2, 14, 0.5),  -- fighting -> psychic
    ( 2, 15, 2.0),  -- fighting -> ice
    ( 2, 16, 1.0),  -- fighting -> dragon
    ( 2, 17, 2.0),  -- fighting -> dark
    ( 2, 18, 0.5),  -- fighting -> fairy
    ( 3,  1, 1.0),  -- flying -> normal
    ( 3,  2, 2.0),  -- flying -> fighting
    ( 3,  3, 1.0),  -- flying -> flying
    ( 3,  4, 1.0),  -- flying -> poison
    ( 3,  5, 1.0),  -- flying -> ground
    ( 3,  6, 0.5),  -- flying -> rock
    ( 3,  7, 2.0),  -- flying -> bug
    ( 3,  8, 1.0),  -- flying -> ghost
    ( 3,  9, 0.5),  -- flying -> steel
    ( 3, 10, 1.0),  -- flying -> fire
    ( 3, 11, 1.0),  -- flying -> water
    ( 3, 12, 2.0),  -- flying -> grass
    ( 3, 13, 0.5),  -- flying -> electric
    ( 3, 14, 1.0),  -- flying -> psychic
    ( 3, 15, 1.0),  -- flying -> ice
    ( 3, 16, 1.0),  -- flying -> dragon
    ( 3, 17, 1.0),  -- flying -> dark
    ( 3, 18, 1.0),  -- flying -> fairy
    ( 4,  1, 1.0),  -- poison -> normal
    ( 4,  2, 1.0),  -- poison -> fighting
    ( 4,  3, 1.0),  -- poison -> flying
    ( 4,  4, 0.5),  -- poison -> poison
    ( 4,  5, 0.5),  -- poison -> ground
    ( 4,  6, 0.5),  -- poison -> rock
    ( 4,  7, 1.0),  -- poison -> bug
    ( 4,  8, 0.5),  -- poison -> ghost
    ( 4,  9, 0.0),  -- poison -> steel
    ( 4, 10, 1.0),  -- poison -> fire
    ( 4, 11, 1.0),  -- poison -> water
    ( 4, 12, 2.0),  -- poison -> grass
    ( 4, 13, 1.0),  -- poison -> electric
    ( 4, 14, 1.0),  -- poison -> psychic
    ( 4, 15, 1.0),  -- poison -> ice
    ( 4, 16, 1.0),  -- poison -> dragon
    ( 4, 17, 1.0),  -- poison -> dark
    ( 4, 18, 2.0),  -- poison -> fairy
    ( 5,  1, 1.0),  -- ground -> normal
    ( 5,  2, 1.0),  -- ground -> fighting
    ( 5,  3, 0.0),  -- ground -> flying
    ( 5,  4, 2.0),  -- ground -> poison
    ( 5,  5, 1.0),  -- ground -> ground
    ( 5,  6, 2.0),  -- ground -> rock
    ( 5,  7, 0.5),  -- ground -> bug
    ( 5,  8, 1.0),  -- ground -> ghost
    ( 5,  9, 2.0),  -- ground -> steel
    ( 5, 10, 2.0),  -- ground -> fire
    ( 5, 11, 1.0),  -- ground -> water
    ( 5, 12, 0.5),  -- ground -> grass
    ( 5, 13, 2.0),  -- ground -> electric
    ( 5, 14, 1.0),  -- ground -> psychic
    ( 5, 15, 1.0),  -- ground -> ice
    ( 5, 16, 1.0),  -- ground -> dragon
    ( 5, 17, 1.0),  -- ground -> dark
    ( 5, 18, 1.0),  -- ground -> fairy
    ( 6,  1, 1.0),  -- rock -> normal
    ( 6,  2, 0.5),  -- rock -> fighting
    ( 6,  3, 2.0),  -- rock -> flying
    ( 6,  4, 1.0),  -- rock -> poison
    ( 6,  5, 0.5),  -- rock -> ground
    ( 6,  6, 1.0),  -- rock -> rock
    ( 6,  7, 2.0),  -- rock -> bug
    ( 6,  8, 1.0),  -- rock -> ghost
    ( 6,  9, 0.5),  -- rock -> steel
    ( 6, 10, 2.0),  -- rock -> fire
    ( 6, 11, 1.0),  -- rock -> water
    ( 6, 12, 1.0),  -- rock -> grass
    ( 6, 13, 1.0),  -- rock -> electric
    ( 6, 14, 1.0),  -- rock -> psychic
    ( 6, 15, 2.0),  -- rock -> ice
    ( 6, 16, 1.0),  -- rock -> dragon
    ( 6, 17, 1.0),  -- rock -> dark
    ( 6, 18, 1.0),  -- rock -> fairy
    ( 7,  1, 1.0),  -- bug -> normal
    ( 7,  2, 0.5),  -- bug -> fighting
    ( 7,  3, 0.5),  -- bug -> flying
    ( 7,  4, 0.5),  -- bug -> poison
    ( 7,  5, 1.0),  -- bug -> ground
    ( 7,  6, 1.0),  -- bug -> rock
    ( 7,  7, 1.0),  -- bug -> bug
    ( 7,  8, 0.5),  -- bug -> ghost
    ( 7,  9, 0.5),  -- bug -> steel
    ( 7, 10, 0.5),  -- bug -> fire
    ( 7, 11, 1.0),  -- bug -> water
    ( 7, 12, 2.0),  -- bug -> grass
    ( 7, 13, 1.0),  -- bug -> electric
    ( 7, 14, 2.0),  -- bug -> psychic
    ( 7, 15, 1.0),  -- bug -> ice
    ( 7, 16, 1.0),  -- bug -> dragon
    ( 7, 17, 2.0),  -- bug -> dark
    ( 7, 18, 0.5),  -- bug -> fairy
    ( 8,  1, 0.0),  -- ghost -> normal
    ( 8,  2, 1.0),  -- ghost -> fighting
    ( 8,  3, 1.0),  -- ghost -> flying
    ( 8,  4, 1.0),  -- ghost -> poison
    ( 8,  5, 1.0),  -- ghost -> ground
    ( 8,  6, 1.0),  -- ghost -> rock
    ( 8,  7, 1.0),  -- ghost -> bug
    ( 8,  8, 2.0),  -- ghost -> ghost
    ( 8,  9, 1.0),  -- ghost -> steel
    ( 8, 10, 1.0),  -- ghost -> fire
    ( 8, 11, 1.0),  -- ghost -> water
    ( 8, 12, 1.0),  -- ghost -> grass
    ( 8, 13, 1.0),  -- ghost -> electric
    ( 8, 14, 2.0),  -- ghost -> psychic
    ( 8, 15, 1.0),  -- ghost -> ice
    ( 8, 16, 1.0),  -- ghost -> dragon
    ( 8, 17, 0.5),  -- ghost -> dark
    ( 8, 18, 1.0),  -- ghost -> fairy
    ( 9,  1, 1.0),  -- steel -> normal
    ( 9,  2, 1.0),  -- steel -> fighting
    ( 9,  3, 1.0),  -- steel -> flying
    ( 9,  4, 1.0),  -- steel -> poison
    ( 9,  5, 1.0),  -- steel -> ground
    ( 9,  6, 2.0),  -- steel -> rock
    ( 9,  7, 1.0),  -- steel -> bug
    ( 9,  8, 1.0),  -- steel -> ghost
    ( 9,  9, 0.5),  -- steel -> steel
    ( 9, 10, 0.5),  -- steel -> fire
    ( 9, 11, 0.5),  -- steel -> water
    ( 9, 12, 1.0),  -- steel -> grass
    ( 9, 13, 0.5),  -- steel -> electric
    ( 9, 14, 1.0),  -- steel -> psychic
    ( 9, 15, 2.0),  -- steel -> ice
    ( 9, 16, 1.0),  -- steel -> dragon
    ( 9, 17, 1.0),  -- steel -> dark
    ( 9, 18, 2.0),  -- steel -> fairy
    (10,  1, 1.0),  -- fire -> normal
    (10,  2, 1.0),  -- fire -> fighting
    (10,  3, 1.0),  -- fire -> flying
    (10,  4, 1.0),  -- fire -> poison
    (10,  5, 1.0),  -- fire -> ground
    (10,  6, 0.5),  -- fire -> rock
    (10,  7, 2.0),  -- fire -> bug
    (10,  8, 1.0),  -- fire -> ghost
    (10,  9, 2.0),  -- fire -> steel
    (10, 10, 0.5),  -- fire -> fire
    (10, 11, 0.5),  -- fire -> water
    (10, 12, 2.0),  -- fire -> grass
    (10, 13, 1.0),  -- fire -> electric
    (10, 14, 1.0),  -- fire -> psychic
    (10, 15, 2.0),  -- fire -> ice
    (10, 16, 0.5),  -- fire -> dragon
    (10, 17, 1.0),  -- fire -> dark
    (10, 18, 1.0),  -- fire -> fairy
    (11,  1, 1.0),  -- water -> normal
    (11,  2, 1.0),  -- water -> fighting
    (11,  3, 1.0),  -- water -> flying
    (11,  4, 1.0),  -- water -> poison
    (11,  5, 2.0),  -- water -> ground
    (11,  6, 2.0),  -- water -> rock
    (11,  7, 1.0),  -- water -> bug
    (11,  8, 1.0),  -- water -> ghost
    (11,  9, 1.0),  -- water -> steel
    (11, 10, 2.0),  -- water -> fire
    (11, 11, 0.5),  -- water -> water
    (11, 12, 0.5),  -- water -> grass
    (11, 13, 1.0),  -- water -> electric
    (11, 14, 1.0),  -- water -> psychic
    (11, 15, 1.0),  -- water -> ice
    (11, 16, 0.5),  -- water -> dragon
    (11, 17, 1.0),  -- water -> dark
    (11, 18, 1.0),  -- water -> fairy
    (12,  1, 1.0),  -- grass -> normal
    (12,  2, 1.0),  -- grass -> fighting
    (12,  3, 0.5),  -- grass -> flying
    (12,  4, 0.5),  -- grass -> poison
    (12,  5, 2.0),  -- grass -> ground
    (12,  6, 2.0),  -- grass -> rock
    (12,  7, 0.5),  -- grass -> bug
    (12,  8, 1.0),  -- grass -> ghost
    (12,  9, 0.5),  -- grass -> steel
    (12, 10, 0.5),  -- grass -> fire
    (12, 11, 2.0),  -- grass -> water
    (12, 12, 0.5),  -- grass -> grass
    (12, 13, 1.0),  -- grass -> electric
    (12, 14, 1.0),  -- grass -> psychic
    (12, 15, 1.0),  -- grass -> ice
    (12, 16, 0.5),  -- grass -> dragon
    (12, 17, 1.0),  -- grass -> dark
    (12, 18, 1.0),  -- grass -> fairy
    (13,  1, 1.0),  -- electric -> normal
    (13,  2, 1.0),  -- electric -> fighting
    (13,  3, 2.0),  -- electric -> flying
    (13,  4, 1.0),  -- electric -> poison
    (13,  5, 0.0),  -- electric -> ground
    (13,  6, 1.0),  -- electric -> rock
    (13,  7, 1.0),  -- electric -> bug
    (13,  8, 1.0),  -- electric -> ghost
    (13,  9, 1.0),  -- electric -> steel
    (13, 10, 1.0),  -- electric -> fire
    (13, 11, 2.0),  -- electric -> water
    (13, 12, 0.5),  -- electric -> grass
    (13, 13, 0.5),  -- electric -> electric
    (13, 14, 1.0),  -- electric -> psychic
    (13, 15, 1.0),  -- electric -> ice
    (13, 16, 0.5),  -- electric -> dragon
    (13, 17, 1.0),  -- electric -> dark
    (13, 18, 1.0),  -- electric -> fairy
    (14,  1, 1.0),  -- psychic -> normal
    (14,  2, 2.0),  -- psychic -> fighting
    (14,  3, 1.0),  -- psychic -> flying
    (14,  4, 2.0),  -- psychic -> poison
    (14,  5, 1.0),  -- psychic -> ground
    (14,  6, 1.0),  -- psychic -> rock
    (14,  7, 1.0),  -- psychic -> bug
    (14,  8, 1.0),  -- psychic -> ghost
    (14,  9, 0.5),  -- psychic -> steel
    (14, 10, 1.0),  -- psychic -> fire
    (14, 11, 1.0),  -- psychic -> water
    (14, 12, 1.0),  -- psychic -> grass
    (14, 13, 1.0),  -- psychic -> electric
    (14, 14, 0.5),  -- psychic -> psychic
    (14, 15, 1.0),  -- psychic -> ice
    (14, 16, 1.0),  -- psychic -> dragon
    (14, 17, 0.0),  -- psychic -> dark
    (14, 18, 1.0),  -- psychic -> fairy
    (15,  1, 1.0),  -- ice -> normal
    (15,  2, 1.0),  -- ice -> fighting
    (15,  3, 2.0),  -- ice -> flying
    (15,  4, 1.0),  -- ice -> poison
    (15,  5, 2.0),  -- ice -> ground
    (15,  6, 1.0),  -- ice -> rock
    (15,  7, 1.0),  -- ice -> bug
    (15,  8, 1.0),  -- ice -> ghost
    (15,  9, 0.5),  -- ice -> steel
    (15, 10, 0.5),  -- ice -> fire
    (15, 11, 0.5),  -- ice -> water
    (15, 12, 2.0),  -- ice -> grass
    (15, 13, 1.0),  -- ice -> electric
    (15, 14, 1.0),  -- ice -> psychic
    (15, 15, 0.5),  -- ice -> ice
    (15, 16, 2.0),  -- ice -> dragon
    (15, 17, 1.0),  -- ice -> dark
    (15, 18, 1.0),  -- ice -> fairy
    (16,  1, 1.0),  -- dragon -> normal
    (16,  2, 1.0),  -- dragon -> fighting
    (16,  3, 1.0),  -- dragon -> flying
    (16,  4, 1.0),  -- dragon -> poison
    (16,  5, 1.0),  -- dragon -> ground
    (16,  6, 1.0),  -- dragon -> rock
    (16,  7, 1.0),  -- dragon -> bug
    (16,  8, 1.0),  -- dragon -> ghost
    (16,  9, 0.5),  -- dragon -> steel
    (16, 10, 1.0),  -- dragon -> fire
    (16, 11, 1.0),  -- dragon -> water
    (16, 12, 1.0),  -- dragon -> grass
    (16, 13, 1.0),  -- dragon -> electric
    (16, 14, 1.0),  -- dragon -> psychic
    (16, 15, 1.0),  -- dragon -> ice
    (16, 16, 2.0),  -- dragon -> dragon
    (16, 17, 1.0),  -- dragon -> dark
    (16, 18, 0.0),  -- dragon -> fairy
    (17,  1, 1.0),  -- dark -> normal
    (17,  2, 0.5),  -- dark -> fighting
    (17,  3, 1.0),  -- dark -> flying
    (17,  4, 1.0),  -- dark -> poison
    (17,  5, 1.0),  -- dark -> ground
    (17,  6, 1.0),  -- dark -> rock
    (17,  7, 1.0),  -- dark -> bug
    (17,  8, 2.0),  -- dark -> ghost
    (17,  9, 0.5),  -- dark -> steel
    (17, 10, 1.0),  -- dark -> fire
    (17, 11, 1.0),  -- dark -> water
    (17, 12, 1.0),  -- dark -> grass
    (17, 13, 1.0),  -- dark -> electric
    (17, 14, 2.0),  -- dark -> psychic
    (17, 15, 1.0),  -- dark -> ice
    (17, 16, 1.0),  -- dark -> dragon
    (17, 17, 0.5),  -- dark -> dark
    (17, 18, 0.5),  -- dark -> fairy
    (18,  1, 1.0),  -- fairy -> normal
    (18,  2, 2.0),  -- fairy -> fighting
    (18,  3, 1.0),  -- fairy -> flying
    (18,  4, 0.5),  -- fairy -> poison
    (18,  5, 1.0),  -- fairy -> ground
    (18,  6, 1.0),  -- fairy -> rock
    (18,  7, 1.0),  -- fairy -> bug
    (18,  8, 1.0),  -- fairy -> ghost
    (18,  9, 0.5),  -- fairy -> steel
    (18, 10, 0.5),  -- fairy -> fire
    (18, 11, 1.0),  -- fairy -> water
    (18, 12, 1.0),  -- fairy -> grass
    (18, 13, 1.0),  -- fairy -> electric
    (18, 14, 1.0),  -- fairy -> psychic
    (18, 15, 1.0),  -- fairy -> ice
    (18, 16, 2.0),  -- fairy -> dragon
    (18, 17, 2.0),  -- fairy -> dark
    (18, 18, 1.0);  -- fairy -> fairy
