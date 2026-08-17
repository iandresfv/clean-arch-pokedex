// Package handler contains the HTTP layer: it parses requests, delegates to
// services, and serialises responses. It holds no business rules — any
// conditional about what a valid request means belongs in internal/service.
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/httperr"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/model"
	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// writeJSON serialises v as the response body.
//
// The body is encoded into a buffer before any header is written: encoding a
// value directly into the ResponseWriter commits a 200 status the moment the
// first byte flushes, so a failure halfway through would produce a truncated
// body under a success status.
func writeJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	body, err := json.Marshal(v)
	if err != nil {
		reqctx.Logger(r.Context()).Error("encoding response failed", "error", err)
		writeProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "")
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		// The client disconnected mid-write. Nothing can be sent now, so the
		// only useful action is to record it.
		reqctx.Logger(r.Context()).Debug("writing response failed", "error", err)
	}
}

// writeProblem emits an RFC 9457 error body through the shared envelope.
func writeProblem(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	httperr.Write(w, r, status, title, detail)
}

// handleServiceError maps a domain error onto an HTTP response.
//
// Unrecognised errors are logged in full and reported as a bare 500: the
// message may embed a SQL fragment or a connection string, and returning it
// would hand an attacker a description of the internals.
func handleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	log := reqctx.Logger(r.Context())

	switch {
	// Not-found details are written for the client rather than taken from
	// err.Error(): the wrapped chain repeats the identifier at every layer and
	// describes the call stack, which is information for the log, not the
	// caller. The resource is already identified by Instance.
	case errors.Is(err, model.ErrPokemonNotFound):
		log.Debug("pokemon not found", "error", err)
		writeProblem(w, r, http.StatusNotFound, "Pokemon not found",
			"no pokemon exists with the requested identifier")
	case errors.Is(err, model.ErrSpeciesNotFound):
		log.Debug("species not found", "error", err)
		writeProblem(w, r, http.StatusNotFound, "Species not found",
			"no species exists for the requested pokemon")
	case errors.Is(err, model.ErrTypeNotFound):
		log.Debug("type not found", "error", err)
		writeProblem(w, r, http.StatusNotFound, "Type not found",
			"no elemental type exists with the requested name")
	case errors.Is(err, model.ErrInvalidID):
		writeProblem(w, r, http.StatusBadRequest, "Invalid identifier", err.Error())
	case errors.Is(err, model.ErrInvalidPagination):
		writeProblem(w, r, http.StatusBadRequest, "Invalid pagination", err.Error())
	case errors.Is(err, model.ErrInvalidQuery):
		writeProblem(w, r, http.StatusBadRequest, "Invalid query", err.Error())
	case errors.Is(err, model.ErrUnknownParameter):
		writeProblem(w, r, http.StatusBadRequest, "Unknown query parameter", err.Error())
	case errors.Is(err, r.Context().Err()) && r.Context().Err() != nil:
		// The client went away or the request timed out. No response will be
		// read, so this is logged at debug and not treated as a server fault.
		log.Debug("request cancelled", "error", err)
	default:
		log.Error("unhandled service error", "error", err)
		writeProblem(w, r, http.StatusInternalServerError, "Internal Server Error",
			"an unexpected error occurred")
	}
}
