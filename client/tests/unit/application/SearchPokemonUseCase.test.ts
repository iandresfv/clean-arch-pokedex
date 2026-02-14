import type { Mock } from 'vitest';

import type { Logger, PokemonRepository } from '@/application/ports';
import type { PaginatedResult } from '@/application/types';
import { SearchPokemonUseCaseImpl } from '@/application/use-cases/search-pokemon';
import type { Pokemon } from '@/domain/pokemon';

const mockPokemon = {
  id: 25,
  name: 'Pikachu',
  types: [{ value: 'electric' }],
  sprites: { getBestQuality: () => 'https://example.com/pikachu.png' },
} as unknown as Pokemon;

const mockLogger: { [K in keyof Logger]: Mock } = {
  info: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
  debug: vi.fn(),
};

const mockRepository: { [K in keyof PokemonRepository]: Mock } = {
  getAll: vi.fn(),
  getById: vi.fn(),
  getSpeciesById: vi.fn(),
  searchByName: vi.fn(),
};

describe('SearchPokemonUseCaseImpl', () => {
  let useCase: SearchPokemonUseCaseImpl;

  beforeEach(() => {
    vi.clearAllMocks();
    useCase = new SearchPokemonUseCaseImpl(mockRepository, mockLogger);
  });

  it('should return search results as DTOs', async () => {
    const paginatedResult: PaginatedResult<Pokemon> = {
      data: [mockPokemon],
      total: 1,
      page: 1,
      limit: 20,
      totalPages: 1,
      hasNextPage: false,
      hasPreviousPage: false,
    };

    mockRepository.searchByName.mockResolvedValue(paginatedResult);

    const result = await useCase.execute('pikachu', { page: 1, limit: 20 });

    expect(result.data).toHaveLength(1);
    expect(result.data[0]).toEqual({
      id: 25,
      name: 'Pikachu',
      types: ['electric'],
      spriteUrl: 'https://example.com/pikachu.png',
    });
  });

  it('should normalize search term (trim + lowercase)', async () => {
    mockRepository.searchByName.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
      hasNextPage: false,
      hasPreviousPage: false,
    });

    await useCase.execute('  PIKACHU  ', { page: 1, limit: 20 });

    expect(mockRepository.searchByName).toHaveBeenCalledWith('pikachu', { page: 1, limit: 20 });
  });

  it('should return empty result for empty search term', async () => {
    const result = await useCase.execute('', { page: 1, limit: 20 });

    expect(result.data).toHaveLength(0);
    expect(result.total).toBe(0);
    expect(mockRepository.searchByName).not.toHaveBeenCalled();
  });

  it('should return empty result for whitespace-only search term', async () => {
    const result = await useCase.execute('   ', { page: 1, limit: 20 });

    expect(result.data).toHaveLength(0);
    expect(mockRepository.searchByName).not.toHaveBeenCalled();
  });

  it('should propagate repository errors', async () => {
    mockRepository.searchByName.mockRejectedValue(new Error('Network error'));

    await expect(useCase.execute('pikachu', { page: 1, limit: 20 })).rejects.toThrow(
      'Network error'
    );
  });
});
