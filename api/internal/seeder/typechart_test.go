package seeder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestTypeChartMatchesClientDomain guards the one piece of knowledge that is
// deliberately duplicated in this repository.
//
// The client keeps the 18x18 effectiveness chart in a domain service, because
// moving the calculation server-side would leave its domain layer without
// behaviour. The database keeps the same matrix because it is the system of
// record. Duplication is acceptable only while something proves the two copies
// agree — without this test the first change to either would silently produce
// two different answers to the same question.
func TestTypeChartMatchesClientDomain(t *testing.T) {
	clientChart := parseClientTypeChart(t)
	migrationChart := parseMigrationMatrix(t)

	if len(clientChart) != 18 {
		t.Fatalf("client chart has %d attacking types, want 18", len(clientChart))
	}

	for attacker, row := range clientChart {
		for defender, want := range row {
			got, ok := migrationChart[attacker+"->"+defender]
			if !ok {
				t.Errorf("migration is missing %s -> %s", attacker, defender)
				continue
			}
			if got != want {
				t.Errorf("%s -> %s: migration has %v, client domain has %v",
					attacker, defender, got, want)
			}
		}
	}

	if len(migrationChart) != 324 {
		t.Errorf("migration defines %d cells, want 324 (18 x 18)", len(migrationChart))
	}
}

// parseClientTypeChart reads TYPE_CHART out of the client's domain service.
func parseClientTypeChart(t *testing.T) map[string]map[string]float64 {
	t.Helper()

	path := filepath.Join("..", "..", "..", "client", "src", "domain", "pokemon", "TypeEffectivenessService.ts")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("client source not available (%v); skipping cross-check", err)
	}

	start := strings.Index(string(src), "const TYPE_CHART")
	if start == -1 {
		t.Fatal("TYPE_CHART not found in the client domain service")
	}
	brace := strings.Index(string(src)[start:], "{") + start

	depth := 0
	end := brace
	for i := brace; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
		}
		if depth == 0 {
			end = i
			break
		}
	}

	// Turn the TypeScript object literal into JSON: quote bare keys, drop the
	// trailing commas TypeScript allows and JSON does not.
	body := string(src[brace : end+1])
	body = regexp.MustCompile(`(\w+)\s*:`).ReplaceAllString(body, `"$1":`)
	body = regexp.MustCompile(`,(\s*[}\]])`).ReplaceAllString(body, `$1`)

	var chart map[string]map[string]float64
	if err := json.Unmarshal([]byte(body), &chart); err != nil {
		t.Fatalf("parsing client TYPE_CHART: %v", err)
	}
	return chart
}

// parseMigrationMatrix reads the effectiveness rows out of the SQL migration,
// using the trailing "-- attacker -> defender" comment each row carries.
func parseMigrationMatrix(t *testing.T) map[string]float64 {
	t.Helper()

	path := filepath.Join("..", "..", "migrations", "000002_create_type_tables.up.sql")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading migration: %v", err)
	}

	re := regexp.MustCompile(`\(\s*\d+,\s*\d+,\s*([\d.]+)\)[,;]\s*--\s*(\w+) -> (\w+)`)
	matches := re.FindAllStringSubmatch(string(src), -1)

	out := make(map[string]float64, len(matches))
	for _, m := range matches {
		v, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			t.Fatalf("parsing multiplier %q: %v", m[1], err)
		}
		out[m[2]+"->"+m[3]] = v
	}
	return out
}
