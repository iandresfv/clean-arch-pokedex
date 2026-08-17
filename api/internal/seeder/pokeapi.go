// Package seeder populates the database from PokeAPI.
//
// It is a separate binary rather than an endpoint: seeding is an operational
// task, and exposing it over HTTP would mean shipping a route capable of
// hammering a third-party API and rewriting the catalogue.
package seeder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is PokeAPI's v2 root.
const DefaultBaseURL = "https://pokeapi.co/api/v2"

// Client fetches resources from PokeAPI with bounded retries.
type Client struct {
	baseURL    string
	http       *http.Client
	log        *slog.Logger
	maxRetries int
}

// NewClient builds a PokeAPI client.
//
// The http.Client carries an explicit timeout: the zero-value client waits
// forever, so a single unresponsive connection would stall a worker for the
// lifetime of the process.
func NewClient(baseURL string, log *slog.Logger) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 16,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		log:        log,
		maxRetries: 4,
	}
}

// ListEntry is one row of a PokeAPI index response.
type ListEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type listResponse struct {
	Count   int         `json:"count"`
	Results []ListEntry `json:"results"`
}

// NamedResource is PokeAPI's ubiquitous {name, url} reference.
type NamedResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// PokemonResponse is the subset of /pokemon/{id} this project stores.
type PokemonResponse struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Order          int    `json:"order"`
	BaseExperience *int   `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Sprites        struct {
		FrontDefault *string `json:"front_default"`
		FrontShiny   *string `json:"front_shiny"`
		BackDefault  *string `json:"back_default"`
		BackShiny    *string `json:"back_shiny"`
		Other        struct {
			OfficialArtwork struct {
				FrontDefault *string `json:"front_default"`
			} `json:"official-artwork"`
		} `json:"other"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int           `json:"base_stat"`
		Stat     NamedResource `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int           `json:"slot"`
		Type NamedResource `json:"type"`
	} `json:"types"`
	Species NamedResource `json:"species"`
}

// SpeciesResponse is the subset of /pokemon-species/{id} this project stores.
type SpeciesResponse struct {
	ID                int            `json:"id"`
	Name              string         `json:"name"`
	Generation        NamedResource  `json:"generation"`
	Habitat           *NamedResource `json:"habitat"`
	Color             *NamedResource `json:"color"`
	Shape             *NamedResource `json:"shape"`
	IsLegendary       bool           `json:"is_legendary"`
	IsMythical        bool           `json:"is_mythical"`
	FlavorTextEntries []struct {
		FlavorText string        `json:"flavor_text"`
		Language   NamedResource `json:"language"`
	} `json:"flavor_text_entries"`
}

// ListPokemon returns the first limit index entries.
func (c *Client) ListPokemon(ctx context.Context, limit int) ([]ListEntry, error) {
	var resp listResponse
	url := fmt.Sprintf("%s/pokemon?limit=%d&offset=0", c.baseURL, limit)
	if err := c.getJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("listing pokemon: %w", err)
	}
	return resp.Results, nil
}

// GetPokemon fetches one Pokemon by name or id.
func (c *Client) GetPokemon(ctx context.Context, ref string) (*PokemonResponse, error) {
	var p PokemonResponse
	if err := c.getJSON(ctx, c.resolve("pokemon", ref), &p); err != nil {
		return nil, fmt.Errorf("getting pokemon %s: %w", ref, err)
	}
	return &p, nil
}

// GetSpecies fetches one species by name, id, or full URL.
func (c *Client) GetSpecies(ctx context.Context, ref string) (*SpeciesResponse, error) {
	var s SpeciesResponse
	if err := c.getJSON(ctx, c.resolve("pokemon-species", ref), &s); err != nil {
		return nil, fmt.Errorf("getting species %s: %w", ref, err)
	}
	return &s, nil
}

// resolve accepts either a bare identifier or an absolute PokeAPI URL, since
// index responses embed full URLs.
func (c *Client) resolve(resource, ref string) string {
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return ref
	}
	return fmt.Sprintf("%s/%s/%s", c.baseURL, resource, ref)
}

// errNotFound marks a 404 so callers can skip a missing resource instead of
// aborting the run.
var errNotFound = errors.New("resource not found")

// getJSON performs a GET with retries.
//
// Only transient failures are retried: 5xx, 429, and transport errors. A 404 is
// a permanent answer, and retrying it would multiply load on the upstream for
// no chance of a different result.
func (c *Client) getJSON(ctx context.Context, url string, dst any) error {
	delay := 500 * time.Millisecond

	for attempt := 0; ; attempt++ {
		retryAfter, err := c.attempt(ctx, url, dst)
		if err == nil {
			return nil
		}
		if errors.Is(err, errNotFound) || errors.Is(err, context.Canceled) {
			return err
		}
		if attempt >= c.maxRetries {
			return fmt.Errorf("after %d attempts: %w", attempt+1, err)
		}

		wait := delay
		// A server that states how long to wait knows better than our backoff
		// curve; ignoring Retry-After is how a client earns a ban.
		if retryAfter > 0 {
			wait = retryAfter
		}
		c.log.Warn("pokeapi request failed, retrying",
			"url", url, "attempt", attempt+1, "retry_in", wait.String(), "error", err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
		delay *= 2
	}
}

// attempt performs a single request, returning any server-requested delay.
func (c *Client) attempt(ctx context.Context, url string, dst any) (time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "clean-arch-pokedex-seeder/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("requesting %s: %w", url, err)
	}
	defer func() {
		// Draining before closing lets the connection return to the pool
		// instead of being discarded, which matters across ~2600 requests.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
	}()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return 0, fmt.Errorf("%s: %w", url, errNotFound)
	case resp.StatusCode == http.StatusTooManyRequests:
		return parseRetryAfter(resp.Header.Get("Retry-After")), fmt.Errorf("rate limited by %s", url)
	case resp.StatusCode >= 500:
		return 0, fmt.Errorf("upstream error %d from %s", resp.StatusCode, url)
	case resp.StatusCode != http.StatusOK:
		return 0, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return 0, fmt.Errorf("decoding response from %s: %w", url, err)
	}
	return 0, nil
}

// parseRetryAfter reads the header in its seconds form. The HTTP-date form is
// not handled: PokeAPI uses seconds, and guessing at clock skew would be worse
// than falling back to the backoff curve.
func parseRetryAfter(v string) time.Duration {
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}
	return 0
}
