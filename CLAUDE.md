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

- stdlib `net/http` with Go 1.25 ServeMux, `log/slog` for logging, `sqlc` for type-safe SQL
- Interfaces defined where used (in `service` package, not `repository`)
- Table-driven tests, context as first parameter, explicit error handling

## Collaboration Mode (Backend)

The Go API is developed hands-on by the engineer. Claude provides guidance, teaching, and explanations but does NOT write code unless explicitly asked. Discuss the "why" behind architectural decisions (graceful shutdown, middleware ordering, TLS, etc.) rather than providing ready-made implementations.

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
