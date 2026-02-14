import { Card, CardContent } from '@/presentation/components/ui/card';
import { Skeleton } from '@/presentation/components/ui/skeleton';

export function PokemonCardSkeleton() {
  return (
    <Card className="overflow-hidden">
      <div className="flex items-center justify-center bg-muted/50 p-4">
        <Skeleton className="h-24 w-24 rounded-lg" />
      </div>
      <CardContent className="p-3">
        <Skeleton className="h-4 w-20" />
        <div className="mt-1.5 flex gap-1">
          <Skeleton className="h-5 w-14 rounded-full" />
          <Skeleton className="h-5 w-14 rounded-full" />
        </div>
      </CardContent>
    </Card>
  );
}
