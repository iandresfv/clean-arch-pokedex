import type { CacheService, Logger, PokemonRepository } from '@/application/ports';
import type { GetPokemonByIdUseCase } from '@/application/use-cases/get-pokemon-by-id';
import { GetPokemonByIdUseCaseImpl } from '@/application/use-cases/get-pokemon-by-id';
import type { ListPokemonUseCase } from '@/application/use-cases/list-pokemon';
import { ListPokemonUseCaseImpl } from '@/application/use-cases/list-pokemon';
import type { SearchPokemonUseCase } from '@/application/use-cases/search-pokemon';
import { SearchPokemonUseCaseImpl } from '@/application/use-cases/search-pokemon';
import { LocalStorageCacheService } from '@/infrastructure/cache/LocalStorageCacheService';
import { ConsoleLogger } from '@/infrastructure/logger/ConsoleLogger';
import { PokeAPIRepository } from '@/infrastructure/repositories/PokeAPIRepository';

export interface DIContainer {
  pokemonRepository: PokemonRepository;
  cacheService: CacheService;
  logger: Logger;
  listPokemonUseCase: ListPokemonUseCase;
  getPokemonByIdUseCase: GetPokemonByIdUseCase;
  searchPokemonUseCase: SearchPokemonUseCase;
}

export function createContainer(): DIContainer {
  const logger = new ConsoleLogger();
  const cacheService = new LocalStorageCacheService();
  const pokemonRepository = new PokeAPIRepository();

  const listPokemonUseCase = new ListPokemonUseCaseImpl(pokemonRepository, logger);
  const getPokemonByIdUseCase = new GetPokemonByIdUseCaseImpl(pokemonRepository, logger);
  const searchPokemonUseCase = new SearchPokemonUseCaseImpl(pokemonRepository, logger);

  return {
    pokemonRepository,
    cacheService,
    logger,
    listPokemonUseCase,
    getPokemonByIdUseCase,
    searchPokemonUseCase,
  };
}
