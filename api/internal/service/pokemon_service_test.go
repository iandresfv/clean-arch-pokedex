package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// stubRepo is a hand-written fake.
//
// Written rather than generated: the interface has eight methods, and a
// generator would be more machinery than the thing it produces. Each field is a
// function so a test can override exactly the behaviour it cares about and
// leave the rest nil.
type stubRepo struct {
	listFn        func(ctx context.Context, limit, offset int32) ([]model.PokemonListItem, error)
	searchFn      func(ctx context.Context, term string, limit, offset int32) ([]model.PokemonListItem, error)
	listByTypeFn  func(ctx context.Context, typeName string, limit, offset int32) ([]model.PokemonListItem, error)
	getByIDFn     func(ctx context.Context, id int32) (model.Pokemon, error)
	countFn       func(ctx context.Context) (int64, error)
	countByNameFn func(ctx context.Context, term string) (int64, error)
	countByTypeFn func(ctx context.Context, typeName string) (int64, error)
	getSpeciesFn  func(ctx context.Context, id int32) (model.Species, error)

	// Recorded so tests can assert what the service passed downstream.
	gotTerm   string
	gotLimit  int32
	gotOffset int32
	listCalls int
}

func (s *stubRepo) List(ctx context.Context, limit, offset int32) ([]model.PokemonListItem, error) {
	s.listCalls++
	s.gotLimit, s.gotOffset = limit, offset
	if s.listFn != nil {
		return s.listFn(ctx, limit, offset)
	}
	return []model.PokemonListItem{}, nil
}

func (s *stubRepo) SearchByName(ctx context.Context, term string, limit, offset int32) ([]model.PokemonListItem, error) {
	s.gotTerm, s.gotLimit, s.gotOffset = term, limit, offset
	if s.searchFn != nil {
		return s.searchFn(ctx, term, limit, offset)
	}
	return []model.PokemonListItem{}, nil
}

func (s *stubRepo) ListByType(ctx context.Context, typeName string, limit, offset int32) ([]model.PokemonListItem, error) {
	s.gotTerm = typeName
	if s.listByTypeFn != nil {
		return s.listByTypeFn(ctx, typeName, limit, offset)
	}
	return []model.PokemonListItem{}, nil
}

func (s *stubRepo) GetByID(ctx context.Context, id int32) (model.Pokemon, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}
	return model.Pokemon{ID: id}, nil
}

func (s *stubRepo) Count(ctx context.Context) (int64, error) {
	if s.countFn != nil {
		return s.countFn(ctx)
	}
	return 0, nil
}

func (s *stubRepo) CountByName(ctx context.Context, term string) (int64, error) {
	if s.countByNameFn != nil {
		return s.countByNameFn(ctx, term)
	}
	return 0, nil
}

func (s *stubRepo) CountByType(ctx context.Context, typeName string) (int64, error) {
	if s.countByTypeFn != nil {
		return s.countByTypeFn(ctx, typeName)
	}
	return 0, nil
}

func (s *stubRepo) GetSpeciesByPokemonID(ctx context.Context, id int32) (model.Species, error) {
	if s.getSpeciesFn != nil {
		return s.getSpeciesFn(ctx, id)
	}
	return model.Species{ID: id}, nil
}

func newService(repo *stubRepo) *PokemonService {
	return NewPokemonService(repo, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func mustPagination(t *testing.T, page, limit int) model.Pagination {
	t.Helper()
	p, err := model.NewPagination(page, limit)
	if err != nil {
		t.Fatalf("building pagination: %v", err)
	}
	return p
}

func TestListPagination(t *testing.T) {
	tests := []struct {
		name            string
		total           int64
		page            int
		limit           int
		wantTotalPages  int
		wantNext        bool
		wantPrev        bool
		wantRepoQueried bool
		wantOffset      int32
	}{
		{
			name: "first of many", total: 1302, page: 1, limit: 20,
			wantTotalPages: 66, wantNext: true, wantPrev: false,
			wantRepoQueried: true, wantOffset: 0,
		},
		{
			name: "middle page", total: 1302, page: 33, limit: 20,
			wantTotalPages: 66, wantNext: true, wantPrev: true,
			wantRepoQueried: true, wantOffset: 640,
		},
		{
			name: "last page", total: 1302, page: 66, limit: 20,
			wantTotalPages: 66, wantNext: false, wantPrev: true,
			wantRepoQueried: true, wantOffset: 1300,
		},
		{
			// Past the end the repository would return nothing anyway, so the
			// query is skipped entirely.
			name: "beyond the last page", total: 1302, page: 100, limit: 20,
			wantTotalPages: 66, wantNext: false, wantPrev: true,
			wantRepoQueried: false,
		},
		{
			name: "empty catalogue", total: 0, page: 1, limit: 20,
			wantTotalPages: 0, wantNext: false, wantPrev: false,
			wantRepoQueried: false,
		},
		{
			// The last page is partial: 25 items at 20 per page is two pages.
			name: "partial last page", total: 25, page: 2, limit: 20,
			wantTotalPages: 2, wantNext: false, wantPrev: true,
			wantRepoQueried: true, wantOffset: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepo{
				countFn: func(context.Context) (int64, error) { return tt.total, nil },
				listFn: func(_ context.Context, limit, _ int32) ([]model.PokemonListItem, error) {
					items := make([]model.PokemonListItem, 0, limit)
					for i := range limit {
						items = append(items, model.PokemonListItem{ID: i + 1})
					}
					return items, nil
				},
			}

			got, err := newService(repo).List(t.Context(), mustPagination(t, tt.page, tt.limit))
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}

			if got.Total != tt.total {
				t.Errorf("Total = %d, want %d", got.Total, tt.total)
			}
			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", got.TotalPages, tt.wantTotalPages)
			}
			if got.HasNextPage != tt.wantNext {
				t.Errorf("HasNextPage = %v, want %v", got.HasNextPage, tt.wantNext)
			}
			if got.HasPreviousPage != tt.wantPrev {
				t.Errorf("HasPreviousPage = %v, want %v", got.HasPreviousPage, tt.wantPrev)
			}
			if queried := repo.listCalls > 0; queried != tt.wantRepoQueried {
				t.Errorf("repository queried = %v, want %v", queried, tt.wantRepoQueried)
			}
			if tt.wantRepoQueried && repo.gotOffset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", repo.gotOffset, tt.wantOffset)
			}
			// Data must never be nil: a nil slice marshals to JSON null and
			// forces every client to null-check each page.
			if got.Data == nil {
				t.Error("Data is nil; it must be an empty slice")
			}
		})
	}
}

