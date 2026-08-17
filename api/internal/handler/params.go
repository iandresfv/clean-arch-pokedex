package handler

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// parsePagination reads page and limit, applying defaults for absent values.
//
// An absent parameter is not an error; a present but unparseable one is.
// Falling back to the default on garbage input would make "?page=abc" silently
// return page 1, which is indistinguishable from a working request.
func parsePagination(r *http.Request) (model.Pagination, error) {
	page, err := intParam(r, "page", model.DefaultPage)
	if err != nil {
		return model.Pagination{}, err
	}
	limit, err := intParam(r, "limit", model.DefaultLimit)
	if err != nil {
		return model.Pagination{}, err
	}
	return model.NewPagination(page, limit)
}

func intParam(r *http.Request, name string, fallback int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be an integer, got %q", model.ErrInvalidPagination, name, raw)
	}
	return v, nil
}

// rejectUnknownParams fails a request carrying a parameter the endpoint does
// not define.
//
// Accepting "?pgae=2" and silently serving page 1 turns a typo into a bug that
// surfaces as "pagination doesn't work" with nothing in the logs to explain it.
// Rejecting it converts a silent misunderstanding into an immediate 400.
func rejectUnknownParams(r *http.Request, allowed ...string) error {
	for name := range r.URL.Query() {
		if !slices.Contains(allowed, name) {
			return fmt.Errorf("%w: %q is not a valid parameter for this endpoint (allowed: %v)",
				model.ErrUnknownParameter, name, allowed)
		}
	}
	return nil
}

// pathID reads a numeric path variable.
func pathID(r *http.Request, name string) (int32, error) {
	raw := r.PathValue(name)
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("%w: %q is not a valid id", model.ErrInvalidID, raw)
	}
	return int32(v), nil
}
