# Clean Architecture Pokédex

Production-ready Pokédex application built with **Clean Architecture** (Hexagonal Architecture), featuring comprehensive testing strategy and enterprise software engineering practices.

[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-blue)](https://www.typescriptlang.org/)
[![React](https://img.shields.io/badge/React-19.2-61dafb)](https://reactjs.org/)
[![Vite](https://img.shields.io/badge/Vite-7.2-646cff)](https://vitejs.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🏗️ Architecture Highlights

- **Clean Architecture** (Hexagonal Architecture) with strict layer separation
- **Domain-Driven Design** principles and bounded contexts
- **SOLID** principles applied throughout the codebase
- **Dependency Inversion** - Domain layer has zero external dependencies
- **Comprehensive test coverage** following the testing pyramid
- **Type-safe patterns** with TypeScript strict mode

## 🏛️ Architecture

```
┌─────────────────────────────────────────────────┐
│           PRESENTATION LAYER                    │
│     React Components + TanStack Query           │
└────────────────┬────────────────────────────────┘
                 │ calls use cases
┌────────────────▼────────────────────────────────┐
│           APPLICATION LAYER                     │
│        Use Cases + Ports (Interfaces)           │
└────────────────┬────────────────────────────────┘
                 │ uses domain
┌────────────────▼────────────────────────────────┐
│             DOMAIN LAYER                        │
│      Entities + Value Objects + Rules           │
└────────────────▲────────────────────────────────┘
                 │ implements ports
┌────────────────┴────────────────────────────────┐
│         INFRASTRUCTURE LAYER                    │
│      API Clients + Repositories + Cache         │
└─────────────────────────────────────────────────┘
```

**Key Principle**: Dependencies point inward. Domain has zero external dependencies.

## 🛠️ Tech Stack

### Frontend (`client/`)
- **React 19.2** - UI library
- **Vite 7** - Build tool (fast HMR)
- **TypeScript 5.9** - Type safety
- **TanStack Query 5** - Server state management
- **Zustand 5** - Client state management
- **Shadcn/ui** - Accessible UI components
- **TailwindCSS 4** - Utility-first styling
- **React Hook Form + Zod** - Forms & validation
- **React Router 7** - Client-side routing

### Backend (`api/`) — In Progress
- **Go 1.25** - Language
- **stdlib `net/http`** - HTTP server (no framework)
- **PostgreSQL** - Database
- **sqlc** - Type-safe SQL-to-Go code generation
- **golang-migrate** - Database migrations
- **`log/slog`** - Structured logging (stdlib)

### Testing
- **Vitest** - Unit & integration tests (frontend)
- **React Testing Library** - Component tests
- **Playwright** - End-to-end tests
- **Go `testing`** - Table-driven tests (backend)

### Development Tools
- **Docker + Docker Compose** - Containerization
- **pnpm** - Fast, efficient package manager
- **ESLint + Prettier** - Code quality (frontend)
- **golangci-lint** - Multi-linter aggregator (backend)
- **TypeScript ESLint (Strict)** - Type-safe linting

## 🚀 Quick Start

### Prerequisites

- Node.js 22+ (or Docker)
- pnpm 9+

### Installation

```bash
# Clone repository
git clone https://github.com/iandresfv/clean-arch-pokedex.git
cd clean-arch-pokedex

# Install dependencies
cd client
pnpm install

# Run development server
pnpm dev

# Open browser at http://localhost:5173
```

### Available Scripts

```bash
# Development
pnpm dev          # Start dev server with HMR
pnpm build        # Build for production
pnpm preview      # Preview production build

# Testing
pnpm test         # Run all tests
pnpm test:watch   # Run tests in watch mode
pnpm test:coverage # Generate coverage report

# Code Quality
pnpm lint         # Run ESLint
pnpm lint:fix     # Fix linting errors
pnpm format       # Format code with Prettier
pnpm format:check # Check formatting
pnpm type-check   # TypeScript type checking
```

## 🐳 Docker

Run the application with Docker Compose:

```bash
# Start development server
docker-compose up

# Stop services
docker-compose down
```

Application will be available at `http://localhost:5173`

## 📁 Project Structure

```
clean-arch-pokedex/
├── client/                     # Frontend application (React + TypeScript)
│   ├── src/
│   │   ├── domain/            # Business logic (zero dependencies)
│   │   ├── application/       # Use cases & ports
│   │   ├── infrastructure/    # External integrations
│   │   ├── presentation/      # React UI
│   │   └── di/                # Dependency injection
│   └── tests/                 # Test suites
└── api/                       # Backend API (Go 1.26)
    ├── cmd/server/            # Entry point
    ├── internal/              # Application code
    │   ├── handler/           # HTTP handlers
    │   ├── service/           # Business logic
    │   ├── repository/        # Data access (sqlc + PostgreSQL)
    │   ├── model/             # Domain types
    │   └── middleware/        # CORS, logging, auth
    └── migrations/            # Database migrations
```

## 🤖 AI-Assisted Development

This project uses a structured **AI-assisted development methodology** with different approaches per component:

| Component | AI Role | Approach |
|-----------|---------|----------|
| **`client/`** (Frontend) | 100% AI-assisted | Architecture and standards defined by the engineer; implementation executed by AI |
| **`api/`** (Backend) | Hybrid | Engineer writes core business logic; AI assists with configuration, scaffolding, and boilerplate |

Both approaches are guided by the same `CLAUDE.md` engineering standards — ensuring consistent quality regardless of who writes the code.

### The Approach

The frontend was developed through collaboration with AI coding assistants (Claude Code), guided by a comprehensive `CLAUDE.md` file that defines:

- **Architectural constraints**: Clean Architecture layer boundaries, dependency direction enforcement
- **Code quality standards**: SOLID principles, self-documenting code, minimal comments policy
- **Documentation mandate**: Every library usage verified against official docs for the exact version in `package.json`
- **Technology conventions**: File naming, import ordering, error handling patterns, testing standards
- **Forbidden practices**: Direct API calls in components, `any` types, domain-infrastructure coupling

### Why This Matters

AI-assisted development is not "vibe coding." The quality of the output is directly proportional to the quality of the engineering context provided. This project demonstrates that:

1. **Architecture must be defined by the engineer** — The hexagonal architecture, layer boundaries, dependency rules, and DDD patterns were design decisions made before any code was written. AI executed the vision; it didn't invent it.

2. **AI rules files are engineering artifacts** — The `CLAUDE.md` file is not a prompt template. It's a technical specification that encodes years of software engineering experience: Clean Architecture principles, SOLID, Clean Code philosophy, and modern frontend best practices.

3. **Documentation-first approach prevents hallucinations** — By mandating official documentation verification for every library (with version-specific notes for TanStack Query v5, React Router v7, TailwindCSS v4), the AI consistently produces correct, up-to-date implementations instead of guessing from outdated training data.

4. **The result speaks for itself** — 192+ unit tests, 17 E2E tests, strict TypeScript, proper domain modeling with Value Objects and Entities, graceful error degradation, multi-layer caching strategy — all following consistent patterns across the entire codebase.

### What the Engineer Defined

- Hexagonal Architecture as the frontend foundation
- Layered Architecture (handler → service → repository) for the Go backend
- Domain model: Entities (Pokemon, Species), Value Objects (PokemonType, Stats, PhysicalMeasurement, Sprites), Domain Services (TypeEffectivenessService)
- Ports & Adapters pattern for infrastructure isolation
- Manual dependency injection strategy using React Context
- Data fetching strategy (N+1 with Promise.all parallelization + multi-layer cache)
- Testing pyramid and coverage targets per layer
- Git Flow workflow with Conventional Commits
- Backend tech stack decisions: Go 1.25 stdlib, PostgreSQL, sqlc

### What AI Executed (Frontend)

- Code implementation following the defined architecture and conventions
- Test writing following the established patterns (AAA, type-safe mocks)
- API mapper implementation (PokeAPI response → domain entities)
- Component development following Shadcn/ui patterns
- Documentation generation based on actual source code analysis

### What AI Assists With (Backend)

- Project scaffolding and boilerplate (Docker, Makefile, config loading)
- Code review and idiomatic Go suggestions
- Test scaffolding and table-driven test structure
- Documentation and architectural reference materials

> **For engineering leaders**: This approach mirrors how a tech lead or architect works with a development team — defining the "what" and "why" while delegating the "how" to capable implementers. The `CLAUDE.md` file is essentially an architectural decision record (ADR) that also serves as a living style guide.

## 🧪 Testing Strategy

```
     🔺 E2E (Playwright) - 10%
    🔺🔺 Integration (RTL) - 30%
   🔺🔺🔺 Unit (Vitest) - 60%
```

### Coverage Goals

- Domain Layer: **100%** (pure logic, easy to test)
- Application Layer: **90%+** (use cases with mocked repos)
- Infrastructure Layer: **80%+** (integration tests)
- Presentation Layer: **70%+** (component behavior)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 📚 Technical References

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) by Robert C. Martin
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/) by Alistair Cockburn
- [Domain-Driven Design](https://www.domainlanguage.com/ddd/) by Eric Evans
- [PokéAPI](https://pokeapi.co/) - RESTful Pokémon API

---

<div align="center">

**Built with Clean Architecture principles**

</div>
