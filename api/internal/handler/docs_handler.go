package handler

import (
	"embed"
	"net/http"
	"strings"
)

// The specification is compiled into the binary with embed.FS (standard library
// since Go 1.16), so the image stays self-contained: no volume mount to keep in
// step with the code, and no CDN request at runtime — which also means the
// documentation works behind the strict Content-Security-Policy this API sets.
//
//go:embed all:assets
var docsFS embed.FS

// DocsHandler serves the OpenAPI document and its interactive viewer.
type DocsHandler struct {
	spec []byte
	page []byte
}

// NewDocsHandler loads the embedded assets.
func NewDocsHandler() (*DocsHandler, error) {
	spec, err := docsFS.ReadFile("assets/openapi.yaml")
	if err != nil {
		return nil, err
	}
	page, err := docsFS.ReadFile("assets/index.html")
	if err != nil {
		return nil, err
	}
	return &DocsHandler{spec: spec, page: page}, nil
}

// UI serves GET /docs.
func (h *DocsHandler) UI(w http.ResponseWriter, r *http.Request) {
	// The viewer loads scripts and styles, which the API-wide CSP forbids. It
	// is relaxed for this one route rather than weakened globally.
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; script-src 'self' 'unsafe-inline'; "+
			"style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.page)
}

// Spec serves GET /docs/openapi.yaml.
func (h *DocsHandler) Spec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.spec)
}

// servableAssets is an allowlist mapping a request path to its content type.
//
// An allowlist rather than a lookup built from the URL: embed.FS already
// rejects traversal, but deriving both the file and its content type from
// caller input means any future addition to the asset directory becomes
// reachable, and possibly served under a type it should not have.
var servableAssets = map[string]struct {
	file        string
	contentType string
}{
	"swagger-ui.css":       {"assets/swagger-ui.css", "text/css; charset=utf-8"},
	"swagger-ui-bundle.js": {"assets/swagger-ui-bundle.js", "application/javascript; charset=utf-8"},
	"LICENSE":              {"assets/LICENSE", "text/plain; charset=utf-8"},
}

// Assets serves the viewer's static files.
func (h *DocsHandler) Assets(w http.ResponseWriter, r *http.Request) {
	asset, ok := servableAssets[strings.TrimPrefix(r.URL.Path, "/docs/")]
	if !ok {
		writeProblem(w, r, http.StatusNotFound, "Not Found", "no such documentation asset")
		return
	}

	data, err := docsFS.ReadFile(asset.file)
	if err != nil {
		writeProblem(w, r, http.StatusNotFound, "Not Found", "no such documentation asset")
		return
	}

	w.Header().Set("Content-Type", asset.contentType)
	// Immutable for a given build: the assets change only when the binary does.
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
}
