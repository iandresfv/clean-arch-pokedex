# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Application layer: Use cases, ports, and DTOs
- Infrastructure layer: PokeAPI repository with caching
- Presentation layer: Pokemon list, detail, and search pages

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

[unreleased]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/iandresfv/clean-arch-pokedex/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/iandresfv/clean-arch-pokedex/releases/tag/v0.1.0
