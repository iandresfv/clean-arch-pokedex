# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Monorepo with two components:

- **`client/`** — Frontend (React 19 + TypeScript 5.9), Hexagonal/Clean Architecture
- **`api/`** — Backend (Go 1.25), Layered Architecture (in progress)

The frontend consumes the [PokeAPI](https://pokeapi.co/) and demonstrates enterprise-grade Clean Architecture patterns.

## Common Commands

All frontend commands run from `client/`:

```bash
cd client
pnpm install          # Install dependencies
pnpm dev              # Dev server at localhost:5173
pnpm build            # Type-check + production build
pnpm type-check       # TypeScript strict validation (tsc --noEmit)
pnpm lint             # ESLint check
pnpm lint:fix         # Auto-fix lint errors
pnpm format           # Prettier format
pnpm fix              # Prettier + ESLint fix combined
pnpm test             # Run all tests (Vitest, watch mode)
pnpm test:unit        # Single run with coverage
pnpm test:coverage    # Same as test:unit
pnpm test:e2e         # Playwright E2E tests
pnpm test:e2e:ui      # Playwright interactive UI
```

Run a single test file: `cd client && pnpm vitest run tests/unit/domain/Pokemon.test.ts`

Pre-commit hook (Husky) runs `lint-staged` automatically on `.ts`, `.tsx`, `.json`, `.md`, `.css` files.

Backend commands run from `api/` (Go 1.26, module `github.com/iandresfv/clean-arch-pokedex/api`):

```bash
cd api
make run              # go run ./cmd/server/ (listens on SERVER_HOST:SERVER_PORT, default localhost:8080)
make build            # go build -o bin/server ./cmd/server/
go test ./...         # Run all Go tests
go test -race ./internal/service/...        # Single package with race detector
go test -run TestName ./internal/service/   # Single test by name
```

The Makefile currently defines only `run` and `build`; the fuller command set (test, lint, migrate, sqlc, docker) in the roadmap is planned, not yet wired.

## Architecture (Frontend)

Hexagonal Architecture with strict inward dependency rule:

```
Presentation → Application → Domain ← Infrastructure
```

### Layer Boundaries (NEVER violate)

- **Domain** (`src/domain/`) — Pure business logic, zero imports from other layers. Entities with factory pattern (`Pokemon.create()`), Value Objects (`PokemonType`, `Stats`, `PhysicalMeasurement`, `Sprites`), Domain Services (`TypeEffectivenessService`), Domain Errors.
- **Application** (`src/application/`) — Use cases, ports (interfaces), DTOs, pagination/search types. Use cases accept port interfaces via constructor injection and return DTOs (not domain entities).
- **Infrastructure** (`src/infrastructure/`) — Implements application ports. `PokeAPIRepository` implements `PokemonRepository`, `ConsoleLogger` implements `Logger`, `LocalStorageCacheService` implements `CacheService`. Contains API response type definitions and mappers (PokeAPI → Domain).
- **Presentation** (`src/presentation/`) — React components, pages, hooks, layout, routes, Zustand stores. Calls use cases via DI hook (`useDI()`). Never contains business logic.
- **DI** (`src/di/`) — Composition root. `createContainer()` wires all dependencies. Provided via React Context, consumed via `useDI()` hook.

### Key Patterns

- **Use cases**: Interface (port) + Impl class, injected with repository and logger
- **TanStack Query + Use Cases**: Presentation hooks call use cases via `useDI()`, TanStack Query handles caching/state
- **Zustand**: Client-only state (favorites with localStorage persistence), separate from server state
- **Shadcn/ui**: Components are owned copies (not npm packages), built on Radix UI primitives
- **Path aliases**: `@/domain/*`, `@/application/*`, `@/infrastructure/*`, `@/presentation/*`

## Architecture (Backend — Go)

Layered: `HTTP Request → Router → Middleware → Handler → Service → Repository → PostgreSQL`

- stdlib `net/http` with method-aware ServeMux patterns (Go 1.22+), `log/slog` for logging, `sqlc` for type-safe SQL, PostgreSQL 18 via `pgx/v5`
- Interfaces defined where used (in `service` package, not `repository`)
- Table-driven tests, context as first parameter, explicit error handling
- Config is struct-based from env vars (`internal/config`) — no web framework, ORM, DI container, or config library. Permitted runtime deps: `pgx/v5`, `golang-migrate`, `go-redis` (Phase 12), and `x/crypto` + `golang-jwt` only if the optional auth phase is built
- **`.dev-notes/docs/API_ROADMAP.md` is the source of truth for the backend** (kept outside version control — `.dev-notes/` is gitignored): 58 stages across 13 phases, each stage = one atomic commit, with a Design Decisions table (D1–D15) recording the rationale for every cross-cutting choice. The progress table there tracks what's Done vs Pending. Most `internal/*` packages are currently `doc.go` placeholders — consult the roadmap before implementing, and update its progress table when a stage lands. `.dev-notes/docs/API_ROADMAP.es.md` is a Spanish translation of the same document — keep it in sync when the English one changes.
- Published documentation is deliberately absent while the backend is in progress: the finished project will be documented in `README.md` only. Do not create `docs/` or other committed reference docs.
- Git Flow with release-per-phase (`v2.x.0` tags); feature branches named `feature/api-<desc>`

## Collaboration Mode (Backend)

Hybrid collaboration: implementation is split between engineer and Claude depending on the stage. All code follows senior-level quality standards. Each stage is committed atomically, merged to develop via Git Flow, and tracked in the roadmap progress table.

## Testing

Vitest (unit + integration) with jsdom environment. E2E tests (Playwright) are excluded from Vitest runner.

- Test setup: `tests/setup.ts` (Testing Library matchers + cleanup)
- Mock pattern: `{ [K in keyof Interface]: Mock }` for type-safe mocks
- AAA pattern: Arrange → Act → Assert
- Coverage targets: Domain 100%, Application 90%+, Infrastructure 80%+, Presentation 70%+

## Code Conventions

- **TypeScript strict mode**: No `any`, use `unknown`. `erasableSyntaxOnly` enabled (no parameter properties). Consistent type imports (`import type`).
- **Import order** (enforced by eslint-plugin-simple-import-sort): React → external packages → parent imports → sibling imports → CSS → side effects
- **File naming**: PascalCase for components/classes/entities, camelCase for functions/hooks, kebab-case for file names
- **Comments**: Default to NO comments. Self-documenting names. Only comment complex algorithms or architectural decisions.
- **Prettier**: 100 char width, single quotes, trailing commas (es5), 2-space indent
- **Git commits**: Conventional Commits — single line only: `type(scope): subject`. No body, no co-authored-by, no multi-line. Types: feat, fix, docs, test, refactor, perf, chore. Scopes: domain, application, infrastructure, presentation, ui, config, deps, api.
