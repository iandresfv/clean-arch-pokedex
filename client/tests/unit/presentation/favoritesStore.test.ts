import { afterEach, beforeEach, describe, expect, it } from 'vitest';

import { useFavoritesStore } from '@/presentation/stores';

describe('favoritesStore', () => {
  beforeEach(() => {
    useFavoritesStore.setState({ favoriteIds: [] });
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('starts with empty favorites', () => {
    const { favoriteIds } = useFavoritesStore.getState();
    expect(favoriteIds).toEqual([]);
  });

  it('adds a pokemon to favorites', () => {
    useFavoritesStore.getState().toggleFavorite(25);
    expect(useFavoritesStore.getState().favoriteIds).toEqual([25]);
  });

  it('removes a pokemon from favorites when toggled again', () => {
    useFavoritesStore.getState().toggleFavorite(25);
    useFavoritesStore.getState().toggleFavorite(25);
    expect(useFavoritesStore.getState().favoriteIds).toEqual([]);
  });

  it('checks if a pokemon is favorite', () => {
    useFavoritesStore.getState().toggleFavorite(25);
    expect(useFavoritesStore.getState().isFavorite(25)).toBe(true);
    expect(useFavoritesStore.getState().isFavorite(1)).toBe(false);
  });

  it('returns correct favorite count', () => {
    useFavoritesStore.getState().toggleFavorite(25);
    useFavoritesStore.getState().toggleFavorite(1);
    useFavoritesStore.getState().toggleFavorite(6);
    expect(useFavoritesStore.getState().favoriteCount()).toBe(3);
  });

  it('handles multiple toggles correctly', () => {
    useFavoritesStore.getState().toggleFavorite(25);
    useFavoritesStore.getState().toggleFavorite(1);
    useFavoritesStore.getState().toggleFavorite(25);
    expect(useFavoritesStore.getState().favoriteIds).toEqual([1]);
    expect(useFavoritesStore.getState().isFavorite(25)).toBe(false);
    expect(useFavoritesStore.getState().isFavorite(1)).toBe(true);
  });
});
