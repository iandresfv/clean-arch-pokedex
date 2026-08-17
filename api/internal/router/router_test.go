package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/handler"
)

// registeredRoutes lists the API surface this package exposes.
//
// Kept as data so two things can be checked against it: that the mux actually
// resolves each pattern, and that the OpenAPI document describes it. A hand
// written specification is only a contract while something proves it matches
// the implementation.
var registeredRoutes = []struct {
	method string
	// pattern is the ServeMux pattern.
	pattern string
	// specPath is the same route in OpenAPI's notation.
	specPath string
	// sample is a concrete URL used to check the mux resolves it.
	sample string
}{
	{http.MethodGet, "/health", "/health", "/health"},
	{http.MethodGet, "/ready", "/ready", "/ready"},
	{http.MethodGet, "/api/v1/pokemon", "/api/v1/pokemon", "/api/v1/pokemon"},
	{http.MethodGet, "/api/v1/pokemon/search", "/api/v1/pokemon/search", "/api/v1/pokemon/search?q=pika"},
	{http.MethodGet, "/api/v1/pokemon/{id}", "/api/v1/pokemon/{id}", "/api/v1/pokemon/25"},
	{http.MethodGet, "/api/v1/pokemon/{id}/species", "/api/v1/pokemon/{id}/species", "/api/v1/pokemon/25/species"},
	{http.MethodGet, "/api/v1/types", "/api/v1/types", "/api/v1/types"},
	{http.MethodGet, "/api/v1/types/{name}/matchups", "/api/v1/types/{name}/matchups", "/api/v1/types/fire/matchups"},
}

// TestEveryRouteIsDocumented is the check that keeps a hand-written
// specification honest: adding an endpoint without documenting it fails here
// rather than being discovered by a client.
func TestEveryRouteIsDocumented(t *testing.T) {
	specPath := filepath.Join("..", "..", "docs", "openapi.yaml")
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("reading the OpenAPI document: %v", err)
	}

	documented := documentedPaths(string(spec))

	for _, route := range registeredRoutes {
		if _, ok := documented[route.specPath]; !ok {
			t.Errorf("route %s %s is served but absent from openapi.yaml", route.method, route.pattern)
		}
	}

	// The reverse direction matters just as much: a documented endpoint that
	// does not exist sends clients at a 404.
	served := make(map[string]struct{}, len(registeredRoutes))
	for _, route := range registeredRoutes {
		served[route.specPath] = struct{}{}
	}
	for path := range documented {
		if _, ok := served[path]; !ok {
			t.Errorf("openapi.yaml documents %s, which no route serves", path)
		}
	}
}

// TestRoutesResolve verifies the mux matches each pattern, which catches a
// shadowed route: without Go 1.22's specificity rules, /search would be
// swallowed by /{id}.
func TestRoutesResolve(t *testing.T) {
	mux := New(Handlers{
		Pokemon: handler.NewPokemonHandler(nil),
		Type:    handler.NewTypeHandler(nil),
		Health:  handler.NewHealthHandler(nil, "test"),
	})

	for _, route := range registeredRoutes {
		t.Run(route.method+" "+route.pattern, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), route.method, route.sample, nil)

			_, pattern := mux.Handler(req)
			if pattern == "" {
				t.Fatalf("%s resolves to no handler", route.sample)
			}
			if !strings.HasSuffix(pattern, route.pattern) {
				t.Errorf("%s resolved to pattern %q, want %q", route.sample, pattern, route.pattern)
			}
		})
	}
}

// TestSearchIsNotShadowedByIDPattern pins the specific ordering hazard.
func TestSearchIsNotShadowedByIDPattern(t *testing.T) {
	mux := New(Handlers{
		Pokemon: handler.NewPokemonHandler(nil),
		Type:    handler.NewTypeHandler(nil),
		Health:  handler.NewHealthHandler(nil, "test"),
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/pokemon/search?q=pika", nil)
	_, pattern := mux.Handler(req)

	if strings.Contains(pattern, "{id}") {
		t.Errorf("/search resolved to %q; the literal segment must win over the wildcard", pattern)
	}
}

// TestUnknownRouteReturns404 confirms nothing catches all paths.
func TestUnknownRouteReturns404(t *testing.T) {
	mux := New(Handlers{
		Pokemon: handler.NewPokemonHandler(nil),
		Type:    handler.NewTypeHandler(nil),
		Health:  handler.NewHealthHandler(nil, "test"),
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/unknown", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

// TestWrongMethodIsRejected relies on the method-aware patterns introduced in
// Go 1.22, which are what removed the last reason to use a third-party router.
func TestWrongMethodIsRejected(t *testing.T) {
	mux := New(Handlers{
		Pokemon: handler.NewPokemonHandler(nil),
		Type:    handler.NewTypeHandler(nil),
		Health:  handler.NewHealthHandler(nil, "test"),
	})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/pokemon", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

// documentedPaths extracts the top-level keys of the specification's `paths`
// block. A YAML parser would be more precise, but it would also be the only
// dependency this module carries purely for a test.
func documentedPaths(spec string) map[string]struct{} {
	start := strings.Index(spec, "\npaths:")
	if start == -1 {
		return nil
	}
	end := strings.Index(spec[start+1:], "\ncomponents:")
	body := spec[start:]
	if end != -1 {
		body = spec[start : start+1+end]
	}

	// Path keys sit at exactly two spaces of indentation and begin with a slash.
	re := regexp.MustCompile(`(?m)^  (/[^\s:]*):`)
	out := map[string]struct{}{}
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		out[m[1]] = struct{}{}
	}
	return out
}
