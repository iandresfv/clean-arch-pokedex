import type {
  PokedexListItemResponse,
  PokedexPokemonResponse,
  PokedexSpeciesResponse,
} from '../api/pokedex.types';

import type { PokemonProps, SpeciesProps } from '@/domain/pokemon';
import { Pokemon, Species } from '@/domain/pokemon';

/**
 * Maps Pokedex API responses onto domain entities.
 *
 * Noticeably smaller than PokemonMapper: the server already returns camelCase
 * fields, resolved stats, ordered types and a numeric generation, so there is
 * no renaming, no stat pivoting, no roman-numeral parsing and no flavour-text
 * cleanup left to do on the client. Those transformations did not disappear —
 * they moved to the seeder, where they run once instead of on every render.
 */

/** Height and weight stay in PokeAPI's raw units; the domain value object formats them. */
function toPokemonProps(response: PokedexPokemonResponse | PokedexListItemResponse): PokemonProps {
  return {
    id: response.id,
    name: response.name,
    // Already ordered by slot server-side, so no client-side sort is needed.
    types: response.types,
    stats: response.stats,
    height: { value: response.heightDm, unit: 'dm' },
    weight: { value: response.weightHg, unit: 'hg' },
    sprites: {
      frontDefault: response.sprites.frontDefault,
      frontShiny: response.sprites.frontShiny,
      backDefault: response.sprites.backDefault,
      backShiny: response.sprites.backShiny,
      officialArtwork: response.sprites.officialArtwork,
    },
    order: response.pokedexOrder,
    baseExperience: response.baseExperience,
  };
}

export function mapPokedexPokemonToDomain(response: PokedexPokemonResponse): Pokemon {
  return Pokemon.create(toPokemonProps(response));
}

export function mapPokedexListItemToDomain(response: PokedexListItemResponse): Pokemon {
  return Pokemon.create(toPokemonProps(response));
}

export function mapPokedexSpeciesToDomain(response: PokedexSpeciesResponse): Species {
  const props: SpeciesProps = {
    id: response.id,
    name: response.name,
    generation: response.generation,
    flavorText: response.flavorText,
    habitat: response.habitat,
    isLegendary: response.isLegendary,
    isMythical: response.isMythical,
  };

  return Species.create(props);
}
