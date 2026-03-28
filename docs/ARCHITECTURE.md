# Architecture Reference — Clean Arch Pokedex

> Comprehensive technical reference covering every layer of the hexagonal architecture,
> code walkthroughs with real project examples, SOLID principles in practice,
> data fetching and caching strategy, and testing approach.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [The Dependency Rule](#2-the-dependency-rule)
3. [Domain Layer](#3-domain-layer)
4. [Application Layer](#4-application-layer)
5. [Infrastructure Layer](#5-infrastructure-layer)
6. [Dependency Injection (DI Container)](#6-dependency-injection-di-container)
7. [Presentation Layer](#7-presentation-layer)
8. [SOLID Principles — Project Examples](#8-solid-principles--project-examples)
9. [Data Fetching & Caching Strategy](#9-data-fetching--caching-strategy)
10. [Testing Strategy](#10-testing-strategy)
11. [Technical Deep Dive (Q&A)](#11-technical-deep-dive-qa)

---

## 1. Architecture Overview

This project implements **Clean Architecture** (also known as **Hexagonal Architecture** or **Ports & Adapters**), an architectural pattern proposed by Robert C. Martin that organizes code in concentric layers where dependencies always point inward — toward the domain.

### Folder Structure

```
client/src/
├── domain/                  ← Innermost layer: entities and business rules
│   ├── errors/              ← Domain-specific errors
│   └── pokemon/             ← Entities, Value Objects, Domain Services
├── application/             ← Use cases and ports (interfaces)
│   ├── dto/                 ← Data Transfer Objects
│   ├── types/               ← Shared types (pagination, search)
│   ├── use-cases/           ← Orchestrated application logic
│   ├── ports/               ← Interfaces (contracts with the outside world)
│   └── errors/              ← Application errors
├── infrastructure/          ← Concrete adapters (API, cache, logger)
│   ├── api/                 ← External API types (PokeAPI)
│   ├── cache/               ← CacheService implementation
│   ├── logger/              ← Logger implementation
│   ├── mappers/             ← API → Domain transformation
│   └── repositories/        ← PokemonRepository implementation
├── di/                      ← Composition Root (dependency injection)
├── presentation/            ← React: components, hooks, routes, stores
│   ├── components/          ← UI components (pokemon/, shared/, ui/)
│   ├── hooks/               ← Custom hooks (TanStack Query wrappers)
│   ├── stores/              ← Zustand stores (favorites)
│   ├── pages/               ← Pages (List, Detail, NotFound)
│   ├── layout/              ← Shared layout (Header, Footer, AppLayout)
│   └── routes/              ← Router configuration
└── lib/                     ← Utilities (cn helper for Tailwind)
```

### Dependency Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    PRESENTATION                          │
│   React Components, Hooks, Pages, Router, Stores        │
│                         │                                │
│                         ▼                                │
│  ┌──────────────────────────────────────────────────┐   │
│  │              DI CONTAINER                         │   │
│  │    Composition Root (dependency wiring)            │   │
│  └──────────────────────────────────────────────────┘   │
│                    │              │                       │
│                    ▼              ▼                       │
│  ┌────────────────────┐  ┌───────────────────────┐      │
│  │    APPLICATION      │  │   INFRASTRUCTURE      │      │
│  │  Use Cases, DTOs,   │  │  PokeAPIRepository,   │      │
│  │  Ports (interfaces) │◄─│  CacheService,        │      │
│  └────────────────────┘  │  ConsoleLogger,        │      │
│           │               │  PokemonMapper        │      │
│           ▼               └───────────────────────┘      │
│  ┌────────────────────┐                                  │
│  │      DOMAIN         │                                 │
│  │  Entities, VOs,     │                                 │
│  │  Domain Services    │                                 │
│  └────────────────────┘                                  │
└─────────────────────────────────────────────────────────┘
```

**Fundamental rule**: Dependency arrows always point inward. Infrastructure depends on Application (implements its interfaces), Application depends on Domain (uses its entities), and Domain has zero external dependencies.

---

## 2. The Dependency Rule

The **Dependency Rule** is the most important principle of Clean Architecture:

> *"Source code dependencies must only point inward, toward higher-level policies."*

### What this means in practice

1. **Domain** does NOT import from application, infrastructure, or presentation
2. **Application** imports from domain, but NOT from infrastructure or presentation
3. **Infrastructure** imports from application (to implement interfaces) and domain (to create entities)
4. **Presentation** imports from application (DTOs, use cases) and domain (types)

### How is inversion achieved?

Through the **Dependency Inversion Principle (DIP)**. The Application layer defines **interfaces** (ports) that the Infrastructure layer implements:

```typescript
// application/ports/PokemonRepository.ts — Application defines the CONTRACT
export interface PokemonRepository {
  getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>>;
  getById(id: number): Promise<Pokemon>;
  getSpeciesById(id: number): Promise<Species>;
  searchByName(name: string, params: PaginationParams): Promise<PaginatedResult<Pokemon>>;
}

// infrastructure/repositories/PokeAPIRepository.ts — Infrastructure IMPLEMENTS the contract
export class PokeAPIRepository implements PokemonRepository {
  async getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    // Concrete implementation using fetch against PokeAPI
  }
  // ...
}
```

The key: **`ListPokemonUseCaseImpl` receives a `PokemonRepository` (interface)**, not a `PokeAPIRepository` (concrete implementation). To switch from PokeAPI to GraphQL or a local database, only the Infrastructure implementation changes — Application and Domain remain untouched.

---

## 3. Domain Layer

The domain layer is the **core** of the application. It contains fundamental business rules that hold true regardless of the UI, database, or framework.

### 3.1 Entities

#### Pokemon Entity

```typescript
// domain/pokemon/Pokemon.ts
export class Pokemon {
  private readonly _id: number;
  private readonly _name: string;
  private readonly _types: readonly PokemonType[];
  private readonly _stats: Stats;
  private readonly _height: PhysicalMeasurement;
  private readonly _weight: PhysicalMeasurement;
  private readonly _sprites: Sprites;
  private readonly _order: number;
  private readonly _baseExperience: number | null;

  // PRIVATE constructor — cannot be instantiated directly
  private constructor(/* ... */) { /* field assignment */ }

  // Static factory method — ONLY entry point for creating a Pokemon
  static create(props: PokemonProps): Pokemon {
    Pokemon.validateId(props.id);
    Pokemon.validateName(props.name);
    Pokemon.validateTypes(props.types);
    Pokemon.validateOrder(props.order);

    const types = props.types.map((t) => new PokemonType(t));
    const stats = new Stats(props.stats);
    const height = new PhysicalMeasurement(props.height.value, props.height.unit);
    const weight = new PhysicalMeasurement(props.weight.value, props.weight.unit);
    const sprites = new Sprites(props.sprites);

    return new Pokemon(
      props.id, props.name.trim(),
      Object.freeze([...types]),  // Runtime immutability
      stats, height, weight, sprites,
      props.order, props.baseExperience
    );
  }
}
```

**Key design decisions:**

1. **Private constructor + Factory Method (`create`)**: Guarantees invalid Pokemon instances cannot exist. All validation runs in `create()` before the constructor is called.
2. **`erasableSyntaxOnly: true` in tsconfig**: No TypeScript parameter properties — fields are assigned manually in the constructor body.
3. **`Object.freeze([...types])`**: Runtime immutability for the types array. Even plain JavaScript cannot mutate it after creation.
4. **Invariant validation**: A Pokemon MUST always have a positive integer ID, non-empty name, 1-2 types, and an integer order. Violations throw `InvalidPokemonEntityError`.

#### Species Entity

```typescript
// domain/pokemon/Species.ts
export class Species {
  private constructor(/* ... */) { /* ... */ }

  static create(props: SpeciesProps): Species {
    Species.validateId(props.id);
    Species.validateName(props.name);
    Species.validateGeneration(props.generation);  // Generation 1-9

    return new Species(/* ... */);
  }

  get isSpecial(): boolean {
    return this._isLegendary || this._isMythical;
  }
}
```

Same pattern: private constructor, factory method, invariant validation. Generation is constrained to 1-9 (current Pokemon generations).

### 3.2 Value Objects

Value Objects are immutable objects defined by their attributes, not by identity. Two VOs with identical values are equal.

#### PokemonType

```typescript
export const VALID_POKEMON_TYPES = [
  'normal', 'fire', 'water', 'electric', 'grass', 'ice',
  'fighting', 'poison', 'ground', 'flying', 'psychic', 'bug',
  'rock', 'ghost', 'dragon', 'dark', 'steel', 'fairy',
] as const;

export type PokemonTypeName = (typeof VALID_POKEMON_TYPES)[number];

export class PokemonType {
  private readonly _value: PokemonTypeName;

  constructor(value: string) {
    const normalizedValue = value.toLowerCase().trim();
    if (!this.isValidType(normalizedValue)) {
      throw new InvalidPokemonTypeError(value);
    }
    this._value = normalizedValue;
  }

  equals(other: PokemonType): boolean {
    return this._value === other._value;
  }
}
```

- `as const` creates a **tuple type** with exactly the 18 official Pokemon types
- `PokemonTypeName` is a **literal union type** derived from the array
- Constructor normalizes input (`"FIRE"` → `"fire"`) and rejects invalid values
- `equals()` implements **value equality** (not reference equality)

#### Stats, PhysicalMeasurement, Sprites

Each Value Object encapsulates domain logic:

- **Stats**: Validates non-negative integer values, computes `total` and `average`
- **PhysicalMeasurement**: Handles unit conversion (decimeters → meters, hectograms → kilograms) since PokeAPI returns raw units
- **Sprites**: Validates URLs, implements `getBestQuality()` fallback strategy (Official Artwork → front default → front shiny)

### 3.3 Domain Service — TypeEffectivenessService

```typescript
export type DamageMultiplier = 0 | 0.5 | 1 | 2;

export class TypeEffectivenessService {
  // Complete 18x18 type effectiveness chart
  getAttackMultiplier(attacking: PokemonType, defending: PokemonType): DamageMultiplier { /* ... */ }
  getWeaknesses(defendingTypes: readonly PokemonType[]): PokemonTypeName[] { /* ... */ }
  getResistances(defendingTypes: readonly PokemonType[]): PokemonTypeName[] { /* ... */ }
  getImmunities(defendingTypes: readonly PokemonType[]): PokemonTypeName[] { /* ... */ }
}
```

This is a **Domain Service** (not an entity method) because type effectiveness involves the **interaction between two Pokemon** — it's domain knowledge that transcends any individual entity.

### 3.4 Domain Errors

```typescript
export abstract class DomainError extends Error {
  readonly code: string;
  constructor(message: string, code: string) {
    super(message);
    this.code = code;
    this.name = this.constructor.name;
    Object.setPrototypeOf(this, new.target.prototype);  // Fix for Error inheritance in ES5
  }
}

// Concrete errors
export class InvalidPokemonTypeError extends DomainError { /* ... */ }
export class InvalidStatsError extends DomainError { /* ... */ }
export class InvalidPhysicalMeasurementError extends DomainError { /* ... */ }
export class InvalidSpriteUrlError extends DomainError { /* ... */ }
export class InvalidPokemonEntityError extends DomainError { /* ... */ }
export class InvalidSpeciesError extends DomainError { /* ... */ }
```

Each error has a `code` string (e.g., `'INVALID_POKEMON_TYPE'`) for programmatic handling without depending on human-readable messages.

---

## 4. Application Layer

The application layer orchestrates data flow between the domain and the outside world. It contains **Use Cases**, **DTOs**, **Ports** (interfaces), and shared types.

### 4.1 Ports (Outbound Interfaces)

Ports define **what the application needs from the outside world** without specifying how:

```typescript
// application/ports/PokemonRepository.ts
export interface PokemonRepository {
  getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>>;
  getById(id: number): Promise<Pokemon>;
  getSpeciesById(id: number): Promise<Species>;
  searchByName(name: string, params: PaginationParams): Promise<PaginatedResult<Pokemon>>;
}

// application/ports/Logger.ts
export interface Logger {
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, error?: unknown, context?: Record<string, unknown>): void;
  debug(message: string, context?: Record<string, unknown>): void;
}

// application/ports/CacheService.ts
export interface CacheService {
  get(key: string): unknown;
  set(key: string, value: unknown, ttlMs: number): void;
  has(key: string): boolean;
  delete(key: string): void;
  clear(): void;
}
```

### 4.2 DTOs (Data Transfer Objects)

DTOs define the shape of data that **exits** the use cases toward the presentation:

```typescript
// Minimal DTO for list cards
export interface PokemonListItemDTO {
  id: number;
  name: string;
  types: PokemonTypeName[];
  spriteUrl: string | null;
}

// Complete DTO for detail page
export interface PokemonDetailDTO {
  id: number;
  name: string;
  types: PokemonTypeName[];
  stats: StatsData;
  height: string;          // Pre-converted: "0.70 m"
  weight: string;          // Pre-converted: "6.90 kg"
  spriteUrl: string | null;
  officialArtworkUrl: string | null;
  sprites: { frontDefault: string | null; /* ... */ };
  baseExperience: number | null;
  species: SpeciesDTO | null;
}
```

**Why DTOs instead of passing entities directly?**

1. **Decoupling**: Presentation does NOT depend on domain entities
2. **Optimized shape**: `PokemonListItemDTO` has only 4 fields vs. the full entity
3. **Pre-processed data**: `height` and `weight` are already converted to display strings — the UI doesn't need to know the API returns decimeters
4. **Stable contract**: DTOs are the contract between Application and Presentation

### 4.3 Use Cases

Each use case follows the **Interface + Implementation** pattern:

#### ListPokemonUseCase

```typescript
export class ListPokemonUseCaseImpl implements ListPokemonUseCase {
  constructor(
    pokemonRepository: PokemonRepository,
    logger: Logger
  ) { /* ... */ }

  async execute(params: PaginationParams): Promise<PaginatedResult<PokemonListItemDTO>> {
    this.logger.info('ListPokemonUseCase.execute', { page: params.page, limit: params.limit });
    const result = await this.pokemonRepository.getAll(params);
    return { ...result, data: result.data.map((pokemon) => this.toDTO(pokemon)) };
  }
}
```

#### GetPokemonByIdUseCase — Graceful Degradation

```typescript
async execute(id: number): Promise<PokemonDetailDTO> {
  const pokemon = await this.pokemonRepository.getById(id);

  // Species fetch can fail — graceful degradation
  let speciesDTO: SpeciesDTO | null = null;
  try {
    const species = await this.pokemonRepository.getSpeciesById(id);
    speciesDTO = this.speciesToDTO(species);
  } catch (error: unknown) {
    this.logger.warn('Failed to fetch species data', { id, error });
    // No re-throw: Pokemon is shown without species data
  }

  return this.toDTO(pokemon, speciesDTO);
}
```

**Graceful Degradation pattern**: If Species fetch fails, the use case does NOT throw. It logs a warning and returns the DTO with `species: null`. The UI shows everything except the Species Info section. This is a **business decision** — partial data is preferred over no data.

#### SearchPokemonUseCase

```typescript
async execute(name: string, params: PaginationParams): Promise<PaginatedResult<PokemonListItemDTO>> {
  const trimmedName = name.trim().toLowerCase();  // Input normalization
  if (trimmedName === '') {
    return { data: [], total: 0, /* ... */ };  // Guard clause
  }
  const result = await this.pokemonRepository.searchByName(trimmedName, params);
  return { ...result, data: result.data.map((pokemon) => this.toDTO(pokemon)) };
}
```

Search term normalization (`trim().toLowerCase()`) happens in the use case, NOT in the UI — guaranteeing consistent behavior regardless of the caller.

---

## 5. Infrastructure Layer

Infrastructure contains the **concrete implementations** of ports defined in Application. It's the layer that "talks" to the outside world.

### 5.1 PokemonMapper (API → Domain)

```typescript
export function mapPokemonToDomain(response: PokeAPIPokemonResponse): Pokemon {
  const props: PokemonProps = {
    id: response.id,
    name: response.name,
    types: response.types.sort((a, b) => a.slot - b.slot).map((t) => t.type.name),
    stats: mapStats(response),
    height: { value: response.height, unit: 'dm' },   // Raw: decimeters
    weight: { value: response.weight, unit: 'hg' },   // Raw: hectograms
    sprites: { /* ... */ },
    order: response.order,
    baseExperience: response.base_experience,
  };

  return Pokemon.create(props);  // Domain factory method validates everything
}
```

**Mapper responsibilities:**
1. Transform flat API structure to rich domain structure
2. Normalize data (e.g., clean control characters from flavor text)
3. Sort types by slot (primary first, secondary second)
4. Map generation strings to numbers
5. Delegate validation to domain (`Pokemon.create()` throws on invalid data)

### 5.2 PokeAPIRepository (Adapter)

```typescript
export class PokeAPIRepository implements PokemonRepository {
  async getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    const offset = (params.page - 1) * params.limit;

    // Step 1: Get the list (names and URLs only)
    const listResponse = await this.fetchJSON<PokeAPIListResponse>(
      `${BASE_URL}/pokemon?limit=${String(params.limit)}&offset=${String(offset)}`
    );

    // Step 2: Parallel fetch of each individual Pokemon (N+1 pattern)
    const pokemonPromises = listResponse.results.map((item) =>
      this.fetchJSON<PokeAPIPokemonResponse>(item.url).then(mapPokemonToDomain)
    );

    const data = await Promise.all(pokemonPromises);
    // ... return PaginatedResult
  }

  async searchByName(name: string, params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    // PokeAPI has no search endpoint — fetch all + filter client-side
    const allResponse = await this.fetchJSON<PokeAPIListResponse>(
      `${BASE_URL}/pokemon?limit=1302&offset=0`
    );
    const matchingItems = allResponse.results.filter((item) => item.name.includes(name));
    // ... paginate and fetch matching items in parallel
  }
}
```

### 5.3 LocalStorageCacheService

TTL-based cache with prefix isolation and defensive JSON parsing. Auto-cleans expired and corrupted entries.

### 5.4 ConsoleLogger

Simple development implementation. Can be replaced with Sentry, DataDog, or any logging service **without changing use cases** — only the DI container wiring changes.

---

## 6. Dependency Injection (DI Container)

### 6.1 Composition Root

The **Composition Root** is the **only place** where concrete instances are created and wired together:

```typescript
// di/container.ts
export function createContainer(): DIContainer {
  // 1. Create concrete infrastructure implementations
  const logger = new ConsoleLogger();
  const cacheService = new LocalStorageCacheService();
  const pokemonRepository = new PokeAPIRepository();

  // 2. Inject dependencies into use cases
  const listPokemonUseCase = new ListPokemonUseCaseImpl(pokemonRepository, logger);
  const getPokemonByIdUseCase = new GetPokemonByIdUseCaseImpl(pokemonRepository, logger);
  const searchPokemonUseCase = new SearchPokemonUseCaseImpl(pokemonRepository, logger);

  // 3. Return container typed with interfaces (not implementations)
  return { pokemonRepository, cacheService, logger,
           listPokemonUseCase, getPokemonByIdUseCase, searchPokemonUseCase };
}
```

**This is the ONLY file that knows about concrete implementations.** Use cases receive interfaces (`PokemonRepository`, `Logger`), not concrete classes.

### 6.2 React Context for DI

```typescript
// di/DIContext.tsx — React 19: Context used directly without .Provider
export const DIContext = createContext<DIContainer | null>(null);

// di/useDI.ts — Custom hook with null check
export function useDI(): DIContainer {
  const context = useContext(DIContext);
  if (!context) throw new Error('useDI must be used within a DIProvider');
  return context;
}
```

### 6.3 Provider Hierarchy (main.tsx)

```typescript
createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <DIContext value={container}>
        <RouterProvider router={router} />
      </DIContext>
    </QueryClientProvider>
  </StrictMode>
);
```

---

## 7. Presentation Layer

### 7.1 Routing & Code Splitting

```typescript
const PokemonDetailPage = lazy(() =>
  import('@/presentation/pages/PokemonDetailPage').then((m) => ({
    default: m.PokemonDetailPage,
  }))
);

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <PokemonListPage /> },
      {
        path: 'pokemon/:id',
        element: (
          <Suspense fallback={<PokemonDetailSkeleton />}>
            <PokemonDetailPage />
          </Suspense>
        ),
      },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
]);
```

`PokemonDetailPage` is loaded on demand. The chunk downloads only when the user navigates to `/pokemon/:id`, reducing the initial bundle size.

### 7.2 Custom Hooks (TanStack Query)

```typescript
// usePokemonList.ts
export function usePokemonList() {
  const { listPokemonUseCase } = useDI();
  const [page, setPage] = useState(1);
  const query = useQuery({
    queryKey: ['pokemon', 'list', page],
    queryFn: () => listPokemonUseCase.execute({ page, limit: 20 }),
  });
  return { ...query, page, setPage };
}

// useSearchPokemon.ts — with debounce
export function useSearchPokemon(searchTerm: string) {
  const { searchPokemonUseCase } = useDI();
  const [page, setPage] = useState(1);
  const debouncedTerm = useDebounce(searchTerm, 300);

  useEffect(() => { setPage(1); }, [debouncedTerm]);

  const query = useQuery({
    queryKey: ['pokemon', 'search', debouncedTerm, page],
    queryFn: () => searchPokemonUseCase.execute(debouncedTerm, { page, limit: 20 }),
    enabled: debouncedTerm.length > 0,
  });
  return { ...query, page, setPage };
}
```

### 7.3 Zustand Store (Global State)

```typescript
export const useFavoritesStore = create<FavoritesState>()(
  persist(
    (set, get) => ({
      favoriteIds: [],
      toggleFavorite: (id: number) => {
        set((state) => ({
          favoriteIds: state.favoriteIds.includes(id)
            ? state.favoriteIds.filter((fid) => fid !== id)
            : [...state.favoriteIds, id],
        }));
      },
      isFavorite: (id: number) => get().favoriteIds.includes(id),
    }),
    { name: 'pokedex-favorites' }
  )
);
```

Zustand + persist middleware: favorites auto-persist in `localStorage`. Not server state (that's TanStack Query's job), not prop drilling — it's user-local state that multiple components need access to.

### 7.4 UI Components

- **Shadcn/ui**: Copied (not installed), customizable, built on Radix UI primitives for accessibility
- **CVA (Class Variance Authority)**: Generates CSS classes dynamically based on variant props
- **ErrorBoundary**: Class component (required — React has no functional equivalent for `getDerivedStateFromError`)
- **Lazy-loaded images**: Native `loading="lazy"` on sprites

---

## 8. SOLID Principles — Project Examples

### S — Single Responsibility

Each use case has **exactly one responsibility**. The mapper only transforms data. Each detail page component owns one visual section (`StatsChart`, `SpeciesInfo`, `PokemonSprites`).

### O — Open/Closed

To add a new data source (e.g., GraphQL): create `GraphQLPokemonRepository implements PokemonRepository`, change one line in `createContainer()`. No use case, hook, or component is modified.

### L — Liskov Substitution

`PokeAPIRepository` and `MockPokemonRepository` are interchangeable — the use case cannot distinguish between them. Both fulfill the `PokemonRepository` contract.

### I — Interface Segregation

Three separate ports (`PokemonRepository`, `Logger`, `CacheService`) instead of one monolithic interface. Each use case interface has a single `execute` method. DTOs are segregated (`PokemonListItemDTO` vs `PokemonDetailDTO`).

### D — Dependency Inversion

The most visible principle:
```
High-level:  ListPokemonUseCaseImpl  →  depends on  →  PokemonRepository (INTERFACE)
Low-level:   PokeAPIRepository       →  implements  →  PokemonRepository (INTERFACE)
```

The interface is defined in **Application** (high-level), not Infrastructure (low-level). The DI Container materializes this inversion.

---

## 9. Data Fetching & Caching Strategy

### 9.1 The N+1 Pattern (and why it works here)

PokeAPI's list endpoint returns only names and URLs. To show cards (with sprites, types, etc.), we make **1 request for the list + 20 individual requests** = 21 total.

**Why this performs well:**
1. **`Promise.all`**: 20 fetches execute **in parallel**
2. **HTTP/2 multiplexing**: Multiple requests over a single TCP connection
3. **Small payloads**: Each response is ~15-30KB
4. **Browser HTTP cache**: `Cache-Control` headers enable automatic disk caching

### 9.2 Multi-Layer Cache Strategy

```
┌─────────────────────────────────────────────────────────┐
│                   Layer 1: TanStack Query                │
│                  (In-Memory, staleTime: 5min)            │
│  Instant serving for "fresh" data, background refetch    │
│  for "stale" data, gcTime: 30min retention               │
└──────────────────────┬──────────────────────────────────┘
                       │ Cache miss...
                       ▼
┌─────────────────────────────────────────────────────────┐
│                 Layer 2: Browser HTTP Cache               │
│               (Disk cache, Cache-Control headers)         │
│  Chrome serves from disk if not expired                   │
│  Network tab shows "(disk cache)" — no server hit         │
└──────────────────────┬──────────────────────────────────┘
                       │ Cache miss...
                       ▼
┌─────────────────────────────────────────────────────────┐
│                  Layer 3: PokeAPI Server                  │
│  Only reached on first request, expired cache,            │
│  or hard refresh (Ctrl+Shift+R)                           │
└─────────────────────────────────────────────────────────┘
```

### 9.3 Navigation Flow: List → Detail → Back

```
User opens app → PokemonListPage
  TanStack Query: queryKey = ['pokemon', 'list', 1]
  → No cache → real fetch
  → 20 parallel fetches, browser caches all responses on disk
  → TanStack Query stores result in memory

User clicks Bulbasaur → PokemonDetailPage
  TanStack Query: queryKey = ['pokemon', 'detail', 1]
  → No memory cache for this key → fetch
  → GET /pokemon/1 → browser already has it in disk cache → "(disk cache)"
  → GET /pokemon-species/1 → may be a real network request

User clicks "Back to list" → PokemonListPage
  TanStack Query: queryKey = ['pokemon', 'list', 1]
  → Memory cache HIT (staleTime not exceeded)
  → Data served INSTANTLY, no fetch
```

### 9.4 TanStack Query Configuration

```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,      // 5 minutes
      gcTime: 30 * 60 * 1000,         // 30 minutes
      retry: 2,
      refetchOnWindowFocus: false,
    },
  },
});
```

---

## 10. Testing Strategy

### 10.1 Testing Pyramid

```
        ╱╲
       ╱  ╲       E2E Tests (Playwright): 17 tests
      ╱    ╲      Real navigation, complete flows
     ╱──────╲
    ╱        ╲    Integration Tests: hooks, stores
   ╱          ╲   Components with mocked providers
  ╱────────────╲
 ╱              ╲  Unit Tests (Vitest): 192 tests
