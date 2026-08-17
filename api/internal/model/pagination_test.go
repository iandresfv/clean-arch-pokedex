package model

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestNewPagination(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		limit   int
		wantErr bool
	}{
		{name: "defaults", page: DefaultPage, limit: DefaultLimit},
		{name: "at the maximum limit", page: 1, limit: MaxLimit},
		{name: "at the minimum limit", page: 1, limit: MinLimit},
		{name: "high but legal page", page: MaxPage, limit: 20},
		{name: "page zero", page: 0, limit: 20, wantErr: true},
		{name: "negative page", page: -1, limit: 20, wantErr: true},
		{name: "page above the cap", page: MaxPage + 1, limit: 20, wantErr: true},
		{name: "limit zero", page: 1, limit: 0, wantErr: true},
		// Without a cap, "?limit=1000000" forces a full table scan and
		// serialisation of every row in one request.
		{name: "limit above the cap", page: 1, limit: MaxLimit + 1, wantErr: true},
		{name: "absurd limit", page: 1, limit: 1_000_000, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPagination(tt.page, tt.limit)

			if tt.wantErr {
				if !errors.Is(err, ErrInvalidPagination) {
					t.Fatalf("error = %v, want ErrInvalidPagination", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPagination(%d, %d) error = %v", tt.page, tt.limit, err)
			}
			if got.Page != tt.page || got.Limit != tt.limit {
				t.Errorf("got %+v, want page %d limit %d", got, tt.page, tt.limit)
			}
		})
	}
}

func TestPaginationOffset(t *testing.T) {
	tests := []struct {
		page, limit int
		want        int32
	}{
		{page: 1, limit: 20, want: 0},
		{page: 2, limit: 20, want: 20},
		{page: 66, limit: 20, want: 1300},
		{page: 1, limit: 1, want: 0},
		{page: 100, limit: 100, want: 9900},
	}

	for _, tt := range tests {
		p, err := NewPagination(tt.page, tt.limit)
		if err != nil {
			t.Fatalf("NewPagination(%d, %d): %v", tt.page, tt.limit, err)
		}
		if got := p.Offset(); got != tt.want {
			t.Errorf("page %d limit %d: Offset() = %d, want %d", tt.page, tt.limit, got, tt.want)
		}
	}
}

// TestOffsetCannotOverflow proves the claim the bounds exist to support: the
// widest legal request stays inside int32, which is what makes the conversion
// in Offset safe.
func TestOffsetCannotOverflow(t *testing.T) {
	p, err := NewPagination(MaxPage, MaxLimit)
	if err != nil {
		t.Fatalf("the widest legal page was rejected: %v", err)
	}

	const maxInt32 = 1<<31 - 1
	if product := int64(MaxPage-1) * int64(MaxLimit); product > maxInt32 {
		t.Fatalf("MaxPage * MaxLimit = %d exceeds int32; Offset() would overflow", product)
	}
	if p.Offset() < 0 {
		t.Errorf("Offset() = %d; a negative offset means the conversion overflowed", p.Offset())
	}
}

func TestNewPaginatedResult(t *testing.T) {
	tests := []struct {
		name           string
		items          int
		total          int64
		page, limit    int
		wantTotalPages int
		wantNext       bool
		wantPrev       bool
	}{
		{name: "exact fit", items: 20, total: 40, page: 1, limit: 20, wantTotalPages: 2, wantNext: true},
		{name: "partial last page", items: 5, total: 25, page: 2, limit: 20, wantTotalPages: 2, wantPrev: true},
		{name: "single page", items: 3, total: 3, page: 1, limit: 20, wantTotalPages: 1},
		{name: "empty", items: 0, total: 0, page: 1, limit: 20, wantTotalPages: 0},
		{
			// Page 2 of an empty set has no previous page to go back to.
			name: "empty beyond page one", items: 0, total: 0, page: 2, limit: 20,
			wantTotalPages: 0, wantPrev: false,
		},
		{name: "one over the boundary", items: 1, total: 21, page: 2, limit: 20, wantTotalPages: 2, wantPrev: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := make([]PokemonListItem, tt.items)
			got := NewPaginatedResult(items, tt.total, tt.page, tt.limit)

			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", got.TotalPages, tt.wantTotalPages)
			}
			if got.HasNextPage != tt.wantNext {
				t.Errorf("HasNextPage = %v, want %v", got.HasNextPage, tt.wantNext)
			}
			if got.HasPreviousPage != tt.wantPrev {
				t.Errorf("HasPreviousPage = %v, want %v", got.HasPreviousPage, tt.wantPrev)
			}
		})
	}
}

// TestNilDataMarshalsAsEmptyArray guards the contract with the client: a nil
// slice encodes as JSON null, which would force every caller to null-check each
// page before iterating it.
func TestNilDataMarshalsAsEmptyArray(t *testing.T) {
	got := NewPaginatedResult[PokemonListItem](nil, 0, 1, 20)

	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
	if decoded["data"] == nil {
		t.Errorf("data encoded as null: %s", body)
	}
}

// TestJSONFieldNamesMatchClientContract pins the wire format. The client's
// PaginatedResult<T> reads these exact keys, so a rename here breaks it
// silently — the fields would decode as undefined rather than error.
func TestJSONFieldNamesMatchClientContract(t *testing.T) {
	body, err := json.Marshal(NewPaginatedResult([]PokemonListItem{{ID: 1}}, 1, 1, 20))
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}

	for _, key := range []string{"data", "total", "page", "limit", "totalPages", "hasNextPage", "hasPreviousPage"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("missing key %q; the client contract requires it", key)
		}
	}
}

func TestPokemonJSONFieldNames(t *testing.T) {
	body, err := json.Marshal(Pokemon{ID: 25, Name: "pikachu"})
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}

	for _, key := range []string{"id", "name", "types", "stats", "heightDm", "weightHg", "baseExperience", "sprites", "species"} {
		if _, ok := decoded[key]; !ok {
			t.Errorf("missing key %q", key)
		}
	}
}