func TestSearchByName(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		wantErr  error
		wantTerm string
	}{
		{name: "normal term", term: "pika", wantTerm: "pika"},
		{name: "trimmed and lowercased", term: "  PIKA  ", wantTerm: "pika"},
		{name: "at the minimum length", term: "pi", wantTerm: "pi"},
		{name: "below the minimum", term: "p", wantErr: model.ErrInvalidQuery},
		{name: "empty", term: "", wantErr: model.ErrInvalidQuery},
		{name: "whitespace only", term: "   ", wantErr: model.ErrInvalidQuery},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepo{
				countByNameFn: func(context.Context, string) (int64, error) { return 5, nil },
				searchFn: func(context.Context, string, int32, int32) ([]model.PokemonListItem, error) {
					return []model.PokemonListItem{{ID: 25, Name: "pikachu"}}, nil
				},
			}

			_, err := newService(repo).SearchByName(t.Context(), tt.term, mustPagination(t, 1, 20))

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("SearchByName() error = %v", err)
			}
			// The term must reach the repository normalised, or the count and
			// the page would disagree about what was searched.
			if repo.gotTerm != tt.wantTerm {
				t.Errorf("term passed to repository = %q, want %q", repo.gotTerm, tt.wantTerm)
			}
		})
	}
}

func TestSearchWithNoMatchesSkipsQuery(t *testing.T) {
	queried := false
	repo := &stubRepo{
		countByNameFn: func(context.Context, string) (int64, error) { return 0, nil },
		searchFn: func(context.Context, string, int32, int32) ([]model.PokemonListItem, error) {
			queried = true
			return nil, nil
		},
	}

	got, err := newService(repo).SearchByName(t.Context(), "zzz", mustPagination(t, 1, 20))
	if err != nil {
		t.Fatalf("SearchByName() error = %v", err)
	}
	if queried {
		t.Error("repository was queried for a search with zero matches")
	}
	if len(got.Data) != 0 || got.Total != 0 {
		t.Errorf("got %d items and total %d, want an empty result", len(got.Data), got.Total)
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int32
		repoErr error
		wantErr error
	}{
		{name: "found", id: 25},
		{name: "zero id", id: 0, wantErr: model.ErrInvalidID},
		{name: "negative id", id: -1, wantErr: model.ErrInvalidID},
		{
			name: "not found propagates", id: 9999,
			repoErr: model.ErrPokemonNotFound, wantErr: model.ErrPokemonNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubRepo{
				getByIDFn: func(_ context.Context, id int32) (model.Pokemon, error) {
					if tt.repoErr != nil {
						return model.Pokemon{}, tt.repoErr
					}
					return model.Pokemon{ID: id, Name: "pikachu"}, nil
				},
			}

			got, err := newService(repo).GetByID(t.Context(), tt.id)

			if tt.wantErr != nil {
				// errors.Is, not ==: the service wraps with %w and a direct
				// comparison would fail on the wrapper.
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("GetByID() error = %v", err)
			}
			if got.ID != tt.id {
				t.Errorf("ID = %d, want %d", got.ID, tt.id)
			}
		})
	}
}

func TestGetSpeciesValidatesID(t *testing.T) {
	repo := &stubRepo{}
	_, err := newService(repo).GetSpecies(t.Context(), 0)
	if !errors.Is(err, model.ErrInvalidID) {
		t.Errorf("error = %v, want %v", err, model.ErrInvalidID)
	}
}

// TestRepositoryErrorsAreWrappedNotSwallowed guards against a service that
// turns an infrastructure failure into an empty page, which would render as
// "no results" instead of an error.
func TestRepositoryErrorsAreWrappedNotSwallowed(t *testing.T) {
	boom := errors.New("connection refused")

	repo := &stubRepo{
		countFn: func(context.Context) (int64, error) { return 0, boom },
	}

	_, err := newService(repo).List(t.Context(), mustPagination(t, 1, 20))
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want it to wrap %v", err, boom)
	}
	if !strings.Contains(err.Error(), "listing pokemon") {
		t.Errorf("error %q lost its context; wrap with a description of the operation", err)
	}
}
