# API Roadmap — Go Backend

> Idiomatic Go 1.25 REST API using only stdlib `net/http`, PostgreSQL, sqlc, and golang-migrate.
> Each stage maps to a single atomic commit on a feature branch.

---

## Previous Project Analysis (.dev-notes/.rest_api_go)

### What we keep
- Pure stdlib `net/http` with Go 1.22+ pattern routing (`GET /path/{id}`)
- Composable middleware chain pattern (`ApplyMiddleware`)
- Argon2id password hashing with constant-time comparison
- JWT authentication via httpOnly cookies
- Security headers (HSTS, CSP, X-Frame-Options, etc.)
- IP-based rate limiting with in-memory visitor map
- Docker multi-stage builds + Air hot reload for dev
- CORS with origin whitelisting

### What we fix
| Previous | Improved |
|----------|----------|
| MySQL/MariaDB | PostgreSQL |
| New DB connection per request | Single `sql.DB` connection pool at startup |
| Manual SQL string concatenation | sqlc type-safe generated queries |
| No migrations | golang-migrate with versioned SQL files |
| `fmt.Println` logging | `log/slog` structured logging (stdlib) |
| No tests | Table-driven tests per layer |
| No service layer (handler → repo) | Handler → Service → Repository |
| Reflection-based validation | Explicit validation in service layer |
| Nested router composition | Flat route registration on single ServeMux |
| `godotenv` for config | Struct-based config from env vars (no external deps) |
| No graceful shutdown | `os/signal` + `server.Shutdown(ctx)` |
| `init.sql` commented out | Proper migrations with up/down SQL |
| No health checks | `/health` (liveness) + `/ready` (DB connectivity) |
| Generic error messages | Domain errors with `errors.Is`/`errors.As` |
| Business logic in handlers | Thin handlers, logic in service layer |

---

## Git Flow & Release Strategy

Same Git Flow used for the frontend — consistency across the monorepo.

### Branching

- **develop** — integration branch, always deployable
- **feature/api-\*** — one per stage or group of related stages, branched from `develop`, merged back with `--no-ff`
- **fix/api-\*** — bug fixes, same flow as features
- **release/vX.Y.0** — cut from `develop` when a phase is complete, merged to both `main` and `develop`, tagged
- **main** — production-ready, only receives merges from release branches

### Release-to-Phase Mapping

One release per completed phase. Versioning continues from the frontend's v1.0.0:

| Release | Phase | Milestone |
|---------|-------|-----------|
| `v2.0.0` | Phase 1 — Foundation | Server boots, health check responds, project structure solid |
| `v2.1.0` | Phase 2 — Database | PostgreSQL connected, migrations run, sqlc generates, repo works |
| `v2.2.0` | Phase 3 — Core API | All Pokemon endpoints functional end-to-end |
| `v2.3.0` | Phase 4 — Middleware | Logging, recovery, CORS, request ID wired |
| `v2.4.0` | Phase 5 — Data Seeding | Database populated from PokeAPI, types + effectiveness working |
| `v2.5.0` | Phase 6 — Testing | Full test suite: unit, handler, integration |
| `v2.6.0` | Phase 7 — Docker & DX | Dockerized dev and prod environments |
| `v2.7.0` | Phase 8 — Hardening | Security headers, rate limiting, TLS/HTTP2 |
| `v2.8.0` | Phase 9 — API Documentation | OpenAPI spec, Swagger UI |

### Release Flow (per phase)

```
git checkout develop
git checkout -b release/v2.X.0
# bump version, update CHANGELOG.md
git commit -m "chore(release): bump version to 2.X.0"
git checkout main && git merge --no-ff release/v2.X.0
git tag v2.X.0
git checkout develop && git merge --no-ff release/v2.X.0
git push origin main develop --tags
git branch -d release/v2.X.0
```

### Feature Branch Naming

`feature/api-<short-description>` — e.g.:
- `feature/api-project-structure` (Stage 1)
- `feature/api-config` (Stage 2)
- `feature/api-http-server` (Stage 3)
- `feature/api-docker-postgres` (Stage 5)

