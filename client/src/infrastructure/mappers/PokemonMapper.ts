import type { PokeAPIPokemonResponse, PokeAPISpeciesResponse } from '../api/pokeapi.types';

import type { PokemonProps, SpeciesProps } from '@/domain/pokemon';
import { Pokemon, Species } from '@/domain/pokemon';

const GENERATION_MAP: Record<string, number> = {
  'generation-i': 1,
  'generation-ii': 2,
  'generation-iii': 3,
  'generation-iv': 4,
  'generation-v': 5,
  'generation-vi': 6,
  'generation-vii': 7,
  'generation-viii': 8,
  'generation-ix': 9,
};

function mapStats(response: PokeAPIPokemonResponse): PokemonProps['stats'] {
  let hp = 0;
  let attack = 0;
  let defense = 0;
  let specialAttack = 0;
  let specialDefense = 0;
  let speed = 0;

  for (const stat of response.stats) {
    switch (stat.stat.name) {
      case 'hp':
        hp = stat.base_stat;
        break;
      case 'attack':
        attack = stat.base_stat;
        break;
      case 'defense':
        defense = stat.base_stat;
        break;
      case 'special-attack':
        specialAttack = stat.base_stat;
        break;
      case 'special-defense':
        specialDefense = stat.base_stat;
        break;
      case 'speed':
        speed = stat.base_stat;
        break;
    }
  }

  return { hp, attack, defense, specialAttack, specialDefense, speed };
}

function getEnglishFlavorText(response: PokeAPISpeciesResponse): string {
  const entry = response.flavor_text_entries.find((e) => e.language.name === 'en');

  if (!entry) {
    return '';
  }

  // PokeAPI flavor texts contain form feed (\f) and newline characters
  return entry.flavor_text
    .replace(/[\f\n\r]/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();
}

export function mapPokemonToDomain(response: PokeAPIPokemonResponse): Pokemon {
  const props: PokemonProps = {
    id: response.id,
    name: response.name,
    types: response.types.sort((a, b) => a.slot - b.slot).map((t) => t.type.name),
    stats: mapStats(response),
    height: { value: response.height, unit: 'dm' },
    weight: { value: response.weight, unit: 'hg' },
    sprites: {
      frontDefault: response.sprites.front_default,
      frontShiny: response.sprites.front_shiny,
      backDefault: response.sprites.back_default,
      backShiny: response.sprites.back_shiny,
      officialArtwork: response.sprites.other?.['official-artwork']?.front_default ?? null,
    },
    order: response.order,
    baseExperience: response.base_experience,
  };

  return Pokemon.create(props);
}

export function mapSpeciesToDomain(response: PokeAPISpeciesResponse): Species {
  const props: SpeciesProps = {
    id: response.id,
    name: response.name,
    generation: GENERATION_MAP[response.generation.name] ?? 1,
    flavorText: getEnglishFlavorText(response),
    habitat: response.habitat?.name ?? null,
    isLegendary: response.is_legendary,
    isMythical: response.is_mythical,
  };

  return Species.create(props);
}
