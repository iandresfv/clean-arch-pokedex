import { useState } from 'react';

import { useQuery } from '@tanstack/react-query';

import type { PokemonListItemDTO } from '@/application/dto';
import type { PaginatedResult } from '@/application/types';
import { useDI } from '@/di/useDI';

const DEFAULT_LIMIT = 20;

export function usePokemonList() {
  const { listPokemonUseCase } = useDI();
  const [page, setPage] = useState(1);

  const query = useQuery<PaginatedResult<PokemonListItemDTO>>({
    queryKey: ['pokemon', 'list', page],
    queryFn: () => listPokemonUseCase.execute({ page, limit: DEFAULT_LIMIT }),
  });

  return {
    ...query,
    page,
    setPage,
  };
}
