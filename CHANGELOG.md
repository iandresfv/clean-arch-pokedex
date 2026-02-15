# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- Search functionality with debounce
- Favorites feature with localStorage persistence
- UI polish and responsive design improvements
- E2E tests with Playwright

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
- Comprehensive documentation (README, .cursorrules, Git workflow)
- TailwindCSS 4 + Shadcn/ui component library setup
- Monorepo structure prepared for future Golang API

### Infrastructure

- Vite 7 + React 19 + TypeScript 5.9 stack
- ESLint strict type checking with custom import sorting
- Prettier with 100-character line width (industry standard)
- EditorConfig for cross-editor consistency
- Docker multi-stage build ready for production
- Hot Module Replacement (HMR) working in Docker
- VSCode/Cursor settings for optimal DX

### Documentation

- Clean Architecture guidelines and principles
- Pedagogical approach for learning and understanding
- Git Flow strategy with real-world examples
- Commit conventions (Conventional Commits)
- CHANGELOG maintenance guide

[unreleased]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/iandresfv/clean-arch-pokedex/releases/tag/v0.1.0
