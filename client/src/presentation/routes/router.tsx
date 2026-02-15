import { lazy, Suspense } from 'react';

import { createBrowserRouter } from 'react-router';

import { PokemonDetailSkeleton } from '@/presentation/components/pokemon/PokemonDetailSkeleton';
import { AppLayout } from '@/presentation/layout';
import { NotFoundPage } from '@/presentation/pages/NotFoundPage';
import { PokemonListPage } from '@/presentation/pages/PokemonListPage';

const PokemonDetailPage = lazy(() =>
  import('@/presentation/pages/PokemonDetailPage').then((m) => ({ default: m.PokemonDetailPage }))
);

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      {
        index: true,
        element: <PokemonListPage />,
      },
      {
        path: 'pokemon/:id',
        element: (
          <Suspense fallback={<PokemonDetailSkeleton />}>
            <PokemonDetailPage />
          </Suspense>
        ),
      },
      {
        path: '*',
        element: <NotFoundPage />,
      },
    ],
  },
]);
