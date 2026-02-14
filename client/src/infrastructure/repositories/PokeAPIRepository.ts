import type {
  PokeAPIListResponse,
  PokeAPIPokemonResponse,
  PokeAPISpeciesResponse,
} from '../api/pokeapi.types';
import { mapPokemonToDomain, mapSpeciesToDomain } from '../mappers/PokemonMapper';

import type { PokemonRepository } from '@/application/ports';
import type { PaginatedResult, PaginationParams } from '@/application/types';
import type { Pokemon, Species } from '@/domain/pokemon';

const BASE_URL = 'https://pokeapi.co/api/v2';

export class PokeAPIRepository implements PokemonRepository {
  async getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    const offset = (params.page - 1) * params.limit;
    const listResponse = await this.fetchJSON<PokeAPIListResponse>(
      `${BASE_URL}/pokemon?limit=${String(params.limit)}&offset=${String(offset)}`
    );

    const pokemonPromises = listResponse.results.map((item) =>
      this.fetchJSON<PokeAPIPokemonResponse>(item.url).then(mapPokemonToDomain)
    );

    const data = await Promise.all(pokemonPromises);
    const totalPages = Math.ceil(listResponse.count / params.limit);

    return {
      data,
      total: listResponse.count,
      page: params.page,
      limit: params.limit,
      totalPages,
      hasNextPage: params.page < totalPages,
      hasPreviousPage: params.page > 1,
    };
  }

  async getById(id: number): Promise<Pokemon> {
    const response = await this.fetchJSON<PokeAPIPokemonResponse>(
      `${BASE_URL}/pokemon/${String(id)}`
    );
    return mapPokemonToDomain(response);
  }

  async getSpeciesById(id: number): Promise<Species> {
    const response = await this.fetchJSON<PokeAPISpeciesResponse>(
      `${BASE_URL}/pokemon-species/${String(id)}`
    );
    return mapSpeciesToDomain(response);
  }

  async searchByName(name: string, params: PaginationParams): Promise<PaginatedResult<Pokemon>> {
    // PokeAPI doesn't support search — fetch all and filter client-side
    // For a real app, this would be handled by a backend with proper search
    const allResponse = await this.fetchJSON<PokeAPIListResponse>(
      `${BASE_URL}/pokemon?limit=1302&offset=0`
    );

    const matchingItems = allResponse.results.filter((item) => item.name.includes(name));

    const totalMatches = matchingItems.length;
    const offset = (params.page - 1) * params.limit;
    const pageItems = matchingItems.slice(offset, offset + params.limit);

    const pokemonPromises = pageItems.map((item) =>
      this.fetchJSON<PokeAPIPokemonResponse>(item.url).then(mapPokemonToDomain)
    );

    const data = await Promise.all(pokemonPromises);
    const totalPages = Math.ceil(totalMatches / params.limit);

    return {
      data,
      total: totalMatches,
      page: params.page,
      limit: params.limit,
      totalPages,
      hasNextPage: params.page < totalPages,
      hasPreviousPage: params.page > 1,
    };
  }

  private async fetchJSON<T>(url: string): Promise<T> {
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error(`PokeAPI request failed: ${String(response.status)} ${response.statusText}`);
    }

    return response.json() as Promise<T>;
  }
}
