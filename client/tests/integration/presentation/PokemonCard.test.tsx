import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';

import type { PokemonListItemDTO } from '@/application/dto';
import { PokemonCard } from '@/presentation/components/pokemon/PokemonCard';

const mockPokemon: PokemonListItemDTO = {
  id: 25,
  name: 'pikachu',
  types: ['electric'],
  spriteUrl: 'https://example.com/pikachu.png',
};

function renderWithRouter(pokemon: PokemonListItemDTO) {
  return render(
    <MemoryRouter>
      <PokemonCard pokemon={pokemon} />
    </MemoryRouter>
  );
}

describe('PokemonCard', () => {
  it('renders pokemon name', () => {
    renderWithRouter(mockPokemon);
    expect(screen.getByText('pikachu')).toBeInTheDocument();
  });

  it('renders pokemon id formatted with leading zeros', () => {
    renderWithRouter(mockPokemon);
    expect(screen.getByText('#025')).toBeInTheDocument();
  });

  it('renders pokemon sprite image', () => {
    renderWithRouter(mockPokemon);
    const img = screen.getByAltText('pikachu');
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src', 'https://example.com/pikachu.png');
  });

  it('renders type badges', () => {
    renderWithRouter(mockPokemon);
    expect(screen.getByText('electric')).toBeInTheDocument();
  });

  it('renders multiple type badges', () => {
    const dualTypePokemon: PokemonListItemDTO = {
      id: 6,
      name: 'charizard',
      types: ['fire', 'flying'],
      spriteUrl: null,
    };
    renderWithRouter(dualTypePokemon);
    expect(screen.getByText('fire')).toBeInTheDocument();
    expect(screen.getByText('flying')).toBeInTheDocument();
  });

  it('renders placeholder when sprite is null', () => {
    const noSpritePokemon: PokemonListItemDTO = {
      id: 1,
      name: 'bulbasaur',
      types: ['grass', 'poison'],
      spriteUrl: null,
    };
    renderWithRouter(noSpritePokemon);
    expect(screen.getByText('?')).toBeInTheDocument();
  });

  it('links to pokemon detail page', () => {
    renderWithRouter(mockPokemon);
    const link = screen.getByRole('link');
    expect(link).toHaveAttribute('href', '/pokemon/25');
  });
});
