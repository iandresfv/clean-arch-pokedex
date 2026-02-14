import type { Mock } from 'vitest';

import type { Logger, PokemonRepository } from '@/application/ports';
import { GetPokemonByIdUseCaseImpl } from '@/application/use-cases/get-pokemon-by-id';
import type { Pokemon, Species } from '@/domain/pokemon';

const mockPokemon = {
  id: 25,
  name: 'Pikachu',
  types: [{ value: 'electric' }],
  stats: {
    toObject: () => ({
      hp: 35,
      attack: 55,
      defense: 40,
      specialAttack: 50,
      specialDefense: 50,
      speed: 90,
    }),
  },
  height: { toDisplayString: () => '0.40 m' },
  weight: { toDisplayString: () => '6.00 kg' },
  sprites: {
    getBestQuality: () => 'https://example.com/pikachu.png',
    officialArtwork: 'https://example.com/pikachu-official.png',
    frontDefault: 'https://example.com/front.png',
    frontShiny: 'https://example.com/shiny.png',
    backDefault: null,
    backShiny: null,
  },
  baseExperience: 112,
} as unknown as Pokemon;

const mockSpecies = {
  id: 25,
  name: 'pikachu',
  generation: 1,
  flavorText:
    'When several of these Pokemon gather, their electricity can build and cause lightning storms.',
  habitat: 'forest',
  isLegendary: false,
  isMythical: false,
} as unknown as Species;

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

describe('GetPokemonByIdUseCaseImpl', () => {
  let useCase: GetPokemonByIdUseCaseImpl;

  beforeEach(() => {
    vi.clearAllMocks();
    useCase = new GetPokemonByIdUseCaseImpl(mockRepository, mockLogger);
  });

  it('should return a PokemonDetailDTO with species data', async () => {
    mockRepository.getById.mockResolvedValue(mockPokemon);
    mockRepository.getSpeciesById.mockResolvedValue(mockSpecies);

    const result = await useCase.execute(25);

    expect(result.id).toBe(25);
    expect(result.name).toBe('Pikachu');
    expect(result.types).toEqual(['electric']);
    expect(result.height).toBe('0.40 m');
    expect(result.weight).toBe('6.00 kg');
    expect(result.spriteUrl).toBe('https://example.com/pikachu.png');
    expect(result.species).toEqual({
      id: 25,
      name: 'pikachu',
      generation: 1,
      flavorText: expect.any(String) as string,
      habitat: 'forest',
      isLegendary: false,
      isMythical: false,
    });
  });

  it('should return null species when species fetch fails', async () => {
    mockRepository.getById.mockResolvedValue(mockPokemon);
    mockRepository.getSpeciesById.mockRejectedValue(new Error('Species not found'));

    const result = await useCase.execute(25);

    expect(result.id).toBe(25);
    expect(result.species).toBeNull();
    expect(mockLogger.warn).toHaveBeenCalledWith('Failed to fetch species data', {
      id: 25,
      error: expect.any(Error) as Error,
    });
  });

  it('should include all sprite URLs', async () => {
    mockRepository.getById.mockResolvedValue(mockPokemon);
    mockRepository.getSpeciesById.mockResolvedValue(mockSpecies);

    const result = await useCase.execute(25);

    expect(result.sprites).toEqual({
      frontDefault: 'https://example.com/front.png',
      frontShiny: 'https://example.com/shiny.png',
      backDefault: null,
      backShiny: null,
    });
    expect(result.officialArtworkUrl).toBe('https://example.com/pikachu-official.png');
  });

  it('should propagate repository errors for pokemon fetch', async () => {
    mockRepository.getById.mockRejectedValue(new Error('Not found'));

    await expect(useCase.execute(999)).rejects.toThrow('Not found');
  });

  it('should log execution', async () => {
    mockRepository.getById.mockResolvedValue(mockPokemon);
    mockRepository.getSpeciesById.mockResolvedValue(mockSpecies);

    await useCase.execute(25);

    expect(mockLogger.info).toHaveBeenCalledWith('GetPokemonByIdUseCase.execute', { id: 25 });
  });
});