Multiple stages can share a feature branch when they're tightly related (e.g., stages 13–16 middleware could be one branch `feature/api-middleware`).

---

## Target API Endpoints

Based on what the frontend needs:

```
GET  /health                          → Liveness check
GET  /ready                           → Readiness check (DB)

GET  /api/v1/pokemon                  → List (paginated, ?page=1&limit=20)
GET  /api/v1/pokemon/search           → Search by name (?q=pikachu&page=1&limit=20)
GET  /api/v1/pokemon/{id}             → Get by ID (full detail)
GET  /api/v1/pokemon/{id}/species     → Get species info

GET  /api/v1/types                    → List all 18 types
GET  /api/v1/types/{name}/matchups    → Type effectiveness (weaknesses, resistances, immunities)
```

---

## Phase 1 — Foundation

### Stage 1: Project structure and Go module
```
chore(api): scaffold project structure and initialize Go module
```

Create the folder layout and `go.mod`:

```
api/
├── cmd/server/
│   └── main.go              # Entry point (placeholder)
├── internal/
│   ├── config/              # Environment-based configuration
│   ├── handler/             # HTTP handlers (thin layer)
│   ├── middleware/           # Cross-cutting concerns
│   ├── model/               # Domain types, errors, value objects
│   ├── repository/          # Interface + PostgreSQL implementation
│   ├── router/              # Route registration
│   └── service/             # Business logic and orchestration
├── migrations/              # SQL migration files (golang-migrate)
├── sqlc/                    # sqlc config and queries
├── Makefile                 # Dev commands
├── .air.toml                # Hot reload config
├── .env.example             # Environment template (no secrets)
└── go.mod
```

### Stage 2: Configuration management
```
feat(api): add struct-based config from environment variables
```

- `internal/config/config.go` — `ServerConfig`, `DatabaseConfig`, `Config`
- `Load() (*Config, error)` reads from `os.Getenv` with sensible defaults
- Helper: `getEnv(key, fallback)`, `getEnvInt(key, fallback)`, `mustGetEnv(key)`
- No external deps (`godotenv` not needed — Docker and shell handle env vars)
- `.env.example` with all keys documented

### Stage 3: HTTP server with graceful shutdown and health endpoint
```
feat(api): add HTTP server with graceful shutdown and health check
```

- `cmd/server/main.go` — loads config, creates `http.Server`, listens
- Graceful shutdown: `os/signal.NotifyContext` + `server.Shutdown(ctx)` with 10s deadline
- `slog.Info("server starting", "addr", cfg.Server.Addr())` on startup
- `GET /health` returns `{"status": "ok"}` (liveness probe)
- Starts with plain HTTP — TLS/HTTPS and HTTP/2 added in Phase 8 (Hardening)

### Stage 4: Makefile for dev workflow
```
chore(api): add Makefile with common dev commands
```

```makefile
run          # go run cmd/server/main.go
build        # go build -o bin/server cmd/server/main.go
test         # go test ./...
test-cover   # go test -coverprofile -race ./...
lint         # golangci-lint run
fmt          # gofmt -w .
migrate-up   # golang-migrate up
migrate-down # golang-migrate down 1
migrate-new  # golang-migrate create -ext sql -dir migrations NAME
sqlc         # sqlc generate
docker-dev   # docker compose -f docker-compose.dev.yml up --build
docker-down  # docker compose -f docker-compose.dev.yml down
```

> **Release: `v2.0.0`** — create `release/v2.0.0`, tag, merge to `main` and `develop`

---

## Phase 2 — Database

### Stage 5: Docker Compose with PostgreSQL
```
chore(api): add Docker Compose with PostgreSQL for development
```

- `docker-compose.dev.yml` with PostgreSQL 17 (Alpine)
- Named volume for data persistence
- Health check: `pg_isready`
- Environment: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- Exposed on `localhost:5432` for local dev

