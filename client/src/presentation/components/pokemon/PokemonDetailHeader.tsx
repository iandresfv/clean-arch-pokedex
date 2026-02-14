import { TypeBadge } from './TypeBadge';

import type { PokemonDetailDTO } from '@/application/dto';

interface PokemonDetailHeaderProps {
  pokemon: PokemonDetailDTO;
}

export function PokemonDetailHeader({ pokemon }: PokemonDetailHeaderProps) {
  const mainImage = pokemon.officialArtworkUrl ?? pokemon.spriteUrl;

  return (
    <div className="flex flex-col items-center gap-4 sm:flex-row sm:items-start sm:gap-8">
      <div className="flex-shrink-0 rounded-2xl bg-muted/50 p-6">
        {mainImage ? (
          <img src={mainImage} alt={pokemon.name} className="h-48 w-48 object-contain" />
        ) : (
          <div className="flex h-48 w-48 items-center justify-center text-4xl text-muted-foreground">
            ?
          </div>
        )}
      </div>
      <div className="flex flex-col items-center gap-3 sm:items-start">
        <span className="text-sm font-mono text-muted-foreground">
          #{pokemon.id.toString().padStart(3, '0')}
        </span>
        <h1 className="text-3xl font-bold capitalize tracking-tight">{pokemon.name}</h1>
        <div className="flex flex-wrap gap-2">
          {pokemon.types.map((type) => (
            <TypeBadge key={type} type={type} />
          ))}
        </div>
        <div className="mt-2 flex gap-6 text-sm text-muted-foreground">
          <div>
            <span className="font-medium text-foreground">Height:</span> {pokemon.height}
          </div>
          <div>
            <span className="font-medium text-foreground">Weight:</span> {pokemon.weight}
          </div>
          {pokemon.baseExperience !== null && (
            <div>
              <span className="font-medium text-foreground">Base XP:</span> {pokemon.baseExperience}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
