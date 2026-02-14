import type { ListPokemonUseCase } from './ListPokemonUseCase';

import type { PokemonListItemDTO } from '@/application/dto';
import type { Logger, PokemonRepository } from '@/application/ports';
import type { PaginatedResult, PaginationParams } from '@/application/types';
import type { Pokemon } from '@/domain/pokemon';

export class ListPokemonUseCaseImpl implements ListPokemonUseCase {
  private readonly pokemonRepository: PokemonRepository;
  private readonly logger: Logger;

  constructor(pokemonRepository: PokemonRepository, logger: Logger) {
    this.pokemonRepository = pokemonRepository;
    this.logger = logger;
  }

  async execute(params: PaginationParams): Promise<PaginatedResult<PokemonListItemDTO>> {
    this.logger.info('ListPokemonUseCase.execute', { page: params.page, limit: params.limit });

    const result = await this.pokemonRepository.getAll(params);

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
