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

### Frontend
- **React 19.2** - UI library
- **Vite 7** - Build tool (fast HMR)
- **TypeScript 5.9** - Type safety
- **TanStack Query 5** - Server state management
- **Zustand 5** - Client state management
- **Shadcn/ui** - Accessible UI components
- **TailwindCSS 4** - Utility-first styling
- **React Hook Form + Zod** - Forms & validation
- **React Router 7** - Client-side routing

### Testing
- **Vitest** - Unit & integration tests
- **React Testing Library** - Component tests
- **Playwright** - End-to-end tests

### Development Tools
- **Docker + Docker Compose** - Containerization
- **pnpm** - Fast, efficient package manager
- **ESLint + Prettier** - Code quality
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
├── client/                     # Frontend application
│   ├── src/
│   │   ├── domain/            # Business logic (zero dependencies)
│   │   ├── application/       # Use cases & ports
│   │   ├── infrastructure/    # External integrations
│   │   ├── presentation/      # React UI
│   │   └── di/                # Dependency injection
│   └── tests/                 # Test suites
└── api/                       # Backend (future - Golang)
```

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