### Stage 6: Database migrations — Pokemon schema
```
feat(api): add initial database migration for pokemon tables
```

- Install `golang-migrate` as CLI tool
- `migrations/000001_create_pokemon_tables.up.sql`:
  - `pokemon` table (id, name, order, base_experience, height, weight, sprite URLs, timestamps)
  - `pokemon_types` table (pokemon_id, type_name, slot) — normalized many-to-many
  - `pokemon_stats` table (pokemon_id, stat_name, base_stat)
  - `species` table (id, pokemon_id, generation, is_legendary, is_mythical, flavor_text, habitat, color, shape)
  - Proper indexes on name, type_name, pokemon_id FKs
- `migrations/000001_create_pokemon_tables.down.sql`: DROP tables in reverse order
- Add `migrate-up` / `migrate-down` to Makefile

### Stage 7: sqlc configuration and queries
```
feat(api): add sqlc config and Pokemon SQL queries
```

- `sqlc/sqlc.yaml` — config pointing to migrations and queries
- `sqlc/queries/pokemon.sql`:
  - `ListPokemon` — paginated with LIMIT/OFFSET, ordered by id
  - `GetPokemonByID` — joins types + stats in single query
  - `SearchPokemonByName` — ILIKE pattern match, paginated
  - `CountPokemon` — total count for pagination metadata
  - `CountPokemonByName` — count for search pagination
- `sqlc/queries/species.sql`:
  - `GetSpeciesByPokemonID` — species info for a pokemon
- `sqlc/queries/types.sql`:
  - `ListTypes` — all 18 unique type names
- Run `sqlc generate` → produces type-safe Go code in `internal/repository/postgres/`

### Stage 8: Repository layer — interface and PostgreSQL implementation
```
feat(api): add Pokemon repository interface and PostgreSQL implementation
```

