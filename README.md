# Clean Architecture Pokédex

A full-stack Pokédex built to demonstrate architecture that survives change: a
React client structured with Hexagonal Architecture, and a Go API that replaced
its original data source without a single line changing in the domain,
application, or presentation layers.

<p align="center">
  <a href="https://www.typescriptlang.org/"><img alt="TypeScript" src="https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript&logoColor=white"></a>
  <a href="https://react.dev/"><img alt="React" src="https://img.shields.io/badge/React-19.2-61DAFB?logo=react&logoColor=black"></a>
  <a href="https://go.dev/"><img alt="Go" src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white"></a>
  <a href="https://www.postgresql.org/"><img alt="PostgreSQL" src="https://img.shields.io/badge/PostgreSQL-18-4169E1?logo=postgresql&logoColor=white"></a>
  <a href="https://kubernetes.io/"><img alt="Kubernetes" src="https://img.shields.io/badge/Kubernetes-1.35-326CE5?logo=kubernetes&logoColor=white"></a>
  <a href="https://opensource.org/licenses/MIT"><img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg"></a>
</p>

---

## The point of this project

Most projects claim their architecture makes change cheap. This one has the
receipt.

The client was originally built against [PokeAPI](https://pokeapi.co/). A Go
backend was then written to replace it — a different protocol shape, a different
pagination contract, a different error format. Swapping it in required:

| Layer | Changes |
|---|---|
| **Domain** (entities, value objects, domain services) | none |
| **Application** (use cases, ports, DTOs) | none |
| **Presentation** (components, hooks, stores, routes) | none |
| **Infrastructure** | one new adapter implementing the existing port |
| **Composition root** | one line selecting which adapter to build |

**203 client tests passed without modification** after the switch. That is the
assertion; everything else in this README is the supporting detail.

The measured effect on the client:

| Operation | Against PokeAPI | Against this API |
|---|---|---|
| List a page of 20 | 1 index request + 20 detail requests | **1 request** |
| Search by name | download all 1302 names, filter in the browser | **1 request** |
| Pagination metadata | computed client-side | returned by the server |
| Species detail | a separate request per Pokémon | embedded in the response |

---

## Technology

### Frontend — `client/`

| | Technology | Role |
|---|---|---|
| <img src="https://cdn.simpleicons.org/react/61DAFB" width="18"> | [React 19](https://react.dev/) | UI runtime |
| <img src="https://cdn.simpleicons.org/typescript/3178C6" width="18"> | [TypeScript 5.9](https://www.typescriptlang.org/) | Strict types, no `any` |
| <img src="https://cdn.simpleicons.org/vite/646CFF" width="18"> | [Vite 7](https://vite.dev/) | Dev server and bundler |
| <img src="https://cdn.simpleicons.org/reactquery/FF4154" width="18"> | [TanStack Query](https://tanstack.com/query) | Server-state caching, retries, invalidation |
| <img src="https://cdn.simpleicons.org/reactrouter/CA4245" width="18"> | [React Router](https://reactrouter.com/) | Routing |
| <img src="https://cdn.simpleicons.org/redux/764ABC" width="18"> | [Zustand](https://zustand.docs.pmnd.rs/) | Client-only state (favourites) |
| <img src="https://cdn.simpleicons.org/tailwindcss/06B6D4" width="18"> | [Tailwind CSS](https://tailwindcss.com/) | Styling |
| <img src="https://cdn.simpleicons.org/shadcnui/000000" width="18"> | [shadcn/ui](https://ui.shadcn.com/) + [Radix](https://www.radix-ui.com/) | Accessible component primitives, owned as source |
| <img src="https://cdn.simpleicons.org/vitest/6E9F18" width="18"> | [Vitest](https://vitest.dev/) + [Testing Library](https://testing-library.com/) | Unit and integration tests |
| <img src="https://cdn.simpleicons.org/playwright/2EAD33" width="18"> | [Playwright](https://playwright.dev/) | End-to-end tests |

### Backend — `api/`

| | Technology | Role |
|---|---|---|
| <img src="https://cdn.simpleicons.org/go/00ADD8" width="18"> | [Go 1.26](https://go.dev/) | Language and, deliberately, most of the stack |
| | [`net/http`](https://pkg.go.dev/net/http) | HTTP server and routing — **no web framework** |
| | [`log/slog`](https://pkg.go.dev/log/slog) | Structured logging |
| <img src="https://cdn.simpleicons.org/postgresql/4169E1" width="18"> | [PostgreSQL 18](https://www.postgresql.org/) | Storage |
| | [pgx v5](https://github.com/jackc/pgx) | PostgreSQL driver and connection pool |
| | [sqlc](https://sqlc.dev/) | Type-safe Go generated from SQL — **no ORM** |
| | [golang-migrate](https://github.com/golang-migrate/migrate) | Versioned schema migrations |
| <img src="https://cdn.simpleicons.org/redis/FF4438" width="18"> | [Redis 8](https://redis.io/) | Rate-limiter state shared across replicas |
| <img src="https://cdn.simpleicons.org/docker/2496ED" width="18"> | [Docker](https://www.docker.com/) + [Compose](https://docs.docker.com/compose/) | Local environment, production image |
| <img src="https://cdn.simpleicons.org/kubernetes/326CE5" width="18"> | [Kubernetes](https://kubernetes.io/) + [Kustomize](https://kustomize.io/) | Deployment manifests and overlays |
| | [minikube](https://minikube.sigs.k8s.io/) | Local cluster |
| | [Tilt](https://tilt.dev/) | Kubernetes development inner loop |
| <img src="https://cdn.simpleicons.org/openapiinitiative/6BA539" width="18"> | [OpenAPI 3.1](https://spec.openapis.org/oas/latest.html) + [Swagger UI](https://swagger.io/tools/swagger-ui/) | Hand-written contract, served from the binary |
| <img src="https://cdn.simpleicons.org/githubactions/2088FF" width="18"> | [GitHub Actions](https://github.com/features/actions) | CI |

**What the backend deliberately does not use**: no web framework, no ORM, no
dependency-injection container, no configuration library, no UUID library, no
assertion library. Go 1.22 gave `net/http` method-aware routing with path
variables, which closed the gap that made frameworks the default. The result is
**five direct runtime dependencies**, each with a reason no standard-library
alternative exists.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  client/  — React, Hexagonal Architecture                   │
│                                                             │
│  Presentation ─▶ Application ─▶ Domain ◀─ Infrastructure    │
│                                              │              │
└──────────────────────────────────────────────┼──────────────┘
                                               │ HTTP/JSON
┌──────────────────────────────────────────────▼──────────────┐
│  api/  — Go, Layered Architecture                           │
│                                                             │
│  Router ─▶ Middleware ─▶ Handler ─▶ Service ─▶ Repository   │
│                                                    │        │
└────────────────────────────────────────────────────┼────────┘
                                                     ▼
                                              PostgreSQL 18
```

### Frontend: Hexagonal Architecture

Also called Ports and Adapters. One rule governs everything: **dependencies
point inward**. The domain knows nothing about React, HTTP, or the database.

```
src/
├── domain/              Business rules. Zero imports from other layers.
│   ├── pokemon/         Entities (Pokemon, Species), value objects
│   │                    (PokemonType, Stats, Sprites, PhysicalMeasurement),
│   │                    and a domain service (TypeEffectivenessService).
│   └── errors/          Domain errors.
│
├── application/         Use cases and the ports they depend on.
│   ├── ports/           Interfaces: PokemonRepository, Logger, CacheService.
│   ├── use-cases/       ListPokemon, GetPokemonById, SearchPokemon.
│   └── dto/             What use cases return — never domain entities.
│
├── infrastructure/      Implementations of the ports.
│   ├── repositories/    PokeAPIRepository and PokedexAPIRepository:
│   │                    two adapters, one port.
│   ├── http/            fetch wrapper with timeouts and problem+json decoding.
│   ├── mappers/         Wire types → domain entities.
│   ├── cache/           LocalStorageCacheService.
│   └── logger/          ConsoleLogger.
│
├── presentation/        React. Components, pages, hooks, routes, stores.
│                        Calls use cases through useDI(). No business logic.
│
└── di/                  Composition root: createContainer() wires everything.
```

**Why the layering pays for itself.** The `PokemonRepository` port declares four
methods. `PokeAPIRepository` satisfies it by talking to a public API;
`PokedexAPIRepository` satisfies it by talking to the Go backend. Both exist in
the codebase simultaneously, and `VITE_DATA_SOURCE` decides which one
`createContainer()` builds. Nothing above the infrastructure layer knows or
cares — which is why swapping data sources broke no tests.

**The domain service that stayed put.** `TypeEffectivenessService` holds the
complete 18×18 type-effectiveness chart as pure domain logic. The API exposes
the same data at `/api/v1/types/{name}/matchups`, but the client keeps
calculating it locally: moving that computation server-side would leave the
domain layer with data and no behaviour. The duplication is deliberate, and a
[test in the backend](api/internal/seeder/typechart_test.go) parses both
representations and fails if they ever disagree.

### Backend: layered, with interfaces owned by consumers

```
api/
├── cmd/
│   ├── server/          Entry point and composition root.
│   ├── seeder/          CLI that populates the database from PokeAPI.
│   └── tlscert/         Self-signed certificate generator for local HTTPS.
├── internal/
│   ├── config/          Environment configuration with full validation.
│   ├── router/          Flat route registration on one ServeMux.
│   ├── middleware/      Recovery, request ID, CORS, security headers,
│   │                    logging, rate limiting, HTTP caching, timeouts.
│   ├── handler/         Parse, delegate, serialise. No business rules.
│   ├── service/         Business logic — and the repository interfaces.
│   ├── repository/      PostgreSQL implementations + sqlc output.
│   ├── redisstore/      Redis implementations of shared state.
│   ├── model/           Domain types, sentinel errors, pagination.
│   ├── httperr/         RFC 9457 error envelope, shared by all layers.
│   └── reqctx/          Request-scoped context values.
├── migrations/          Versioned SQL (golang-migrate).
├── sqlc/                Query definitions and codegen config.
└── deploy/              Compose files and Kubernetes manifests.
```

Repository interfaces are declared in `internal/service`, not next to their
implementations. That is the Go convention — the consumer owns the abstraction —
and it means `internal/repository/postgres` satisfies them without importing the
service, so changing storage never touches business logic.

---

## Quick start

Requires [Go 1.26+](https://go.dev/dl/), [Node 20+](https://nodejs.org/),
[pnpm](https://pnpm.io/), and [Docker](https://www.docker.com/).

```bash
# 1. Database
cd api
make db-up                 # PostgreSQL 18 in Docker
make tools                 # sqlc, golang-migrate, golangci-lint, air
make migrate-up            # apply the schema

# 2. Data — fetches the full national dex from PokeAPI (~35s)
make seed

# 3. API
make run                   # http://localhost:8080
                           # reference at http://localhost:8080/docs

# 4. Client
cd ../client
pnpm install
cp .env.example .env.local
pnpm dev                   # http://localhost:5173
```

To run the client against the public PokeAPI instead, set
`VITE_DATA_SOURCE=pokeapi` in `client/.env.local`. Both adapters are always
available.

### Everything in containers

```bash
cd api
make docker-dev            # Postgres + migrations + API with hot reload
```

### On Kubernetes

```bash
minikube start
eval $(minikube docker-env)
docker build -t pokedex-api:dev api/

kubectl apply -k api/deploy/k8s/overlays/dev
kubectl -n pokedex-dev port-forward svc/pokedex-api 8080:80
```

Or use [Tilt](https://tilt.dev/) for the fast loop — it rebuilds the binary and
syncs it into the running pod without an image rebuild:

```bash
cd api && tilt up          # UI at http://localhost:10350
```

---

## API

Interactive reference at **`/docs`**, served from assets embedded in the binary
— no CDN, works offline. The [OpenAPI 3.1 document](api/docs/openapi.yaml) is
hand-written, and a [test](api/internal/router/router_test.go) fails if any
route is served but undocumented, or documented but unserved.

```
GET  /health                          Liveness  — process only, never the DB
GET  /ready                           Readiness — checks dependencies
GET  /api/v1/pokemon?page=1&limit=20  Paginated list (optional ?type=fire)
GET  /api/v1/pokemon/search?q=pika    Search by name substring
GET  /api/v1/pokemon/{id}             Full detail, species embedded
GET  /api/v1/pokemon/{id}/species     Species detail
GET  /api/v1/types                    All 18 elemental types
GET  /api/v1/types/{name}/matchups    Weaknesses, resistances, immunities
GET  /docs                            Swagger UI
```

Errors use [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457) problem details,
so a client has one documented error shape regardless of what failed:

```json
{
  "type": "https://github.com/iandresfv/clean-arch-pokedex/errors/not-found",
  "title": "Pokemon not found",
  "status": 404,
  "detail": "no pokemon exists with the requested identifier",
  "instance": "/api/v1/pokemon/9999",
  "requestId": "8f14e45fceea167a"
}
```

`requestId` appears on every log line for that request, so a user-reported
failure can be traced without guessing.

---

## Decisions worth explaining

### The database, not the ORM, does the work

The catalogue is ~1300 rows read constantly and written once. That shapes
everything:

**Substring search uses a trigram index.** A B-tree is ordered, so it can seek
`LIKE 'pika%'` but not `LIKE '%kach%'` — there is nowhere to jump to when the
match may start anywhere. A [GIN trigram
index](https://www.postgresql.org/docs/current/pgtrgm.html) decomposes each name
into three-character pieces, turning a substring match into an index lookup.

Then honesty intervened: at this table size, **the planner prefers a sequential
scan anyway**. 728 kB fits entirely in `shared_buffers`, and scanning it costs
less than walking the GIN index. So the [regression
test](api/internal/repository/postgres/queryplan_test.go) asserts what remains
true as the table grows — the index exists, is usable, and returns identical
rows — rather than asserting a plan the planner has good reason not to choose.

**Foreign keys are indexed by hand.** PostgreSQL creates indexes automatically
for `PRIMARY KEY` and `UNIQUE`, but never for a `FOREIGN KEY`. Without them,
filtering by type scans the whole join table.

**Ordered listing has no `Sort` node.** `ORDER BY id LIMIT 20` walks the primary
key index and stops after twenty rows — four buffer reads. A `Sort` in that plan
would mean reading and sorting 1300 rows to return 20, and a test fails if one
ever appears.

**Rows are aggregated in SQL.** A dual-type Pokémon joined to its types returns
two rows; twenty per page becomes forty, carrying twenty duplicate sprite URLs.
`array_agg` collapses them so one Pokémon is always one row, and pgx maps the
resulting `text[]` straight to `[]string`.

**`ANALYZE` runs after seeding.** Until the planner's statistics are refreshed
they still describe an empty table, so a freshly seeded database picks
sequential scans for everything — slower than an idle one, for reasons invisible
in the code.

### Redis is a correctness fix, not a cache

The rate limiter keeps per-IP counters in a Go map. Correct with one instance.
With more, a Kubernetes Service spreads a client's connections across pods, each
pod counts only what it receives, and the limit silently multiplies.

Measured on minikube, 30 requests against a limit of 20/min with a burst of 10:

| Setup | Allowed | |
|---|---|---|
| 1 replica, in-process | **10** | the configured allowance |
| 3 replicas, in-process | **24** | the defect — roughly 3× |
| 3 replicas, Redis | **10** | fixed, identical to one replica |

Redis is not in front of PostgreSQL. Caching 1300 immutable rows that Postgres
already serves from memory in under a millisecond would add a network hop to
save a network hop. What actually reduces latency here is
[HTTP caching](api/internal/middleware/caching.go): a weak `ETag` folded with a
dataset version, so a conditional request returns `304` with **zero bytes** of
body — and a seeder run invalidates every cached response by incrementing one
number.

The token bucket lives in a [Lua script](api/internal/redisstore/ratelimit.go)
because the read-modify-write must be atomic; two replicas interleaving a read
and a write both see the same value and both allow the request. It **fails
open**: if Redis is unreachable, traffic is allowed and the failure is logged.
Rejecting everything because a limiter's store is down turns a degraded
dependency into an outage.

### Middleware order is load-bearing

```
Recovery → RequestID → CORS → SecurityHeaders → Logging → RateLimit → HTTPCache → Timeout → Router
```

Every position earns its place, but one matters more than the rest: **CORS must
sit outside the error paths**. With CORS applied innermost, a `500` from
Recovery or a `429` from the limiter arrives without `Access-Control-Allow-Origin`,
and the browser reports a CORS failure that completely masks the real error.
There is a [test](api/internal/middleware/middleware_test.go) that panics on
purpose and asserts the CORS header survives.

### The probes exist for the orchestrator

`/health` reports on the process and deliberately never touches the database: if
PostgreSQL is down, restarting the API fixes nothing, and a liveness probe that
fails on a database blip turns it into a cluster-wide restart storm. `/ready`
does check the database, so an unhealthy instance leaves the load balancer
without being killed.

`terminationGracePeriodSeconds: 30` sits comfortably above the server's
10-second shutdown deadline. And `preStop: sleep 5` exists because Kubernetes
sends `SIGTERM` and removes the pod from Service endpoints *concurrently* —
without the pause, a pod stops accepting connections while traffic is still
being routed to it, producing a burst of failed requests on every deployment.

### Configuration fails fast, and reports everything at once

`config.Load()` accumulates every problem rather than returning the first:

```
loading configuration: SERVER_PORT must be an integer, got "abc"
LOG_LEVEL must be one of debug, info, warn, error, got "verbose"
DB_SSLMODE "maybe" is not a valid libpq sslmode
```

Production tightens the rules further: `DB_PASSWORD` loses its development
default and becomes mandatory, and loopback CORS origins are rejected outright —
almost always a variable someone forgot to set for the environment.

---

## Testing

```bash
cd api
make test           # unit + integration, race detector on
make test-cover     # with an HTML coverage report
make lint           # golangci-lint

cd ../client
pnpm test:unit      # Vitest
pnpm test:e2e       # Playwright
```

**64 Go test functions** across 8 packages, and **203 client tests**. Beyond
counting them, a few are worth calling out because they assert things that are
easy to get wrong and silent when broken:

- **Query plans** — `EXPLAIN ANALYZE` parsed and asserted, because a dropped
  index breaks nothing visible until traffic grows.
- **LIKE metacharacter escaping** — searching for `%` must match nothing, not
  everything.
- **Type chart agreement** — the SQL migration and the client's domain service
  are parsed and compared cell by cell.
- **Route/spec agreement** — every served route is documented and vice versa.
- **Rate limiter refill** — uses [`testing/synctest`](https://pkg.go.dev/testing/synctest)
  so a limiter defined in minutes can be tested in milliseconds of virtual time.
- **Constraint enforcement** — invalid data is rejected by the database itself,
  not merely by the service layer.

Integration tests run against a real PostgreSQL. A mocked database proves the Go
code calls the repository correctly; it cannot prove the SQL is valid, that
`array_agg` preserves slot order, or that a `CHECK` constraint fires.

---

## Production image

```
22.9 MB   gcr.io/distroless/static  •  non-root  •  read-only root filesystem
```

Multi-stage build: the Go toolchain and source live only in the build stage.
With `CGO_ENABLED=0` the binary is statically linked and needs no shell, no
package manager, and no libc — so the runtime image has none. Alpine's value is
having a shell, which is precisely what should not be in a production runtime.

That creates one problem worth mentioning: a Docker `HEALTHCHECK` normally shells
out to `curl`, and there is neither. The binary therefore probes itself with
`/app/server -health`.

---

## Repository layout

```
clean-arch-pokedex/
├── client/                 React frontend
│   ├── src/
│   │   ├── domain/         Business logic, zero dependencies
│   │   ├── application/    Use cases and ports
│   │   ├── infrastructure/ Adapters: two repositories, one port
│   │   ├── presentation/   React UI
│   │   └── di/             Composition root
│   └── tests/              Unit, integration, E2E
└── api/                    Go backend
    ├── cmd/                server, seeder, tlscert
    ├── internal/           Application code
    ├── migrations/         Versioned SQL
    ├── sqlc/               Query definitions
    ├── docs/               OpenAPI 3.1 specification
    └── deploy/             Compose and Kubernetes manifests
```

---

## Credits and licence

Data sourced from [PokeAPI](https://pokeapi.co/) and stored locally. Sprite
images remain hosted by PokeAPI: this project replaces it as a source of
*data*, not of binary assets. Pokémon and Pokémon character names are trademarks
of Nintendo.

Released under the [MIT License](LICENSE).
