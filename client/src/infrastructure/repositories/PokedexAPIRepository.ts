import type {
  PokedexListItemResponse,
  PokedexPaginatedResponse,
  PokedexPokemonResponse,
  PokedexSpeciesResponse,
} from '../api/pokedex.types';
import type { HttpClient } from '../http/HttpClient';
import {
  mapPokedexListItemToDomain,
  mapPokedexPokemonToDomain,
  mapPokedexSpeciesToDomain,
} from '../mappers/PokedexMapper';

import type { PokemonRepository } from '@/application/ports';
import type { PaginatedResult, PaginationParams } from '@/application/types';
import type { Pokemon, Species } from '@/domain/pokemon';

/**
 * PokemonRepository backed by this project's own Go API.
 *
 * It implements the same port as PokeAPIRepository, unchanged. That is the
 * point of the port: swapping the data source touches this file and one line of
 * the composition root, while domain, application and presentation stay exactly
 * as they were.
 *
 * Failures surface as HttpError, which extends Error and carries the status and
 * the server's correlation id. That matches how PokeAPIRepository reports
 * failures, so the presentation layer needs no change.
 *
 * Compared with the PokeAPI adapter:
 *
 *   listing 20   1 request        vs 1 index request + 20 detail requests (N+1)
 *   search       1 request        vs downloading all 1302 names and filtering
 *   pagination   server-provided  vs computed on the client
 *   species      embedded         vs a separate request per Pokemon
 */
export class PokedexAPIRepository implements PokemonRepository {
  private readonly http: HttpClient;

  constructor(http: HttpClient) {
    this.http = http;
  }

  async getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    const query = new URLSearchParams({
      page: String(params.page),
      limit: String(params.limit),
    });

    const response = await this.http.getJSON<PokedexPaginatedResponse<PokedexListItemResponse>>(
      `/api/v1/pokemon?${query.toString()}`
    );

    return this.toPaginatedResult(response);
  }

  async searchByName(name: string, params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    const query = new URLSearchParams({
      q: name,
      page: String(params.page),
      limit: String(params.limit),
    });

    const response = await this.http.getJSON<PokedexPaginatedResponse<PokedexListItemResponse>>(
      `/api/v1/pokemon/search?${query.toString()}`
    );

    return this.toPaginatedResult(response);
  }

  async getById(id: number): Promise<Pokemon> {
    const response = await this.http.getJSON<PokedexPokemonResponse>(
      `/api/v1/pokemon/${String(id)}`
    );
    return mapPokedexPokemonToDomain(response);
  }

  async getSpeciesById(id: number): Promise<Species> {
    const response = await this.http.getJSON<PokedexSpeciesResponse>(
      `/api/v1/pokemon/${String(id)}/species`
    );
    return mapPokedexSpeciesToDomain(response);
  }

  /**
   * The server's envelope already uses the field names PaginatedResult<T>
   * expects, so this only maps the items.
   */
  private toPaginatedResult(
    response: PokedexPaginatedResponse<PokedexListItemResponse>
  ): PaginatedResult<Pokemon> {
    return {
      data: response.data.map(mapPokedexListItemToDomain),
      total: response.total,
      page: response.page,
      limit: response.limit,
      totalPages: response.totalPages,
      hasNextPage: response.hasNextPage,
      hasPreviousPage: response.hasPreviousPage,
    };
  }
}
