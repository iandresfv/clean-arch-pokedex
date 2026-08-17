-- name: ListTypes :many
SELECT id, name FROM type ORDER BY id;

-- name: GetTypeByName :one
SELECT id, name FROM type WHERE name = $1;

-- name: GetTypeMatchups :many
-- Defensive matchups: how much damage the given type takes from each attacking
-- type. Neutral matchups are excluded because the caller only renders
-- weaknesses, resistances and immunities.
SELECT
    ta.name       AS attacking_type,
    te.multiplier AS multiplier
FROM type_effectiveness te
JOIN type td ON td.id = te.defending_type_id
JOIN type ta ON ta.id = te.attacking_type_id
WHERE td.name = $1
  AND te.multiplier <> 1
ORDER BY te.multiplier DESC, ta.name;

-- name: GetDatasetVersion :one
SELECT version FROM dataset_version WHERE id = 1;

-- name: BumpDatasetVersion :one
-- Cache keys are namespaced by this value, so invalidating every cached
-- response after a seeder run is a single increment rather than a key scan.
UPDATE dataset_version
SET version = version + 1, updated_at = now()
WHERE id = 1
RETURNING version;
