import type { CacheService, Logger, PokemonRepository } from '@/application/ports';
import type { GetPokemonByIdUseCase } from '@/application/use-cases/get-pokemon-by-id';
import { GetPokemonByIdUseCaseImpl } from '@/application/use-cases/get-pokemon-by-id';
import type { ListPokemonUseCase } from '@/application/use-cases/list-pokemon';
import { ListPokemonUseCaseImpl } from '@/application/use-cases/list-pokemon';
import type { SearchPokemonUseCase } from '@/application/use-cases/search-pokemon';
import { SearchPokemonUseCaseImpl } from '@/application/use-cases/search-pokemon';
import { env } from '@/config/env';
import { LocalStorageCacheService } from '@/infrastructure/cache/LocalStorageCacheService';
import { HttpClient } from '@/infrastructure/http/HttpClient';
import { ConsoleLogger } from '@/infrastructure/logger/ConsoleLogger';
import { PokeAPIRepository } from '@/infrastructure/repositories/PokeAPIRepository';
import { PokedexAPIRepository } from '@/infrastructure/repositories/PokedexAPIRepository';

export interface DIContainer {
  pokemonRepository: PokemonRepository;
  cacheService: CacheService;
  logger: Logger;
  listPokemonUseCase: ListPokemonUseCase;
  getPokemonByIdUseCase: GetPokemonByIdUseCase;
  searchPokemonUseCase: SearchPokemonUseCase;
}

/**
 * Selects the data source.
 *
 * Both adapters implement PokemonRepository, so this is the only place in the
 * codebase that knows which one is in use. Everything above depends on the
 * port: no use case, hook, component or domain entity changes when this line
 * does — which is the whole argument for the architecture.
 *
 * PokeAPIRepository is kept rather than deleted so the switch is reversible and
 * the E2E suite has a fallback when the API is not running.
 */
function createPokemonRepository(): PokemonRepository {
  if (env.dataSource === 'pokeapi') {
    return new PokeAPIRepository();
  }
  return new PokedexAPIRepository(new HttpClient(env.apiUrl, env.requestTimeoutMs));
}

export function createContainer(): DIContainer {
  const logger = new ConsoleLogger();
  const cacheService = new LocalStorageCacheService();
  const pokemonRepository = createPokemonRepository();

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