- Interface defined in `internal/service/` (where it's used — idiomatic Go):
  ```go
  type PokemonRepository interface {
      List(ctx context.Context, limit, offset int) ([]model.Pokemon, error)
      GetByID(ctx context.Context, id int) (*model.Pokemon, error)
      SearchByName(ctx context.Context, name string, limit, offset int) ([]model.Pokemon, error)
      Count(ctx context.Context) (int, error)
      CountByName(ctx context.Context, name string) (int, error)
  }
  ```
- `internal/repository/postgres/pokemon_repo.go` — implements interface using sqlc-generated code
- Database connection pool created once in `main.go`:
  ```go
  db.SetMaxOpenConns(25)
  db.SetMaxIdleConns(5)
  db.SetConnMaxLifetime(5 * time.Minute)
  ```
- `GET /ready` endpoint wired — pings DB, returns 503 if unreachable

> **Release: `v2.1.0`** — create `release/v2.1.0`, tag, merge to `main` and `develop`

---

## Phase 3 — Core API

### Stage 9: Domain models and error definitions
```
feat(api): add domain models and custom error types
```

- `internal/model/pokemon.go`:
  - `Pokemon` struct (full detail, matches frontend's `PokemonDetailDTO`)
  - `PokemonListItem` struct (minimal, matches frontend's `PokemonListItemDTO`)
  - `Species` struct
  - `PaginatedResult[T]` generic struct
- `internal/model/errors.go`:
  - Sentinel errors: `ErrPokemonNotFound`, `ErrSpeciesNotFound`, `ErrInvalidPokemonID`
  - `errors.Is` / `errors.As` compatible

### Stage 10: Service layer — business logic
```
feat(api): add Pokemon service with business logic
```

- `internal/service/pokemon_service.go`:
  - `NewPokemonService(repo PokemonRepository, log *slog.Logger) *PokemonService`
  - `List(ctx, page, limit)` — validates pagination params, computes offset, returns `PaginatedResult`
  - `GetByID(ctx, id)` — validates ID, returns full Pokemon detail
  - `SearchByName(ctx, name, page, limit)` — trims/lowercases input, returns paginated results
  - `GetSpecies(ctx, pokemonID)` — returns species or `ErrSpeciesNotFound`
- Input validation happens here (not in handlers, not in repo)
- All methods accept `context.Context` as first param
- Errors wrapped with `fmt.Errorf("getting pokemon %d: %w", id, err)`

### Stage 11: Response helpers and handler layer
```
feat(api): add HTTP handlers and JSON response helpers
```

- `internal/handler/response.go`:
  - `writeJSON(w, status, data)` — sets Content-Type, encodes JSON
  - `writeError(w, status, message)` — standardized error envelope
  - `handleServiceError(w, err)` — maps domain errors to HTTP status codes
- `internal/handler/pokemon_handler.go`:
  - `NewPokemonHandler(svc *service.PokemonService, log *slog.Logger) *PokemonHandler`
  - `List` — parses `?page=` and `?limit=`, delegates to service
  - `GetByID` — parses `{id}` path value, delegates to service
  - `Search` — parses `?q=` query param, delegates to service
  - `GetSpecies` — parses `{id}`, delegates to service
- `internal/handler/health_handler.go`:
  - `Health` — returns liveness status
  - `Ready` — pings DB, returns readiness status

### Stage 12: Router — route registration
```
feat(api): add route registration with versioned API prefix
```

- `internal/router/router.go`:
  - `Setup(mux *http.ServeMux, ph *handler.PokemonHandler, hh *handler.HealthHandler)`
  - Flat registration:
    ```go
    mux.HandleFunc("GET /health", hh.Health)
    mux.HandleFunc("GET /ready", hh.Ready)
    mux.HandleFunc("GET /api/v1/pokemon", ph.List)
    mux.HandleFunc("GET /api/v1/pokemon/search", ph.Search)
    mux.HandleFunc("GET /api/v1/pokemon/{id}", ph.GetByID)
    mux.HandleFunc("GET /api/v1/pokemon/{id}/species", ph.GetSpecies)
    ```
- Wire everything in `cmd/server/main.go`: config → DB → repo → service → handler → router → server

> **Release: `v2.2.0`** — create `release/v2.2.0`, tag, merge to `main` and `develop`

---

## Phase 4 — Middleware

### Stage 13: Logging middleware with slog
```
feat(api): add structured request logging middleware
```

- `internal/middleware/logging.go`:
  - Logs: method, path, status code, duration, remote addr
  - Uses `slog.Info` for 2xx/3xx, `slog.Warn` for 4xx, `slog.Error` for 5xx
  - Custom `responseWriter` wrapper to capture status code

### Stage 14: Recovery and request ID middleware
```
feat(api): add panic recovery and request ID middleware
```

- `internal/middleware/recovery.go`:
  - Catches panics, logs stack trace with `slog.Error`, returns 500
  - Prevents server crash from unhandled panics
- `internal/middleware/request_id.go`:
  - Generates UUID per request, sets `X-Request-ID` header
  - Adds to context for downstream logging correlation

### Stage 15: CORS middleware
```
feat(api): add CORS middleware with configurable origins
```

- `internal/middleware/cors.go`:
  - Reads allowed origins from config
  - Handles preflight `OPTIONS` requests
  - Sets `Access-Control-Allow-*` headers
  - Supports credentials for cookie-based auth (future)

### Stage 16: Middleware chain and server wiring
```
feat(api): wire middleware chain in server startup
```

- `internal/middleware/chain.go`:
  - `Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler`
- Apply order in `main.go`: Recovery → RequestID → Logging → CORS → Router
- Verify middleware stack works end-to-end with `/health`

> **Release: `v2.3.0`** — create `release/v2.3.0`, tag, merge to `main` and `develop`

---

## Phase 5 — Data Seeding

### Stage 17: PokeAPI seeder — populate database from external API
```
feat(api): add CLI seeder to populate database from PokeAPI
```

- `cmd/seeder/main.go` — standalone CLI command, not part of the server
- Fetches from PokeAPI: list → parallel detail fetches (reuse frontend's N+1 pattern)
- Inserts into PostgreSQL using batch operations within a transaction
- Seeds: first 151 Pokemon (Gen 1) by default, configurable via flag
- Run via `make seed` or `go run cmd/seeder/main.go --limit=151`
- Idempotent: uses `ON CONFLICT DO UPDATE` (upsert)

### Stage 18: Types seeding and effectiveness data
```
feat(api): seed Pokemon types and add effectiveness endpoint
```

- Migration: `000002_create_type_effectiveness_table.up.sql`
  - `type_effectiveness` table (attacking_type, defending_type, multiplier)
  - 324 rows (18 x 18 matrix)
- Seeder populates the effectiveness matrix
- `GET /api/v1/types` — list all types
- `GET /api/v1/types/{name}/matchups` — returns weaknesses, resistances, immunities
- Service + handler + repository for types

> **Release: `v2.4.0`** — create `release/v2.4.0`, tag, merge to `main` and `develop`

---

## Phase 6 — Testing

### Stage 19: Service layer unit tests
```
test(api): add Pokemon service unit tests with table-driven patterns
```

- `internal/service/pokemon_service_test.go`
- Mock repository implementing the interface
- Table-driven tests for: List, GetByID, SearchByName, GetSpecies
- Edge cases: invalid ID, empty search, page out of range, not found
- `go test -race -count=1 ./internal/service/...`

### Stage 20: Handler layer tests
```
test(api): add handler tests with httptest
```

- `internal/handler/pokemon_handler_test.go`
- Uses `httptest.NewRecorder()` + `httptest.NewRequest()`
- Tests HTTP status codes, response body structure, error responses
- Mock service layer
- Table-driven: valid requests, invalid params, not found, server errors

### Stage 21: Repository integration tests
```
test(api): add repository integration tests with test database
```

- Uses test PostgreSQL (Docker or testcontainers)
- `internal/repository/postgres/pokemon_repo_test.go`
- Tests real SQL queries against actual database
- Setup: run migrations, seed test data
- Teardown: truncate tables
- Tests: List with pagination, GetByID, SearchByName, Count

> **Release: `v2.5.0`** — create `release/v2.5.0`, tag, merge to `main` and `develop`

---

## Phase 7 — Docker & DX

### Stage 22: Dockerfile — multi-stage production build
```
chore(api): add multi-stage Dockerfile for production
```

- Stage 1: `golang:1.25-alpine` — build static binary (`CGO_ENABLED=0`)
- Stage 2: `alpine:3.21` — minimal runtime image
- Non-root user, health check instruction
- Binary: `/app/server`

### Stage 23: Development Dockerfile with Air hot reload
```
chore(api): add dev Dockerfile and Air config for hot reload
```

- `Dockerfile.dev` — installs Air, watches `.go` files
- `.air.toml` — build command, exclude dirs, color output
- `docker-compose.dev.yml` updated: API service + PostgreSQL + volumes

### Stage 24: Full Docker Compose stack
```
chore(api): add production Docker Compose with all services
```

- `docker-compose.yml`:
  - PostgreSQL 17 with health check and named volume
  - API server (production image)
  - Network isolation
  - Restart policies
- Works with `make docker-dev` (dev) and `make docker-prod` (prod)

> **Release: `v2.6.0`** — create `release/v2.6.0`, tag, merge to `main` and `develop`

---

## Phase 8 — Hardening

### Stage 25: Security headers middleware
```
feat(api): add security headers middleware
```

- X-Content-Type-Options: nosniff
- X-Frame-Options: DENY
- Referrer-Policy: strict-origin-when-cross-origin
- Content-Security-Policy (API-appropriate)
- Cache-Control headers for API responses

### Stage 26: Rate limiting middleware
```
feat(api): add IP-based rate limiting middleware
```

- In-memory rate limiter (token bucket or sliding window)
- Per-IP tracking with configurable limits from config
- Auto-cleanup goroutine for stale entries
- Returns 429 Too Many Requests with `Retry-After` header
- Health/ready endpoints excluded from rate limiting

### Stage 27: Pagination helpers and query parameter validation
```
feat(api): add pagination helpers and strict query validation
```

- `internal/model/pagination.go`:
  - `ParsePagination(r)` — extracts page/limit with defaults and bounds
  - Max limit cap (100), min page (1)
  - `PaginatedResponse` with `page`, `limit`, `total`, `total_pages`, `has_next`, `has_previous`
- Reject unknown query parameters with 400

### Stage 28: TLS termination and HTTP/2 support
```
feat(api): add TLS support with self-signed certs for development
```

- `server.ListenAndServeTLS(certFile, keyFile)` — Go's stdlib automatically negotiates HTTP/2 when TLS is enabled (no extra config)
- Config: `TLS_ENABLED`, `TLS_CERT_PATH`, `TLS_KEY_PATH` env vars
- `make tls-cert` — generates self-signed cert for local dev (using `openssl` or Go's `crypto/x509`)
- Conditional: plain HTTP when `TLS_ENABLED=false` (default for dev), HTTPS when `true`
- Key concepts to understand and demonstrate:
  - TLS handshake (client hello, server hello, certificate exchange, key exchange)
  - Why HTTP/2 requires TLS in practice (ALPN negotiation during TLS handshake)
  - HTTP/2 multiplexing: multiple streams over a single TCP connection (no head-of-line blocking at HTTP level)
  - Difference between TLS termination at the app vs at a reverse proxy (Nginx, Caddy)
- Prod consideration: typically a reverse proxy handles TLS, but knowing how to do it in Go demonstrates depth

> **Release: `v2.7.0`** — create `release/v2.7.0`, tag, merge to `main` and `develop`

---

## Phase 9 — API Documentation

### Stage 29: OpenAPI 3.1 specification
```
docs(api): add OpenAPI 3.1 spec for all API endpoints
```

- Hand-written `api/docs/openapi.yaml` (not code-generated — demonstrates understanding of the spec)
- Document every endpoint: paths, parameters, request/response schemas, error envelopes, examples
- Reusable components: `#/components/schemas/Pokemon`, `#/components/schemas/PaginatedResponse`, `#/components/schemas/Error`
- `make validate-openapi` in Makefile for spec validation

### Stage 30: Swagger UI endpoint
```
feat(api): serve Swagger UI at /docs endpoint
```

- Download Swagger UI dist (static HTML/JS/CSS)
- Embed via Go 1.16+ `embed.FS` — no external dependencies at runtime
- `GET /docs` serves the interactive Swagger UI
- `GET /docs/openapi.yaml` serves the raw spec
- Route registration in router

> **Release: `v2.8.0`** — create `release/v2.8.0`, tag, merge to `main` and `develop`

---

## Commit Convention Reference

All commits follow the project's Conventional Commits format:

| Type | When |
|------|------|
| `feat(api)` | New feature or endpoint |
| `fix(api)` | Bug fix |
| `test(api)` | Adding or updating tests |
| `chore(api)` | Tooling, Docker, Makefile, config |
| `refactor(api)` | Code restructuring without behavior change |
| `docs(api)` | Documentation updates |
| `perf(api)` | Performance improvements |

---

## Dependencies (Final)

Only what's strictly necessary:

| Package | Purpose | Why not stdlib |
|---------|---------|----------------|
| `github.com/lib/pq` | PostgreSQL driver | stdlib has no PG driver |
| `github.com/golang-migrate/migrate/v4` | Schema migrations | No stdlib equivalent |
| `golang.org/x/crypto` | Argon2id hashing | stdlib only has bcrypt-adjacent |
| `github.com/golang-jwt/jwt/v5` | JWT tokens | No stdlib JWT |

Everything else is pure stdlib: `net/http`, `database/sql`, `encoding/json`, `log/slog`, `context`, `os/signal`, `fmt`, `errors`, `strconv`, `strings`, `time`, `sync`, `crypto/rand`.

**No web frameworks. No ORM. No DI containers. No config libraries.**
