// Package postgres implements the repository interfaces declared in
// internal/service, backed by PostgreSQL through pgx and sqlc-generated
// queries.
//
// Its responsibilities stop at the storage boundary: it maps rows onto domain
// models and translates driver errors (pgx.ErrNoRows) into domain sentinels, so
// no layer above ever imports the database driver.
package postgres
