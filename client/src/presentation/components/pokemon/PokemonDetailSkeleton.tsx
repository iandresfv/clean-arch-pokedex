import { Skeleton } from '@/presentation/components/ui/skeleton';

export function PokemonDetailSkeleton() {
  return (
    <div className="space-y-8">
      <div className="flex flex-col items-center gap-4 sm:flex-row sm:items-start sm:gap-8">
        <Skeleton className="h-60 w-60 rounded-2xl" />
        <div className="flex flex-col items-center gap-3 sm:items-start">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-8 w-40" />
          <div className="flex gap-2">
            <Skeleton className="h-6 w-16 rounded-full" />
            <Skeleton className="h-6 w-16 rounded-full" />
          </div>
          <Skeleton className="mt-2 h-4 w-48" />
        </div>
      </div>
      <div className="space-y-3">
        <Skeleton className="h-6 w-24" />
        {Array.from({ length: 6 }, (_, i) => (
          <div key={i} className="flex items-center gap-3">
            <Skeleton className="h-4 w-16" />
            <Skeleton className="h-4 w-8" />
            <Skeleton className="h-2 flex-1 rounded-full" />
          </div>
        ))}
      </div>
    </div>
  );
}
