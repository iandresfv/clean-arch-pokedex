// Package httperr defines the API's error envelope.
//
// It lives in its own package because both the handler layer (404, 400) and the
// middleware layer (500 from recovery, 429 from the rate limiter) emit errors.
// If each wrote its own shape, a client would face two different error formats
// depending on how deep the request got before failing.
package httperr

import (
	"encoding/json"
	"net/http"

	"github.com/iandresfv/clean-arch-pokedex/api/internal/reqctx"
)

// ContentType is the media type RFC 9457 defines for problem details.
const ContentType = "application/problem+json"

// baseURI namespaces the machine-readable error identifiers.
const baseURI = "https://github.com/iandresfv/clean-arch-pokedex/errors/"

// Problem is an RFC 9457 problem details object.
//
// Type is a stable identifier a client may branch on; Title and Detail are
// human-facing prose that may be reworded without breaking anyone.
type Problem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	RequestID string `json:"requestId,omitempty"`
}

// Write emits a problem response.
func Write(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	p := Problem{
		Type:      baseURI + slug(status),
		Title:     title,
		Status:    status,
		Detail:    detail,
		Instance:  r.URL.Path,
		RequestID: reqctx.RequestID(r.Context()),
	}

	body, err := json.Marshal(p)
	if err != nil {
		http.Error(w, title, status)
		return
	}

	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

// slug maps a status code onto the stable identifier used in the type URI.
func slug(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad-request"
	case http.StatusNotFound:
		return "not-found"
	case http.StatusRequestTimeout:
		return "timeout"
	case http.StatusTooManyRequests:
		return "rate-limited"
	case http.StatusServiceUnavailable:
		return "unavailable"
	default:
		return "internal"
	}
}
