import type { Mock } from 'vitest';

import type { Logger, PokemonRepository } from '@/application/ports';
import type { PaginatedResult } from '@/application/types';
import { ListPokemonUseCaseImpl } from '@/application/use-cases/list-pokemon';
import type { Pokemon } from '@/domain/pokemon';

const mockPokemon = {
  id: 1,
  name: 'Bulbasaur',
  types: [{ value: 'grass' }, { value: 'poison' }],
  sprites: { getBestQuality: () => 'https://example.com/bulbasaur.png' },
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

describe('ListPokemonUseCaseImpl', () => {
  let useCase: ListPokemonUseCaseImpl;

  beforeEach(() => {
    vi.clearAllMocks();
    useCase = new ListPokemonUseCaseImpl(mockRepository, mockLogger);
  });

  it('should return paginated list of PokemonListItemDTOs', async () => {
    const paginatedResult: PaginatedResult<Pokemon> = {
      data: [mockPokemon],
      total: 1,
      page: 1,
      limit: 20,
      totalPages: 1,
      hasNextPage: false,
      hasPreviousPage: false,
    };

    mockRepository.getAll.mockResolvedValue(paginatedResult);

    const result = await useCase.execute({ page: 1, limit: 20 });

    expect(result.data).toHaveLength(1);
    expect(result.data[0]).toEqual({
      id: 1,
      name: 'Bulbasaur',
      types: ['grass', 'poison'],
      spriteUrl: 'https://example.com/bulbasaur.png',
    });
    expect(result.total).toBe(1);
    expect(result.page).toBe(1);
  });

  it('should pass pagination params to repository', async () => {
    mockRepository.getAll.mockResolvedValue({
      data: [],
      total: 0,
      page: 3,
      limit: 10,
      totalPages: 0,
      hasNextPage: false,
      hasPreviousPage: true,
    });

    await useCase.execute({ page: 3, limit: 10 });

    expect(mockRepository.getAll).toHaveBeenCalledWith({ page: 3, limit: 10 });
  });

  it('should log execution', async () => {
    mockRepository.getAll.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      limit: 20,
      totalPages: 0,
      hasNextPage: false,
      hasPreviousPage: false,
    });

    await useCase.execute({ page: 1, limit: 20 });

    expect(mockLogger.info).toHaveBeenCalledWith('ListPokemonUseCase.execute', {
      page: 1,
      limit: 20,
    });
  });

  it('should propagate repository errors', async () => {
    mockRepository.getAll.mockRejectedValue(new Error('Network error'));

    await expect(useCase.execute({ page: 1, limit: 20 })).rejects.toThrow('Network error');
  });
});
