import type {
  PokeAPIPokemonResponse,
  PokeAPISpeciesResponse,
} from '@/infrastructure/api/pokeapi.types';
import { mapPokemonToDomain, mapSpeciesToDomain } from '@/infrastructure/mappers';

const mockPokemonResponse: PokeAPIPokemonResponse = {
  id: 1,
  name: 'bulbasaur',
  order: 1,
  base_experience: 64,
  height: 7,
  weight: 69,
  types: [
    { slot: 1, type: { name: 'grass', url: 'https://pokeapi.co/api/v2/type/12/' } },
    { slot: 2, type: { name: 'poison', url: 'https://pokeapi.co/api/v2/type/4/' } },
  ],
  stats: [
    { base_stat: 45, effort: 0, stat: { name: 'hp', url: '' } },
    { base_stat: 49, effort: 0, stat: { name: 'attack', url: '' } },
    { base_stat: 49, effort: 0, stat: { name: 'defense', url: '' } },
    { base_stat: 65, effort: 1, stat: { name: 'special-attack', url: '' } },
    { base_stat: 65, effort: 0, stat: { name: 'special-defense', url: '' } },
    { base_stat: 45, effort: 0, stat: { name: 'speed', url: '' } },
  ],
  sprites: {
    front_default: 'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/1.png',
    front_shiny:
      'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/shiny/1.png',
    back_default:
      'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/back/1.png',
    back_shiny: null,
    other: {
      'official-artwork': {
        front_default:
          'https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/1.png',
      },
    },
  },
};

const mockSpeciesResponse: PokeAPISpeciesResponse = {
  id: 1,
  name: 'bulbasaur',
  generation: { name: 'generation-i', url: '' },
  flavor_text_entries: [
    {
      flavor_text: 'A strange seed was\nplanted on its\fback at birth.',
      language: { name: 'en', url: '' },
      version: { name: 'red', url: '' },
    },
    {
      flavor_text: 'Bulbasaur texto en español.',
      language: { name: 'es', url: '' },
      version: { name: 'red', url: '' },
    },
  ],
  habitat: { name: 'grassland', url: '' },
  is_legendary: false,
  is_mythical: false,
};

describe('PokemonMapper', () => {
  describe('toDomain', () => {
    it('should map PokeAPI response to Pokemon entity', () => {
      const pokemon = mapPokemonToDomain(mockPokemonResponse);

      expect(pokemon.id).toBe(1);
      expect(pokemon.name).toBe('bulbasaur');
      expect(pokemon.types).toHaveLength(2);
      expect(pokemon.primaryType.value).toBe('grass');
      expect(pokemon.secondaryType?.value).toBe('poison');
      expect(pokemon.stats.hp).toBe(45);
      expect(pokemon.stats.attack).toBe(49);
      expect(pokemon.stats.specialAttack).toBe(65);
      expect(pokemon.height.value).toBe(7);
      expect(pokemon.height.unit).toBe('dm');
      expect(pokemon.weight.value).toBe(69);
      expect(pokemon.weight.unit).toBe('hg');
      expect(pokemon.order).toBe(1);
      expect(pokemon.baseExperience).toBe(64);
    });

    it('should map sprites including official artwork', () => {
      const pokemon = mapPokemonToDomain(mockPokemonResponse);

      expect(pokemon.sprites.frontDefault).toContain('1.png');
      expect(pokemon.sprites.officialArtwork).toContain('official-artwork');
    });

    it('should handle null official artwork', () => {
      const response: PokeAPIPokemonResponse = {
        ...mockPokemonResponse,
        sprites: {
          ...mockPokemonResponse.sprites,
          other: undefined,
        },
      };

      const pokemon = mapPokemonToDomain(response);
      expect(pokemon.sprites.officialArtwork).toBeNull();
    });

    it('should sort types by slot', () => {
      const response: PokeAPIPokemonResponse = {
        ...mockPokemonResponse,
        types: [
          { slot: 2, type: { name: 'poison', url: '' } },
          { slot: 1, type: { name: 'grass', url: '' } },
        ],
      };

      const pokemon = mapPokemonToDomain(response);
      expect(pokemon.primaryType.value).toBe('grass');
      expect(pokemon.secondaryType?.value).toBe('poison');
    });

    it('should handle null base_experience', () => {
      const response: PokeAPIPokemonResponse = {
        ...mockPokemonResponse,
        base_experience: null,
      };

      const pokemon = mapPokemonToDomain(response);
      expect(pokemon.baseExperience).toBeNull();
    });
  });

  describe('speciesToDomain', () => {
    it('should map species response to Species entity', () => {
      const species = mapSpeciesToDomain(mockSpeciesResponse);

      expect(species.id).toBe(1);
      expect(species.name).toBe('bulbasaur');
      expect(species.generation).toBe(1);
      expect(species.habitat).toBe('grassland');
      expect(species.isLegendary).toBe(false);
      expect(species.isMythical).toBe(false);
    });

    it('should extract English flavor text and clean whitespace', () => {
      const species = mapSpeciesToDomain(mockSpeciesResponse);

      expect(species.flavorText).toBe('A strange seed was planted on its back at birth.');
      expect(species.flavorText).not.toContain('\n');
      expect(species.flavorText).not.toContain('\f');
    });

    it('should handle null habitat', () => {
      const response: PokeAPISpeciesResponse = {
        ...mockSpeciesResponse,
        habitat: null,
      };

      const species = mapSpeciesToDomain(response);
      expect(species.habitat).toBeNull();
    });

    it('should map all generation names correctly', () => {
      const generations = [
        { name: 'generation-i', expected: 1 },
        { name: 'generation-v', expected: 5 },
        { name: 'generation-ix', expected: 9 },
      ];

      for (const gen of generations) {
        const response: PokeAPISpeciesResponse = {
          ...mockSpeciesResponse,
          generation: { name: gen.name, url: '' },
        };

        const species = mapSpeciesToDomain(response);
        expect(species.generation).toBe(gen.expected);
      }
    });

    it('should handle legendary Pokemon', () => {
      const response: PokeAPISpeciesResponse = {
        ...mockSpeciesResponse,
        is_legendary: true,
      };

      const species = mapSpeciesToDomain(response);
      expect(species.isLegendary).toBe(true);
      expect(species.isSpecial).toBe(true);
    });

    it('should return empty string when no English flavor text exists', () => {
      const response: PokeAPISpeciesResponse = {
        ...mockSpeciesResponse,
        flavor_text_entries: [
          {
            flavor_text: 'Solo en español',
            language: { name: 'es', url: '' },
            version: { name: 'red', url: '' },
          },
        ],
      };

      const species = mapSpeciesToDomain(response);
      expect(species.flavorText).toBe('');
    });
  });
});
