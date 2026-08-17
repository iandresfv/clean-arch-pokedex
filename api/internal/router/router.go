// Package router registers every route on a single ServeMux.
package router

import (
	"net/http"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/handler"
)

// Handlers groups the handler set the router needs.
type Handlers struct {
	Pokemon *handler.PokemonHandler
	Type    *handler.TypeHandler
	Health  *handler.HealthHandler
	Docs    *handler.DocsHandler
}

// New registers every route and returns the mux.
//
// Registration is flat and in one place: a single greppable list of the API's
// surface. Nested sub-routers hide the real URL behind prefix composition, so
// answering "what serves /api/v1/pokemon/25/species?" means reading several
// files instead of one.
//
// Patterns use the method-aware syntax available since Go 1.22, which is what
// removed the last reason to reach for a third-party router.
func New(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Probes are unversioned: they describe the process, not the API contract.
	mux.HandleFunc("GET /health", h.Health.Health)
	mux.HandleFunc("GET /ready", h.Health.Ready)

	// More specific patterns win regardless of registration order, so /search
	// is not shadowed by /{id}.
	mux.HandleFunc("GET /api/v1/pokemon", h.Pokemon.List)
	mux.HandleFunc("GET /api/v1/pokemon/search", h.Pokemon.Search)
	mux.HandleFunc("GET /api/v1/pokemon/{id}", h.Pokemon.GetByID)
	mux.HandleFunc("GET /api/v1/pokemon/{id}/species", h.Pokemon.GetSpecies)

	mux.HandleFunc("GET /api/v1/types", h.Type.List)
	mux.HandleFunc("GET /api/v1/types/{name}/matchups", h.Type.GetMatchups)

	// Documentation is unversioned: it describes every version the binary
	// serves. Registered last so the more specific API patterns take priority.
	if h.Docs != nil {
		mux.HandleFunc("GET /docs", h.Docs.UI)
		mux.HandleFunc("GET /docs/{$}", h.Docs.UI)
		mux.HandleFunc("GET /docs/openapi.yaml", h.Docs.Spec)
		mux.HandleFunc("GET /docs/{asset}", h.Docs.Assets)
	}

	return mux
}
