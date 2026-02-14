import type { PokemonListItemDTO } from '@/application/dto';
import type { PaginatedResult, PaginationParams } from '@/application/types';

export interface SearchPokemonUseCase {
  execute(name: string, params: PaginationParams): Promise<PaginatedResult<PokemonListItemDTO>>;
}
