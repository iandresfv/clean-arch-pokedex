package model

import "fmt"

// Pagination bounds. The maximum limit is not cosmetic: without it
// "?limit=1000000" is a trivially available denial of service, forcing a full
// table scan and serialisation of every row in a single request.
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
	MinLimit     = 1

	// MaxPage caps how far a client may page. (MaxPage * MaxLimit) stays well
	// inside int32, which is what makes Offset's conversion safe.
	MaxPage = 1_000_000

	// MinSearchLength exists because a trigram index cannot serve patterns
	// shorter than three characters — there is no complete trigram to look up,
	// so the query degrades to a sequential scan. Two is the pragmatic floor at
	// this table size.
	MinSearchLength = 2
)

// Pagination is a validated page request. Constructing one through
// NewPagination is the only way to obtain a value that the repository can trust
// without re-checking bounds.
type Pagination struct {
	Page  int
	Limit int
}

// NewPagination validates and returns a page request.
//
// Validation lives here rather than in the handler so that every entry point —
// HTTP today, a CLI or gRPC surface tomorrow — enforces identical bounds.
func NewPagination(page, limit int) (Pagination, error) {
	if page < 1 {
		return Pagination{}, fmt.Errorf("%w: page must be at least 1, got %d", ErrInvalidPagination, page)
	}
	if page > MaxPage {
		return Pagination{}, fmt.Errorf("%w: page must not exceed %d, got %d", ErrInvalidPagination, MaxPage, page)
	}
	if limit < MinLimit {
		return Pagination{}, fmt.Errorf("%w: limit must be at least %d, got %d", ErrInvalidPagination, MinLimit, limit)
	}
	if limit > MaxLimit {
		return Pagination{}, fmt.Errorf("%w: limit must not exceed %d, got %d", ErrInvalidPagination, MaxLimit, limit)
	}
	return Pagination{Page: page, Limit: limit}, nil
}

// Offset returns the number of rows to skip.
//
// The int32 return type matches the generated query parameters. Overflow is not
// reachable: Limit is capped at MaxLimit and Page is bounded by MaxPage below.
func (p Pagination) Offset() int32 {
	return int32((p.Page - 1) * p.Limit)
}

// LimitInt32 returns the page size in the width the generated queries expect.
func (p Pagination) LimitInt32() int32 {
	return int32(p.Limit)
}
