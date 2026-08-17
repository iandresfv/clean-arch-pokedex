package model

import "errors"

// Sentinel errors shared across layers.
//
// A sentinel is a package-level error value compared with errors.Is. Because
// errors.Is unwraps, a service that returns fmt.Errorf("getting pokemon %d: %w",
// id, err) preserves the original for the handler to identify — the %w verb is
// what performs that wrapping.
//
// They exist so the repository can translate driver-specific failures
// (pgx.ErrNoRows) into something the service understands without importing the
// driver, and so the handler can map failures to status codes without importing
// the repository.
var (
	ErrPokemonNotFound = errors.New("pokemon not found")
	ErrSpeciesNotFound = errors.New("species not found")
	ErrTypeNotFound    = errors.New("type not found")

	// ErrInvalidID covers a non-numeric or out-of-range path parameter.
	ErrInvalidID = errors.New("invalid pokemon id")

	// ErrInvalidPagination covers page or limit outside their accepted bounds.
	ErrInvalidPagination = errors.New("invalid pagination parameters")

	// ErrInvalidQuery covers a search term that cannot be served, such as one
	// shorter than the trigram index can use.
	ErrInvalidQuery = errors.New("invalid search query")

	// ErrUnknownParameter is returned when a request carries a query parameter
	// the endpoint does not define. Silently ignoring "?pgae=2" turns a typo
	// into "pagination is broken" with no error anywhere to explain it.
	ErrUnknownParameter = errors.New("unknown query parameter")
)
