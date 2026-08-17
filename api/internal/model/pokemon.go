// Package model defines the domain types the API exposes and the sentinel
// errors every layer agrees on.
//
// JSON field names are camelCase because the sole consumer is the TypeScript
// client, whose DTOs already use that convention. Emitting snake_case would add
// a renaming layer in the client adapter that buys nothing.
package model

// PokemonListItem is an entry in a list or search response.
//
// It carries every field of a full Pokemon except its species, which is
// deliberate rather than wasteful: the client's repository port returns domain
// entities, and a domain Pokemon cannot be constructed without its stats,
// measurements and order. A slimmer shape would force the client to fetch each
// row's detail separately — the N+1 pattern this API exists to remove.
//
// SpriteURL mirrors the best available artwork, matching the client's
// Sprites.getBestQuality so both data sources render identically.
type PokemonListItem struct {
	ID             int32    `json:"id"`
	Name           string   `json:"name"`
	PokedexOrder   int32    `json:"pokedexOrder"`
	Types          []string `json:"types"`
	Stats          Stats    `json:"stats"`
	HeightDm       int32    `json:"heightDm"`
	WeightHg       int32    `json:"weightHg"`
	BaseExperience *int32   `json:"baseExperience"`
	Sprites        Sprites  `json:"sprites"`
	SpriteURL      *string  `json:"spriteUrl"`
}

// Stats holds the six base stats. They are a fixed set defined by the games,
// not a user-extensible collection, which is why they are six fields here and
// six columns in the schema rather than a key/value structure.
type Stats struct {
	HP             int16 `json:"hp"`
	Attack         int16 `json:"attack"`
	Defense        int16 `json:"defense"`
	SpecialAttack  int16 `json:"specialAttack"`
	SpecialDefense int16 `json:"specialDefense"`
	Speed          int16 `json:"speed"`
}

// Sprites holds the artwork URLs. Every field is nullable because PokeAPI does
// not provide every sprite for every Pokemon.
type Sprites struct {
	FrontDefault    *string `json:"frontDefault"`
	FrontShiny      *string `json:"frontShiny"`
	BackDefault     *string `json:"backDefault"`
	BackShiny       *string `json:"backShiny"`
	OfficialArtwork *string `json:"officialArtwork"`
}

// Pokemon is the full detail projection.
//
// Height and weight stay in PokeAPI's raw units. Converting them to a display
// string is a presentation concern that already lives in the client's
// PhysicalMeasurement value object, and duplicating it here would create two
// formatting rules to keep in step.
type Pokemon struct {
	ID             int32    `json:"id"`
	Name           string   `json:"name"`
	PokedexOrder   int32    `json:"pokedexOrder"`
	Types          []string `json:"types"`
	Stats          Stats    `json:"stats"`
	HeightDm       int32    `json:"heightDm"`
	WeightHg       int32    `json:"weightHg"`
	BaseExperience *int32   `json:"baseExperience"`
	Sprites        Sprites  `json:"sprites"`
	Species        *Species `json:"species"`
}

// Species is the conceptual creature a Pokemon is a form of.
type Species struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Generation  int16   `json:"generation"`
	FlavorText  string  `json:"flavorText"`
	Habitat     *string `json:"habitat"`
	Color       *string `json:"color"`
	Shape       *string `json:"shape"`
	IsLegendary bool    `json:"isLegendary"`
	IsMythical  bool    `json:"isMythical"`
}

// Type is one of the eighteen elemental types.
type Type struct {
	ID   int16  `json:"id"`
	Name string `json:"name"`
}

// Matchups describes how much damage a defending type takes. Neutral matchups
// are omitted: they are the default and listing them would trade clarity for
// completeness nobody renders.
type Matchups struct {
	Type        string   `json:"type"`
	Weaknesses  []string `json:"weaknesses"`
	Resistances []string `json:"resistances"`
	Immunities  []string `json:"immunities"`
}

// PaginatedResult wraps a page of results. Field names mirror the client's
// PaginatedResult<T> exactly so the adapter needs no translation.
type PaginatedResult[T any] struct {
	Data            []T   `json:"data"`
	Total           int64 `json:"total"`
	Page            int   `json:"page"`
	Limit           int   `json:"limit"`
	TotalPages      int   `json:"totalPages"`
	HasNextPage     bool  `json:"hasNextPage"`
	HasPreviousPage bool  `json:"hasPreviousPage"`
}

// NewPaginatedResult computes the derived pagination fields in one place, so
// the handler, the service and the tests cannot disagree about what
// "hasNextPage" means.
//
// Data is normalised to an empty slice: a nil slice marshals to JSON null, and
// a client iterating the response would have to null-check every page.
func NewPaginatedResult[T any](data []T, total int64, page, limit int) PaginatedResult[T] {
	if data == nil {
		data = []T{}
	}

	totalPages := 0
	if limit > 0 {
		totalPages = int((total + int64(limit) - 1) / int64(limit))
	}

	return PaginatedResult[T]{
		Data:            data,
		Total:           total,
		Page:            page,
		Limit:           limit,
		TotalPages:      totalPages,
		HasNextPage:     page < totalPages,
		HasPreviousPage: page > 1 && total > 0,
	}
}
