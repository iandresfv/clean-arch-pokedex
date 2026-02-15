import { Heart } from 'lucide-react';
import { Link } from 'react-router';

import { Badge } from '@/presentation/components/ui/badge';
import { useFavoritesStore } from '@/presentation/stores';

export function Header() {
  const favoriteCount = useFavoritesStore((s) => s.favoriteIds.length);

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto flex h-14 max-w-screen-xl items-center justify-between px-4">
        <Link to="/" className="flex items-center gap-2">
          <div className="h-8 w-8 rounded-lg bg-red-600 flex items-center justify-center">
            <div className="h-4 w-4 rounded-full border-2 border-white bg-white/30" />
          </div>
          <span className="text-lg font-bold tracking-tight">Pokédex</span>
        </Link>
        {favoriteCount > 0 && (
          <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
            <Heart className="h-4 w-4 fill-red-500 text-red-500" />
            <Badge variant="secondary">{favoriteCount}</Badge>
          </div>
        )}
      </div>
    </header>
  );
}
