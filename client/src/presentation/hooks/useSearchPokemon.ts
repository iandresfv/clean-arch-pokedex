import { useQuery } from '@tanstack/react-query';

import type { PokemonListItemDTO } from '@/application/dto';
import type { PaginatedResult } from '@/application/types';
import { useDI } from '@/di/useDI';

export function useSearchPokemon(searchTerm: string) {
  const { searchPokemonUseCase } = useDI();

  return useQuery<PaginatedResult<PokemonListItemDTO>>({
    queryKey: ['pokemon', 'search', searchTerm],
    queryFn: () => searchPokemonUseCase.execute(searchTerm, { page: 1, limit: 50 }),
    enabled: searchTerm.trim().length > 0,
  });
}
