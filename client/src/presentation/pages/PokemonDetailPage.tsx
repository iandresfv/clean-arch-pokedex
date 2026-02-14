import { useParams } from 'react-router';

export function PokemonDetailPage() {
  const { id } = useParams<{ id: string }>();

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold tracking-tight">Pokemon #{id}</h1>
      <p className="text-muted-foreground">Pokemon detail page — coming soon.</p>
    </div>
  );
}
