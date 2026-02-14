import { PokemonGrid } from '@/presentation/components/pokemon/PokemonGrid';
import { ErrorState } from '@/presentation/components/shared/ErrorState';
import { Pagination } from '@/presentation/components/shared/Pagination';
import { usePokemonList } from '@/presentation/hooks/usePokemonList';

export function PokemonListPage() {
  const { data, isLoading, isError, refetch, page, setPage } = usePokemonList();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Pokédex</h1>
        <p className="text-muted-foreground">Browse and discover Pokemon from all generations.</p>
      </div>

      {isError ? (
        <ErrorState
          message="Failed to load Pokemon. Please try again."
          onRetry={() => {
            void refetch();
          }}
        />
      ) : (
        <>
          <PokemonGrid pokemon={data?.data ?? []} isLoading={isLoading} />
          {data && (
            <Pagination
              page={page}
              totalPages={data.totalPages}
              hasNextPage={data.hasNextPage}
              hasPreviousPage={data.hasPreviousPage}
              onPageChange={setPage}
            />
          )}
        </>
      )}
    </div>
  );
}
