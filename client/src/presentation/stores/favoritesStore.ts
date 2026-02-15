import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface FavoritesState {
  favoriteIds: number[];
  isFavorite: (id: number) => boolean;
  toggleFavorite: (id: number) => void;
  favoriteCount: () => number;
}

export const useFavoritesStore = create<FavoritesState>()(
  persist(
    (set, get) => ({
      favoriteIds: [],
      isFavorite: (id: number) => get().favoriteIds.includes(id),
      toggleFavorite: (id: number) => {
        set((state) => ({
          favoriteIds: state.favoriteIds.includes(id)
            ? state.favoriteIds.filter((fid) => fid !== id)
            : [...state.favoriteIds, id],
        }));
      },
      favoriteCount: () => get().favoriteIds.length,
    }),
    {
      name: 'pokemon-favorites',
    }
  )
);
