# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [2.11.0] - 2026-08-17

### Added
- Redis-backed rate limiter using an atomic Lua token bucket, so the configured
  limit holds across replicas instead of multiplying by the replica count.
- Redis service in both the Compose and Kubernetes stacks, with LRU eviction and
  no persistence: the state is rate-limiting counters with a 10-minute TTL.
- Load-test script documenting the measured behaviour — 3 replicas admitted 24
  requests against a limit of 10 before the fix, and exactly 10 after.

## [2.10.0] - 2026-08-17

### Added
- Kubernetes manifests: Deployment with liveness, readiness and startup probes,
  PostgreSQL StatefulSet, migration Job, ConfigMap and Secret.
- `preStop` sleep and a 30s termination grace period so rolling deployments drop
  no in-flight requests.
- `GOMEMLIMIT` and `GOMAXPROCS` derived from the container's own resource limits.
- Kustomize overlays for development and production, plus HorizontalPodAutoscaler
  and PodDisruptionBudget.
- Tiltfile with `live_update`, which syncs a rebuilt binary into the running pod
  instead of rebuilding the image.

## [2.9.0] - 2026-08-17

### Added
- Hand-written OpenAPI 3.1 specification for every endpoint.
- Swagger UI served from assets embedded in the binary — no CDN, works offline.
- Test asserting every served route is documented and every documented route is
  served.

## [2.8.0] - 2026-08-17

### Added
- Security headers middleware (CSP, HSTS when TLS is on, nosniff, frame denial).
- HTTP caching with weak `ETag` and `Cache-Control`; a conditional request
  returns `304` with an empty body.
- IP-based rate limiting with trusted-proxy resolution of `X-Forwarded-For`.
- Optional TLS, which also enables HTTP/2 through ALPN negotiation.
- `tlscert` command for generating development certificates.

## [2.7.0] - 2026-08-17

### Added
- Multi-stage Dockerfile producing a 22.9 MB distroless image running as
  non-root with a read-only root filesystem.
- `-health` and `-version` flags so the shell-less image can still be probed.
- Development Dockerfile with Air hot reload, and a full Compose stack that runs
  migrations to completion before starting the API.

## [2.6.0] - 2026-08-17

### Added
- `PokedexAPIRepository`, a second adapter implementing the existing
  `PokemonRepository` port.
- `VITE_DATA_SOURCE` selects the adapter at the composition root; both remain
  available.
- HTTP client with timeouts and RFC 9457 error decoding.

### Changed
- The client now reads from this project's Go API by default. Listing a page
  costs one request instead of twenty-one; search is one request instead of
  downloading all 1302 names.
- API list responses carry complete entities, because the client's port returns
  domain entities and a slimmer shape would reintroduce the N+1.

### Notes
- All 203 client tests passed without modification, and no file in the domain,
  application or presentation layers changed.

## [2.5.0] - 2026-08-17

### Added
- Service and model unit tests, handler tests with `httptest`, and repository
  integration tests against a real PostgreSQL.
- Query plan regression tests parsing `EXPLAIN ANALYZE` output.

## [2.4.0] - 2026-08-17

### Added
- CLI seeder with bounded concurrency, exponential backoff honouring
  `Retry-After`, and a single all-or-nothing transaction.
- `ANALYZE` after bulk loading, without which the planner still believes the
  table is empty.
- Test asserting the seeded type chart matches the client's domain service.

## [2.3.0] - 2026-08-17

### Added
- Middleware: panic recovery, request ID with a context-scoped logger,
  structured request logging, CORS, and per-request timeouts.
- Shared RFC 9457 error envelope so every layer reports failures identically.

### Notes
- CORS sits outside the error paths so a 500 or 429 still carries
  `Access-Control-Allow-Origin`; without that the browser masks the real error.

## [2.2.0] - 2026-08-17

### Added
- Domain models, sentinel errors and validated pagination.
- Service layer holding all business validation.
- HTTP handlers and flat route registration on a single `ServeMux`.
- Strict query parameter validation: unknown parameters are rejected rather than
  ignored.

## [2.1.0] - 2026-08-17

### Added
- PostgreSQL 18 via Docker Compose, mounting the version-specific `PGDATA` path.
- Migrations for the catalogue schema, the 18 types and the 324-entry
  effectiveness matrix.
- GIN trigram index for substring search, plus the foreign-key indexes
  PostgreSQL does not create on its own.
- sqlc-generated queries and a pgx connection pool with startup backoff.
- Repository layer translating driver errors into domain errors.

## [2.0.0] - 2026-08-17

### Added
- Go 1.26 API skeleton: configuration with full validation, HTTP server with
  graceful shutdown, Makefile, golangci-lint configuration and CI workflow.

### Notes
- Configuration reports every problem at once rather than one per restart, and
  production requires credentials that development defaults.

---

## [1.0.0] - 2026-02-14

### Added

- Playwright E2E test suite with 17 tests (list, detail, favorites, navigation)
- Playwright configuration with Chromium and auto web server
- test:e2e and test:e2e:ui npm scripts
- Custom Pokéball SVG favicon replacing Vite default
- HTML meta description and theme-color tags
- React.lazy code splitting for PokemonDetailPage