╱                ╲ Domain, Application, Infrastructure, Presentation
╱──────────────────╲
```

### 10.2 Coverage by Layer

- **Domain**: 100% — All entities, VOs, Domain Service, and errors
- **Application**: 90%+ — All use cases with mocked ports
- **Infrastructure**: 80%+ — Mapper, CacheService, Logger
- **Presentation**: 70%+ — Components with Testing Library, hooks

### 10.3 Mock Pattern

```typescript
// Type-safe mock pattern to avoid eslint unbound-method errors
const mockRepository: { [K in keyof PokemonRepository]: Mock } = {
  getAll: vi.fn(),
  getById: vi.fn(),
  getSpeciesById: vi.fn(),
  searchByName: vi.fn(),
};
```

### 10.4 E2E Tests (Playwright)

```
tests/e2e/
├── pokemon-list.spec.ts    — 7 tests (title, cards, pagination, search)
├── pokemon-detail.spec.ts  — 5 tests (navigation, detail, stats, favorite)
├── favorites.spec.ts       — 2 tests (toggle from list and detail)
└── navigation.spec.ts      — 3 tests (404, home, logo)
```

Configured to run separately from Vitest (`vitest.config.ts` excludes `tests/e2e/**`).

---

## 11. Technical Deep Dive (Q&A)

### Architecture

**Q: Why Clean Architecture for a frontend application?**

The layers don't add much code — mostly interfaces and a factory function (`createContainer`). In return: tests that don't depend on React or any API, the ability to swap data sources without touching the UI, and a folder structure that self-documents the architecture.

**Q: Clean Architecture vs Hexagonal Architecture?**

Same concept, different emphasis. Clean Architecture (Uncle Bob) emphasizes concentric layers and the Dependency Rule. Hexagonal/Ports & Adapters (Alistair Cockburn) emphasizes ports (interfaces) and adapters (implementations). This project uses both vocabularies.

### Domain-Driven Design

**Q: Entity vs Value Object?**

- **Entity** (Pokemon, Species): Has identity. Two Pokemon with the same data but different IDs are DIFFERENT.
- **Value Object** (PokemonType, Stats, Sprites): No identity. Two `PokemonType('fire')` are equal. Compared by value.

**Q: Why factory methods instead of public constructors?**

`Pokemon.create(props)` runs all validation before instantiation. If it returns without throwing, the Pokemon is guaranteed valid.

### React & Frontend

**Q: Why TanStack Query over useEffect + useState?**

Deduplication, automatic cache, background refetch, retry, loading/error/success states, and race condition handling — all out of the box.

**Q: Why Zustand over Redux or Context API for favorites?**

Redux: too much boilerplate. Context API: causes unnecessary re-renders. Zustand: minimal boilerplate, granular selectors, built-in persist middleware, works outside React (testable).

### Design Decisions

**Q: Why does search fetch all 1,302 Pokemon?**

PokeAPI has no search endpoint. The list payload (name + url x 1,302) is only ~50KB. The browser caches it in HTTP cache, making subsequent searches instant.

**Q: What happens if PokeAPI goes down?**

1. TanStack Query retries automatically (2x)
2. If still failing, `isError: true` triggers `<ErrorState />` with retry button
3. If user had cached data (staleTime not exceeded), stale data continues to display

**Q: DI pattern without a framework like NestJS?**

Manual DI with three components: `createContainer()` (factory function), `DIContext` (React Context), `useDI()` (custom hook). Simple enough for a frontend app — no decorators, tokens, modules, or scopes needed.

---

## Appendix: Technology Stack

| Category | Technology | Version |
|----------|------------|---------|
| Runtime | Node.js | 22 (Alpine) |
| Framework | React | 19.2 |
| Language | TypeScript | 5.9 |
| Build Tool | Vite | 7.2 |
| Styling | TailwindCSS | 4.1 |
| Component Library | Shadcn/ui (manual) | — |
| Data Fetching | TanStack Query | 5.90 |
| Routing | React Router | 7.13 |
| State Management | Zustand | 5.0 |
| Forms | React Hook Form + Zod | 7.71 / 4.3 |
| Testing (Unit) | Vitest + Testing Library | 4.0 / 16.3 |
| Testing (E2E) | Playwright | 1.58 |
| Linting | ESLint | 9.39 |
| Formatting | Prettier | 3.8 |
| Git Hooks | Husky + lint-staged | 9.1 / 16.2 |
| Container | Docker + Docker Compose | — |
| Package Manager | pnpm | 9 |
