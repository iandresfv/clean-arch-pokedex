package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// DatasetVersion holds the counter the seeder bumps.
//
// It is read on every request and refreshed in the background, so it is atomic
// rather than mutex-guarded: the read is far hotter than the write.
type DatasetVersion struct {
	v atomic.Int64
}

// NewDatasetVersion starts at the given version.
func NewDatasetVersion(initial int64) *DatasetVersion {
	d := &DatasetVersion{}
	d.v.Store(initial)
	return d
}

// Load returns the current version.
func (d *DatasetVersion) Load() int64 { return d.v.Load() }

// Store replaces the current version.
func (d *DatasetVersion) Store(v int64) { d.v.Store(v) }

// HTTPCache adds ETag and Cache-Control to safe responses and answers
// conditional requests with 304.
//
// This is the largest latency win available on a read-only catalogue, and it
// costs one middleware. Rather than making a response faster to produce, it
// removes the response body from the wire entirely, and it works through any
// CDN or proxy in front of the API — no infrastructure required.
//
// The ETag folds in the dataset version, so a seeder run invalidates every
// cached entry at once without touching a single key.
func HTTPCache(maxAge time.Duration, version *DatasetVersion) Middleware {
	cacheControl := "public, max-age=" + strconv.Itoa(int(maxAge.Seconds())) +
		", stale-while-revalidate=60"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only safe methods are cacheable, and probes must never be cached
			// or an orchestrator would act on a stale answer.
			if r.Method != http.MethodGet || isProbePath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			buf := &bufferingWriter{ResponseWriter: w, body: &bytes.Buffer{}}
			next.ServeHTTP(buf, r)

			// Errors are not cacheable: a 404 stored for five minutes would
			// survive the seeding run that creates the missing row.
			if buf.status >= http.StatusBadRequest {
				buf.flush()
				return
			}

			etag := computeETag(buf.body.Bytes(), version.Load())
			w.Header().Set("ETag", etag)
			w.Header().Set("Cache-Control", cacheControl)
			// Responses vary by origin because CORS headers are origin-specific.
			w.Header().Add("Vary", "Origin")

			if matchesETag(r.Header.Get("If-None-Match"), etag) {
				// 304 must carry no body. The client reuses its stored copy,
				// so the payload never crosses the network at all.
				w.WriteHeader(http.StatusNotModified)
				return
			}

			buf.flush()
		})
	}
}

// computeETag derives a weak validator from the body and the dataset version.
//
// Weak (W/) is correct here: the tag guarantees semantic equivalence, not that
// the bytes are octet-for-octet identical, which is all a JSON response needs.
func computeETag(body []byte, version int64) string {
	h := sha256.New()
	h.Write([]byte(strconv.FormatInt(version, 10)))
	h.Write([]byte{0})
	h.Write(body)
	// 16 hex characters is ample: a collision would have to occur between two
	// responses for the same URL at the same dataset version.
	return `W/"` + hex.EncodeToString(h.Sum(nil))[:16] + `"`
}

// matchesETag implements the If-None-Match comparison, which accepts a
// comma-separated list and the "*" wildcard.
func matchesETag(header, etag string) bool {
	if header == "" {
		return false
	}
	if strings.TrimSpace(header) == "*" {
		return true
	}
	// SplitSeq iterates without allocating the intermediate slice.
	for candidate := range strings.SplitSeq(header, ",") {
		candidate = strings.TrimSpace(candidate)
		// Weak comparison: W/"x" and "x" are equivalent for this purpose.
		if strings.TrimPrefix(candidate, "W/") == strings.TrimPrefix(etag, "W/") {
			return true
		}
	}
	return false
}

// bufferingWriter captures the response so an ETag can be computed over the
// complete body before any byte is committed to the client.
type bufferingWriter struct {
	http.ResponseWriter
	body    *bytes.Buffer
	status  int
	flushed bool
}

func (bw *bufferingWriter) WriteHeader(status int) {
	if bw.status == 0 {
		bw.status = status
	}
}

func (bw *bufferingWriter) Write(b []byte) (int, error) {
	if bw.status == 0 {
		bw.status = http.StatusOK
	}
	return bw.body.Write(b)
}

// flush writes the buffered response through to the real writer.
func (bw *bufferingWriter) flush() {
	if bw.flushed {
		return
	}
	bw.flushed = true

	if bw.status == 0 {
		bw.status = http.StatusOK
	}
	bw.ResponseWriter.WriteHeader(bw.status)
	_, _ = bw.ResponseWriter.Write(bw.body.Bytes())
}

// Unwrap keeps http.ResponseController able to reach the real writer.
func (bw *bufferingWriter) Unwrap() http.ResponseWriter { return bw.ResponseWriter }
