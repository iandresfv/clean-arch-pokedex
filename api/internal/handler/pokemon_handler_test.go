package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/httperr"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
)

// fakeService lets each test define only the behaviour it exercises.
type fakeService struct {
	listFn       func(ctx context.Context, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error)
	searchFn     func(ctx context.Context, term string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error)
	listByTypeFn func(ctx context.Context, t string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error)
	getByIDFn    func(ctx context.Context, id int32) (model.Pokemon, error)
	getSpeciesFn func(ctx context.Context, id int32) (model.Species, error)

	gotTerm string
	gotPage model.Pagination
}

func (f *fakeService) List(ctx context.Context, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
	f.gotPage = p
	if f.listFn != nil {
		return f.listFn(ctx, p)
	}
	return model.NewPaginatedResult([]model.PokemonListItem{}, 0, p.Page, p.Limit), nil
}

func (f *fakeService) SearchByName(ctx context.Context, term string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
	f.gotTerm, f.gotPage = term, p
	if f.searchFn != nil {
		return f.searchFn(ctx, term, p)
	}
	return model.NewPaginatedResult([]model.PokemonListItem{}, 0, p.Page, p.Limit), nil
}

func (f *fakeService) ListByType(ctx context.Context, t string, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
	f.gotTerm, f.gotPage = t, p
	if f.listByTypeFn != nil {
		return f.listByTypeFn(ctx, t, p)
	}
	return model.NewPaginatedResult([]model.PokemonListItem{}, 0, p.Page, p.Limit), nil
}

func (f *fakeService) GetByID(ctx context.Context, id int32) (model.Pokemon, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return model.Pokemon{ID: id, Name: "pikachu"}, nil
}

func (f *fakeService) GetSpecies(ctx context.Context, id int32) (model.Species, error) {
	if f.getSpeciesFn != nil {
		return f.getSpeciesFn(ctx, id)
	}
	return model.Species{ID: id, Name: "pikachu"}, nil
}

// serve routes a request through a real ServeMux so that path patterns and
// r.PathValue behave exactly as they do in production. Calling the handler
// directly would leave path variables empty.
func serve(t *testing.T, pattern, target string, h http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(pattern, h)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
	return rec
}

func TestListStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantTitle  string
	}{
		{name: "defaults", target: "/api/v1/pokemon", wantStatus: http.StatusOK},
		{name: "explicit paging", target: "/api/v1/pokemon?page=2&limit=50", wantStatus: http.StatusOK},
		{
			name: "limit above the cap", target: "/api/v1/pokemon?limit=101",
			wantStatus: http.StatusBadRequest, wantTitle: "Invalid pagination",
		},
		{
			name: "page zero", target: "/api/v1/pokemon?page=0",
			wantStatus: http.StatusBadRequest, wantTitle: "Invalid pagination",
		},
		{
			name: "non-numeric page", target: "/api/v1/pokemon?page=abc",
			wantStatus: http.StatusBadRequest, wantTitle: "Invalid pagination",
		},
		{
			// A typo must not silently serve page 1.
			name: "misspelled parameter", target: "/api/v1/pokemon?pgae=2",
			wantStatus: http.StatusBadRequest, wantTitle: "Unknown query parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPokemonHandler(&fakeService{})
			rec := serve(t, "GET /api/v1/pokemon", tt.target, h.List)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body)
			}
			if tt.wantTitle == "" {
				return
			}

			if ct := rec.Header().Get("Content-Type"); ct != httperr.ContentType {
				t.Errorf("Content-Type = %q, want %q", ct, httperr.ContentType)
			}
			var p httperr.Problem
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
				t.Fatalf("error body is not valid problem+json: %v", err)
			}
			if p.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", p.Title, tt.wantTitle)
			}
			if p.Status != tt.wantStatus {
				t.Errorf("body status = %d, want %d", p.Status, tt.wantStatus)
			}
		})
	}
}

func TestListResponseShapeMatchesClientContract(t *testing.T) {
	svc := &fakeService{
		listFn: func(_ context.Context, p model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
			return model.NewPaginatedResult([]model.PokemonListItem{
				{ID: 25, Name: "pikachu", Types: []string{"electric"}},
			}, 1302, p.Page, p.Limit), nil
		},
	}

	rec := serve(t, "GET /api/v1/pokemon", "/api/v1/pokemon", NewPokemonHandler(svc).List)

	// Decoded into a raw map so the assertion is about the wire format, which
	// is the actual contract with the TypeScript client, not about Go types.
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	for _, key := range []string{"data", "total", "page", "limit", "totalPages", "hasNextPage", "hasPreviousPage"} {
		if _, ok := body[key]; !ok {
			t.Errorf("response is missing %q; the client's PaginatedResult<T> requires it", key)
		}
	}
	if body["totalPages"] != float64(66) {
		t.Errorf("totalPages = %v, want 66", body["totalPages"])
	}
}

func TestListWithTypeFilterDelegates(t *testing.T) {
	svc := &fakeService{}
	rec := serve(t, "GET /api/v1/pokemon", "/api/v1/pokemon?type=fire", NewPokemonHandler(svc).List)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if svc.gotTerm != "fire" {
		t.Errorf("type passed to service = %q, want %q", svc.gotTerm, "fire")
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		svcErr     error
		wantStatus int
	}{
		{name: "found", target: "/api/v1/pokemon/25", wantStatus: http.StatusOK},
		{
			name: "not found", target: "/api/v1/pokemon/9999",
			svcErr: model.ErrPokemonNotFound, wantStatus: http.StatusNotFound,
		},
		{name: "non-numeric id", target: "/api/v1/pokemon/abc", wantStatus: http.StatusBadRequest},
		{
			name: "invalid id rejected by the service", target: "/api/v1/pokemon/0",
			svcErr: model.ErrInvalidID, wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{
				getByIDFn: func(_ context.Context, id int32) (model.Pokemon, error) {
					if tt.svcErr != nil {
						return model.Pokemon{}, tt.svcErr
					}
					return model.Pokemon{ID: id, Name: "pikachu"}, nil
				},
			}

			rec := serve(t, "GET /api/v1/pokemon/{id}", tt.target, NewPokemonHandler(svc).GetByID)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d (body: %s)", rec.Code, tt.wantStatus, rec.Body)
			}
		})
	}
}

// TestInternalErrorsAreNotLeaked guards the boundary: an unrecognised error may
// carry a SQL fragment or a connection string, and returning it would hand a
// caller a description of the internals.
func TestInternalErrorsAreNotLeaked(t *testing.T) {
	secret := "pq: relation \"pokemon\" does not exist at 10.0.0.5:5432"
	svc := &fakeService{
		getByIDFn: func(context.Context, int32) (model.Pokemon, error) {
			return model.Pokemon{}, errString(secret)
		},
	}

	rec := serve(t, "GET /api/v1/pokemon/{id}", "/api/v1/pokemon/25", NewPokemonHandler(svc).GetByID)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if body := rec.Body.String(); strings.Contains(body, "does not exist") || strings.Contains(body, "10.0.0.5") {
		t.Errorf("internal error leaked to the client: %s", body)
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	svc := &fakeService{
		searchFn: func(context.Context, string, model.Pagination) (model.PaginatedResult[model.PokemonListItem], error) {
			return model.PaginatedResult[model.PokemonListItem]{}, model.ErrInvalidQuery
		},
	}

	rec := serve(t, "GET /api/v1/pokemon/search", "/api/v1/pokemon/search?q=x", NewPokemonHandler(svc).Search)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