### Changed

- Vitest excludes E2E test directory to prevent conflicts

---

## [0.5.0] - 2026-02-14

### Added

- Search functionality with debounced input and instant results
- SearchBar component with clear button and keyboard support
- useSearchPokemon hook with TanStack Query integration
- useDebounce generic hook for input delay
- Favorites feature with Zustand store and localStorage persistence
- FavoriteButton component with heart toggle animation
- Favorites counter badge in header navigation
- ErrorBoundary component with retry functionality
- ScrollToTop component for route change scroll restoration
- Footer with PokéAPI attribution
- Fade-in page transitions via CSS animations
- Enhanced EmptyState with SearchX icon and custom icon prop
- Enhanced NotFoundPage with CSS Pokéball illustration
- 192 total tests passing across 21 test files

### Changed

- AppLayout now includes ErrorBoundary, ScrollToTop, and Footer
- List and detail pages use fade-in animation on mount

---

## [0.4.0] - 2026-02-14

### Added

- Shadcn/ui initialization with Button, Card, Badge, Skeleton components
- App layout with sticky header and main content area
- React Router configuration with `/`, `/pokemon/:id`, and 404 routes
- TanStack Query provider with DI context integration
- Pokemon List page with responsive card grid and pagination
- Pokemon Detail page with stats chart, sprites, and species info
- TypeBadge component with Pokemon type color mapping
- PokemonCard and PokemonCardSkeleton components
- StatsChart component for base stats visualization with color-coded bars
- PokemonSprites component with sprite gallery
- PokemonDetailHeader with official artwork and metadata
- SpeciesInfo component with flavor text, generation, and habitat
- usePokemonList and usePokemonDetail custom hooks (TanStack Query)
- Pagination component with page controls
- ErrorState component with retry functionality
- cn() utility for TailwindCSS class merging
- 176 total tests passing (70%+ presentation layer coverage)

### Changed

- Replaced Game Boy Advance loading screen with functional app shell
- Updated main.tsx with provider hierarchy (Query + DI + Router)

### Removed

- App.tsx and App.css (replaced by router-based architecture)

---

## [0.3.0] - 2026-02-14

### Added

- PokemonRepository, CacheService, and Logger outbound ports (interfaces)
- ListPokemonUseCase with pagination support
- GetPokemonByIdUseCase with graceful species fetch failure handling
- SearchPokemonUseCase with term normalization (trim + lowercase)
- PokeAPIRepository implementing PokemonRepository port
- PokemonMapper standalone functions for API-to-domain transformation
- LocalStorageCacheService with TTL-based expiry and prefix-scoped keys
- ConsoleLogger implementing Logger port
- DI container with composition root and React context provider
- useDI hook for accessing container in components
- ApplicationError hierarchy (PokemonNotFoundError, RepositoryError)
- PokemonListItemDTO, PokemonDetailDTO, and SpeciesDTO
- SearchCriteria and PaginationParams/PaginatedResult types
- 154 total tests passing (90%+ application, 80%+ infrastructure coverage)

---

## [0.2.0] - 2026-02-14

### Added

- Pokemon entity with factory method and invariant enforcement
- Species entity with generation validation (1-9)
- PokemonType value object with 18 valid types
- Stats value object with non-negative integer validation
- PhysicalMeasurement value object with unit conversion (dm→m, hg→kg)
- Sprites value object with quality selection logic
- TypeEffectivenessService with full 18x18 damage multiplier chart
- DomainError hierarchy with specific error classes
- Vitest configuration with path aliases and test scripts
- Husky + lint-staged pre-commit hooks for code quality
- 100% unit test coverage for entire domain layer (114 tests)

---

## [0.1.0] - 2026-02-08

### Added

- Complete project setup with Clean Architecture (Hexagonal Architecture)
- Professional tooling configuration (ESLint 9 + Prettier 3)
- Docker Compose development environment with HMR
- Pokédex retro loading screen (Game Boy Advance style)
- Comprehensive documentation (README, CLAUDE.md, Git workflow)
- TailwindCSS 4 + Shadcn/ui component library setup
- Monorepo structure prepared for future Golang API

### Infrastructure

- Vite 7 + React 19 + TypeScript 5.9 stack
- ESLint strict type checking with custom import sorting
- Prettier with 100-character line width (industry standard)
- EditorConfig for cross-editor consistency
- Docker multi-stage build ready for production
- Hot Module Replacement (HMR) working in Docker
- VSCode settings for optimal DX

### Documentation

- Clean Architecture guidelines and principles
- Pedagogical approach for learning and understanding
- Git Flow strategy with real-world examples
- Commit conventions (Conventional Commits)
- CHANGELOG maintenance guide

[unreleased]: https://github.com/iandresfv/clean-arch-pokedex/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.5.0...v1.0.0
[0.5.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/iandresfv/clean-arch-pokedex/releases/tag/v0.1.0
