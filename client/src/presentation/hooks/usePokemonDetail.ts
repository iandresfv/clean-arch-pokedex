import { useQuery } from '@tanstack/react-query';

import type { PokemonDetailDTO } from '@/application/dto';
import { useDI } from '@/di/useDI';

export function usePokemonDetail(id: number) {
  const { getPokemonByIdUseCase } = useDI();

  return useQuery<PokemonDetailDTO>({
    queryKey: ['pokemon', 'detail', id],
    queryFn: () => getPokemonByIdUseCase.execute(id),
    enabled: id > 0,
  });
}
