import { useCallback } from 'react';

import { Heart } from 'lucide-react';
import type { MouseEvent } from 'react';

import { cn } from '@/lib/utils';
import { Button } from '@/presentation/components/ui/button';
import { useFavoritesStore } from '@/presentation/stores';

interface FavoriteButtonProps {
  pokemonId: number;
  className?: string;
}

export function FavoriteButton({ pokemonId, className }: FavoriteButtonProps) {
  const isFavorite = useFavoritesStore((s) => s.isFavorite(pokemonId));
  const toggleFavorite = useFavoritesStore((s) => s.toggleFavorite);

  const handleClick = useCallback(
    (e: MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      toggleFavorite(pokemonId);
    },
    [pokemonId, toggleFavorite]
  );

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={handleClick}
      className={cn('h-8 w-8', className)}
      aria-label={isFavorite ? 'Remove from favorites' : 'Add to favorites'}
    >
      <Heart
        className={cn(
          'h-4 w-4 transition-colors',
          isFavorite ? 'fill-red-500 text-red-500' : 'text-muted-foreground'
        )}
      />
    </Button>
  );
}
