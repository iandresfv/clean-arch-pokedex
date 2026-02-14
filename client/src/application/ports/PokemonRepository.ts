import type { PaginatedResult, PaginationParams } from '../types/pagination.types';

import type { Pokemon, Species } from '@/domain/pokemon';

export interface PokemonRepository {
  getAll(params: PaginationParams): Promise<PaginatedResult<Pokemon>>;
  getById(id: number): Promise<Pokemon>;
  getSpeciesById(id: number): Promise<Species>;
  searchByName(name: string, params: PaginationParams): Promise<PaginatedResult<Pokemon>>;
}
