import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import type { Mock } from 'vitest';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { PokemonListItemDTO } from '@/application/dto';
import type { DIContainer } from '@/di/container';
import { DIContext } from '@/di/DIContext';
import { PokemonListPage } from '@/presentation/pages/PokemonListPage';

function createMockContainer(overrides: Partial<{ [K in keyof DIContainer]: Mock }> = {}) {
  return {
    pokemonRepository: {} as DIContainer['pokemonRepository'],
    cacheService: {} as DIContainer['cacheService'],
    logger: { info: vi.fn(), warn: vi.fn(), error: vi.fn(), debug: vi.fn() },
    listPokemonUseCase: { execute: vi.fn() },
    getPokemonByIdUseCase: { execute: vi.fn() },
    searchPokemonUseCase: { execute: vi.fn() },
    ...overrides,
  } as unknown as DIContainer;
}

const mockPokemonList: PokemonListItemDTO[] = [
  { id: 1, name: 'bulbasaur', types: ['grass', 'poison'], spriteUrl: 'https://example.com/1.png' },
  { id: 2, name: 'ivysaur', types: ['grass', 'poison'], spriteUrl: 'https://example.com/2.png' },
  { id: 3, name: 'venusaur', types: ['grass', 'poison'], spriteUrl: 'https://example.com/3.png' },
];

function renderPage(container: DIContainer) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <DIContext value={container}>
        <MemoryRouter>
          <PokemonListPage />
        </MemoryRouter>
      </DIContext>
    </QueryClientProvider>
  );
}

describe('PokemonListPage', () => {
  let mockContainer: DIContainer;

  beforeEach(() => {
    mockContainer = createMockContainer();
  });

  it('shows loading skeletons while fetching', () => {
    (mockContainer.listPokemonUseCase.execute as Mock).mockReturnValue(
      new Promise(() => {
        /* never resolves */
      })
    );

    renderPage(mockContainer);

    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders pokemon cards after loading', async () => {
    (mockContainer.listPokemonUseCase.execute as Mock).mockResolvedValue({
      data: mockPokemonList,
      total: 3,
      page: 1,
      limit: 20,
      totalPages: 1,
      hasNextPage: false,
      hasPreviousPage: false,
    });

    renderPage(mockContainer);

    await waitFor(() => {
      expect(screen.getByText('bulbasaur')).toBeInTheDocument();
    });

    expect(screen.getByText('ivysaur')).toBeInTheDocument();
    expect(screen.getByText('venusaur')).toBeInTheDocument();
  });

  it('shows error state when fetch fails', async () => {
    (mockContainer.listPokemonUseCase.execute as Mock).mockRejectedValue(
      new Error('Network error')
    );

    renderPage(mockContainer);

    await waitFor(() => {
      expect(screen.getByText('Failed to load Pokemon. Please try again.')).toBeInTheDocument();
    });

    expect(screen.getByText('Try again')).toBeInTheDocument();
  });

  it('shows pagination when data is loaded', async () => {
    (mockContainer.listPokemonUseCase.execute as Mock).mockResolvedValue({
      data: mockPokemonList,
      total: 100,
      page: 1,
      limit: 20,
      totalPages: 5,
      hasNextPage: true,
      hasPreviousPage: false,
    });

    renderPage(mockContainer);

    await waitFor(() => {
      expect(screen.getByText('Page 1 of 5')).toBeInTheDocument();
    });

    expect(screen.getByText('Previous')).toBeDisabled();
    expect(screen.getByText('Next')).toBeEnabled();
  });

  it('renders the page title', () => {
    (mockContainer.listPokemonUseCase.execute as Mock).mockReturnValue(
      new Promise(() => {
        /* never resolves */
      })
    );

    renderPage(mockContainer);

    expect(screen.getByText('Pokédex')).toBeInTheDocument();
  });
});
