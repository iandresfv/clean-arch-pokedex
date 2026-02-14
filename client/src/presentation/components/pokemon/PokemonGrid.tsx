import { PokemonCard } from './PokemonCard';
import { PokemonCardSkeleton } from './PokemonCardSkeleton';

import type { PokemonListItemDTO } from '@/application/dto';

interface PokemonGridProps {
  pokemon: PokemonListItemDTO[];
  isLoading?: boolean;
  skeletonCount?: number;
}

export function PokemonGrid({ pokemon, isLoading, skeletonCount = 20 }: PokemonGridProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {Array.from({ length: skeletonCount }, (_, i) => (
          <PokemonCardSkeleton key={i} />
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
      {pokemon.map((p) => (
        <PokemonCard key={p.id} pokemon={p} />
      ))}
    </div>
  );
}
