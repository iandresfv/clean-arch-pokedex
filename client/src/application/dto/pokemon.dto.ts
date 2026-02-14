import type { PokemonTypeName, StatsData } from '@/domain/pokemon';

export interface PokemonListItemDTO {
  id: number;
  name: string;
  types: PokemonTypeName[];
  spriteUrl: string | null;
}

export interface PokemonDetailDTO {
  id: number;
  name: string;
  types: PokemonTypeName[];
  stats: StatsData;
  height: string;
  weight: string;
  spriteUrl: string | null;
  officialArtworkUrl: string | null;
  sprites: {
    frontDefault: string | null;
    frontShiny: string | null;
    backDefault: string | null;
    backShiny: string | null;
  };
  baseExperience: number | null;
  species: SpeciesDTO | null;
}

export interface SpeciesDTO {
  id: number;
  name: string;
  generation: number;
  flavorText: string;
  habitat: string | null;
  isLegendary: boolean;
  isMythical: boolean;
}
