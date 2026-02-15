import { useState } from 'react';

import { PokemonGrid } from '@/presentation/components/pokemon/PokemonGrid';
import { SearchBar } from '@/presentation/components/pokemon/SearchBar';
import { EmptyState } from '@/presentation/components/shared/EmptyState';
import { ErrorState } from '@/presentation/components/shared/ErrorState';
import { Pagination } from '@/presentation/components/shared/Pagination';
import { useDebounce } from '@/presentation/hooks/useDebounce';
import { usePokemonList } from '@/presentation/hooks/usePokemonList';
import { useSearchPokemon } from '@/presentation/hooks/useSearchPokemon';

export function PokemonListPage() {
  const [searchInput, setSearchInput] = useState('');
  const debouncedSearch = useDebounce(searchInput, 300);
  const isSearching = debouncedSearch.trim().length > 0;

  const listQuery = usePokemonList();
  const searchQuery = useSearchPokemon(debouncedSearch);

  const activeQuery = isSearching ? searchQuery : listQuery;
  const { data, isLoading, isError } = activeQuery;

  const handleRetry = () => {
    void activeQuery.refetch();
  };

  return (
    <div className="animate-fade-in space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Pokédex</h1>
        <p className="text-muted-foreground">Browse and discover Pokemon from all generations.</p>
      </div>

      <SearchBar value={searchInput} onChange={setSearchInput} />

      {isError ? (
        <ErrorState message="Failed to load Pokemon. Please try again." onRetry={handleRetry} />
      ) : (
        <>
          {data?.data.length === 0 && !isLoading ? (
            <EmptyState
              title="No Pokemon found"
              description={`No results for "${debouncedSearch}". Try a different search term.`}
            />
          ) : (
            <PokemonGrid pokemon={data?.data ?? []} isLoading={isLoading} />
          )}
          {!isSearching && listQuery.data && (
            <Pagination
              page={listQuery.page}
              totalPages={listQuery.data.totalPages}
              hasNextPage={listQuery.data.hasNextPage}
              hasPreviousPage={listQuery.data.hasPreviousPage}
              onPageChange={listQuery.setPage}
            />
          )}
        </>
      )}
    </div>
  );
}
