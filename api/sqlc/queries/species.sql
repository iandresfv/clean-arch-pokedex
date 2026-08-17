-- name: GetSpeciesByPokemonID :one
SELECT
    s.id,
    s.name,
    s.generation,
    s.flavor_text,
    s.habitat,
    s.color,
    s.shape,
    s.is_legendary,
    s.is_mythical
FROM species s
JOIN pokemon p ON p.species_id = s.id
WHERE p.id = $1;

-- name: UpsertSpecies :exec
INSERT INTO species (
    id, name, generation, flavor_text, habitat, color, shape,
    is_legendary, is_mythical
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (id) DO UPDATE SET
    name         = EXCLUDED.name,
    generation   = EXCLUDED.generation,
    flavor_text  = EXCLUDED.flavor_text,
    habitat      = EXCLUDED.habitat,
    color        = EXCLUDED.color,
    shape        = EXCLUDED.shape,
    is_legendary = EXCLUDED.is_legendary,
    is_mythical  = EXCLUDED.is_mythical;
