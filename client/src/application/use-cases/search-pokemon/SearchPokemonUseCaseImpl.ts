import type { SearchPokemonUseCase } from './SearchPokemonUseCase';

import type { PokemonListItemDTO } from '@/application/dto';
import type { Logger, PokemonRepository } from '@/application/ports';
import type { PaginatedResult, PaginationParams } from '@/application/types';
import type { Pokemon } from '@/domain/pokemon';

export class SearchPokemonUseCaseImpl implements SearchPokemonUseCase {
  private readonly pokemonRepository: PokemonRepository;
  private readonly logger: Logger;

  constructor(pokemonRepository: PokemonRepository, logger: Logger) {
    this.pokemonRepository = pokemonRepository;
    this.logger = logger;
  }

  async execute(
    name: string,
    params: PaginationParams
  ): Promise<PaginatedResult<PokemonListItemDTO>> {
    const trimmedName = name.trim().toLowerCase();
    this.logger.info('SearchPokemonUseCase.execute', { name: trimmedName, ...params });

    if (trimmedName === '') {
      return {
        data: [],
        total: 0,
        page: params.page,
        limit: params.limit,
        totalPages: 0,
        hasNextPage: false,
        hasPreviousPage: false,
      };
    }

    const result = await this.pokemonRepository.searchByName(trimmedName, params);

    return {
      ...result,
      data: result.data.map((pokemon) => this.toDTO(pokemon)),
    };
  }

  private toDTO(pokemon: Pokemon): PokemonListItemDTO {
    return {
      id: pokemon.id,
      name: pokemon.name,
      types: pokemon.types.map((t) => t.value),
      spriteUrl: pokemon.sprites.getBestQuality(),
    };
  }
}
