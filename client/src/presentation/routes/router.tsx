import { createBrowserRouter } from 'react-router';

import { AppLayout } from '@/presentation/layout';
import { NotFoundPage } from '@/presentation/pages/NotFoundPage';
import { PokemonDetailPage } from '@/presentation/pages/PokemonDetailPage';
import { PokemonListPage } from '@/presentation/pages/PokemonListPage';

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
        element: <PokemonDetailPage />,
      },
      {
        path: '*',
        element: <NotFoundPage />,
      },
    ],
  },
]);
