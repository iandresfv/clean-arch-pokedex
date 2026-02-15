import { ArrowLeft } from 'lucide-react';
import { Link, useParams } from 'react-router';

import { FavoriteButton } from '@/presentation/components/pokemon/FavoriteButton';
import { PokemonDetailHeader } from '@/presentation/components/pokemon/PokemonDetailHeader';
import { PokemonDetailSkeleton } from '@/presentation/components/pokemon/PokemonDetailSkeleton';
import { PokemonSprites } from '@/presentation/components/pokemon/PokemonSprites';
import { SpeciesInfo } from '@/presentation/components/pokemon/SpeciesInfo';
import { StatsChart } from '@/presentation/components/pokemon/StatsChart';
import { ErrorState } from '@/presentation/components/shared/ErrorState';
import { buttonVariants } from '@/presentation/components/ui/button';
import { usePokemonDetail } from '@/presentation/hooks/usePokemonDetail';

export function PokemonDetailPage() {
  const { id } = useParams<{ id: string }>();
  const pokemonId = Number(id);
  const { data: pokemon, isLoading, isError, refetch } = usePokemonDetail(pokemonId);

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <Link to="/" className={buttonVariants({ variant: 'ghost', size: 'sm' })}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to list
        </Link>
        {pokemon && <FavoriteButton pokemonId={pokemonId} />}
      </div>

      {isLoading && <PokemonDetailSkeleton />}

      {isError && (
        <ErrorState
          message="Failed to load Pokemon details. Please try again."
          onRetry={() => {
            void refetch();
          }}
        />
      )}

      {pokemon && (
        <>
          <PokemonDetailHeader pokemon={pokemon} />
          <div className="grid gap-8 md:grid-cols-2">
            <StatsChart stats={pokemon.stats} />
            <div className="space-y-8">
              {pokemon.species && <SpeciesInfo species={pokemon.species} />}
              <PokemonSprites sprites={pokemon.sprites} name={pokemon.name} />
            </div>
          </div>
        </>
      )}
    </div>
  );
}
