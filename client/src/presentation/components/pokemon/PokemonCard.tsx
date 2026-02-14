import { Link } from 'react-router';

import { TypeBadge } from './TypeBadge';

import type { PokemonListItemDTO } from '@/application/dto';
import { Card, CardContent } from '@/presentation/components/ui/card';

interface PokemonCardProps {
  pokemon: PokemonListItemDTO;
}

export function PokemonCard({ pokemon }: PokemonCardProps) {
  return (
    <Link to={`/pokemon/${pokemon.id.toString()}`}>
      <Card className="group overflow-hidden transition-all hover:shadow-lg hover:-translate-y-1">
        <div className="relative flex items-center justify-center bg-muted/50 p-4">
          {pokemon.spriteUrl ? (
            <img
              src={pokemon.spriteUrl}
              alt={pokemon.name}
              className="h-24 w-24 object-contain transition-transform group-hover:scale-110"
              loading="lazy"
            />
          ) : (
            <div className="flex h-24 w-24 items-center justify-center text-muted-foreground">
              ?
            </div>
          )}
          <span className="absolute right-2 top-2 text-xs font-mono text-muted-foreground">
            #{pokemon.id.toString().padStart(3, '0')}
          </span>
        </div>
        <CardContent className="p-3">
          <h3 className="text-sm font-semibold capitalize truncate">{pokemon.name}</h3>
          <div className="mt-1.5 flex flex-wrap gap-1">
            {pokemon.types.map((type) => (
              <TypeBadge key={type} type={type} />
            ))}
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}
