import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router';
import type { Mock } from 'vitest';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { PokemonDetailDTO } from '@/application/dto';
import type { DIContainer } from '@/di/container';
import { DIContext } from '@/di/DIContext';
import { PokemonDetailPage } from '@/presentation/pages/PokemonDetailPage';

function createMockContainer() {
  return {
    pokemonRepository: {} as DIContainer['pokemonRepository'],
    cacheService: {} as DIContainer['cacheService'],
    logger: { info: vi.fn(), warn: vi.fn(), error: vi.fn(), debug: vi.fn() },
    listPokemonUseCase: { execute: vi.fn() },
    getPokemonByIdUseCase: { execute: vi.fn() },
    searchPokemonUseCase: { execute: vi.fn() },
  } as unknown as DIContainer;
}

const mockPokemonDetail: PokemonDetailDTO = {
  id: 25,
  name: 'pikachu',
  types: ['electric'],
  stats: { hp: 35, attack: 55, defense: 40, specialAttack: 50, specialDefense: 50, speed: 90 },
  height: '0.4 m',
  weight: '6.0 kg',
  spriteUrl: 'https://example.com/pikachu.png',
  officialArtworkUrl: 'https://example.com/pikachu-art.png',
  sprites: {
    frontDefault: 'https://example.com/front.png',
    frontShiny: 'https://example.com/front-shiny.png',
    backDefault: 'https://example.com/back.png',
    backShiny: null,
  },
  baseExperience: 112,
  species: {
    id: 25,
    name: 'pikachu',
    generation: 1,
    flavorText:
      'When several of these Pokemon gather, their electricity could build and cause lightning storms.',
    habitat: 'forest',
    isLegendary: false,
    isMythical: false,
  },
};

function renderPage(container: DIContainer) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <DIContext value={container}>
        <MemoryRouter initialEntries={['/pokemon/25']}>
          <Routes>
            <Route path="/pokemon/:id" element={<PokemonDetailPage />} />
          </Routes>
        </MemoryRouter>
      </DIContext>
    </QueryClientProvider>
  );
}

describe('PokemonDetailPage', () => {
  let mockContainer: DIContainer;

  beforeEach(() => {
    mockContainer = createMockContainer();
  });

  it('shows skeleton while loading', () => {
    (mockContainer.getPokemonByIdUseCase.execute as Mock).mockReturnValue(
      new Promise(() => {
        /* never resolves */
      })
    );

    renderPage(mockContainer);

    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders pokemon detail after loading', async () => {
    (mockContainer.getPokemonByIdUseCase.execute as Mock).mockResolvedValue(mockPokemonDetail);

    renderPage(mockContainer);

    await waitFor(() => {
      expect(screen.getByText('pikachu')).toBeInTheDocument();
    });

    expect(screen.getByText('#025')).toBeInTheDocument();
    expect(screen.getByText('electric')).toBeInTheDocument();
    expect(screen.getByText('0.4 m')).toBeInTheDocument();
    expect(screen.getByText('6.0 kg')).toBeInTheDocument();
  });

  it('renders stats chart', async () => {
    (mockContainer.getPokemonByIdUseCase.execute as Mock).mockResolvedValue(mockPokemonDetail);

    renderPage(mockContainer);

    await waitFor(() => {
      expect(screen.getByText('Base Stats')).toBeInTheDocument();
    });

    expect(screen.getByText('HP')).toBeInTheDocument();
    expect(screen.getByText('Attack')).toBeInTheDocument();
  });

  it('renders species info when available', async () => {
    (mockContainer.getPokemonByIdUseCase.execute as Mock).mockResolvedValue(mockPokemonDetail);

    renderPage(mockContainer);

    await waitFor(() => {
      expect(screen.getByText('Species Info')).toBeInTheDocument();
    });

    expect(screen.getByText('Gen 1')).toBeInTheDocument();
    expect(screen.getByText('forest')).toBeInTheDocument();
  });

  it('shows error state when fetch fails', async () => {
    (mockContainer.getPokemonByIdUseCase.execute as Mock).mockRejectedValue(new Error('Not found'));

    renderPage(mockContainer);

    await waitFor(() => {
      expect(
        screen.getByText('Failed to load Pokemon details. Please try again.')
      ).toBeInTheDocument();
    });
  });

  it('renders back navigation link', () => {
    (mockContainer.getPokemonByIdUseCase.execute as Mock).mockReturnValue(
      new Promise(() => {
        /* never resolves */
      })
    );

    renderPage(mockContainer);

    expect(screen.getByText('Back to list')).toBeInTheDocument();
  });
});
