import type { PokemonListItemDTO } from '@/application/dto';
import type { PaginatedResult, PaginationParams } from '@/application/types';

export interface ListPokemonUseCase {
  execute(params: PaginationParams): Promise<PaginatedResult<PokemonListItemDTO>>;
}
