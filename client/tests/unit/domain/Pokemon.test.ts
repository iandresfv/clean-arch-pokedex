import { InvalidPokemonEntityError } from '@/domain/errors';
import type { PokemonProps } from '@/domain/pokemon';
import { Pokemon } from '@/domain/pokemon';

const validProps: PokemonProps = {
  id: 1,
  name: 'Bulbasaur',
  types: ['grass', 'poison'],
  stats: { hp: 45, attack: 49, defense: 49, specialAttack: 65, specialDefense: 65, speed: 45 },
  height: { value: 7, unit: 'dm' },
  weight: { value: 69, unit: 'hg' },
  sprites: {
    frontDefault: 'https://example.com/front.png',
    frontShiny: 'https://example.com/shiny.png',
    backDefault: null,
    backShiny: null,
    officialArtwork: 'https://example.com/official.png',
  },
  order: 1,
  baseExperience: 64,
};

describe('Pokemon Entity', () => {
  describe('create', () => {
    it('should create a valid Pokemon entity', () => {
      const pokemon = Pokemon.create(validProps);

      expect(pokemon.id).toBe(1);
      expect(pokemon.name).toBe('Bulbasaur');
      expect(pokemon.types).toHaveLength(2);
      expect(pokemon.primaryType.value).toBe('grass');
      expect(pokemon.secondaryType?.value).toBe('poison');
      expect(pokemon.isDualType).toBe(true);
      expect(pokemon.stats.hp).toBe(45);
      expect(pokemon.height.value).toBe(7);
      expect(pokemon.weight.value).toBe(69);
      expect(pokemon.order).toBe(1);
      expect(pokemon.baseExperience).toBe(64);
    });

    it('should create a single-type Pokemon', () => {
      const pokemon = Pokemon.create({ ...validProps, types: ['fire'] });

      expect(pokemon.types).toHaveLength(1);
      expect(pokemon.primaryType.value).toBe('fire');
      expect(pokemon.secondaryType).toBeNull();
      expect(pokemon.isDualType).toBe(false);
    });

    it('should trim the Pokemon name', () => {
      const pokemon = Pokemon.create({ ...validProps, name: '  Bulbasaur  ' });
      expect(pokemon.name).toBe('Bulbasaur');
    });

    it('should allow null baseExperience', () => {
      const pokemon = Pokemon.create({ ...validProps, baseExperience: null });
      expect(pokemon.baseExperience).toBeNull();
    });

    it('should freeze the types array', () => {
      const pokemon = Pokemon.create(validProps);
      expect(Object.isFrozen(pokemon.types)).toBe(true);
    });
  });

  describe('validation', () => {
    it('should reject id of 0', () => {
      expect(() => Pokemon.create({ ...validProps, id: 0 })).toThrow(InvalidPokemonEntityError);
    });

    it('should reject negative id', () => {
      expect(() => Pokemon.create({ ...validProps, id: -1 })).toThrow(InvalidPokemonEntityError);
    });

    it('should reject non-integer id', () => {
      expect(() => Pokemon.create({ ...validProps, id: 1.5 })).toThrow(InvalidPokemonEntityError);
    });

    it('should reject empty name', () => {
      expect(() => Pokemon.create({ ...validProps, name: '' })).toThrow(InvalidPokemonEntityError);
    });

    it('should reject whitespace-only name', () => {
      expect(() => Pokemon.create({ ...validProps, name: '   ' })).toThrow(
        InvalidPokemonEntityError
      );
    });

    it('should reject empty types array', () => {
      expect(() => Pokemon.create({ ...validProps, types: [] })).toThrow(InvalidPokemonEntityError);
    });

    it('should reject more than two types', () => {
      expect(() => Pokemon.create({ ...validProps, types: ['fire', 'water', 'grass'] })).toThrow(
        InvalidPokemonEntityError
      );
    });

    it('should reject non-integer order', () => {
      expect(() => Pokemon.create({ ...validProps, order: 1.5 })).toThrow(
        InvalidPokemonEntityError
      );
    });
  });

  describe('hasType', () => {
    it('should return true for a type the Pokemon has', () => {
      const pokemon = Pokemon.create(validProps);
      expect(pokemon.hasType('grass')).toBe(true);
      expect(pokemon.hasType('poison')).toBe(true);
    });

    it('should return false for a type the Pokemon does not have', () => {
      const pokemon = Pokemon.create(validProps);
      expect(pokemon.hasType('fire')).toBe(false);
    });

    it('should be case-insensitive', () => {
      const pokemon = Pokemon.create(validProps);
      expect(pokemon.hasType('GRASS')).toBe(true);
      expect(pokemon.hasType('Poison')).toBe(true);
    });
  });

  describe('equals', () => {
    it('should return true for Pokemon with the same id', () => {
      const pokemon1 = Pokemon.create(validProps);
      const pokemon2 = Pokemon.create({ ...validProps, name: 'Different' });
      expect(pokemon1.equals(pokemon2)).toBe(true);
    });

    it('should return false for Pokemon with different ids', () => {
      const pokemon1 = Pokemon.create(validProps);
      const pokemon2 = Pokemon.create({ ...validProps, id: 2 });
      expect(pokemon1.equals(pokemon2)).toBe(false);
    });
  });

  describe('sprites access', () => {
    it('should expose sprites value object', () => {
      const pokemon = Pokemon.create(validProps);
      expect(pokemon.sprites.getBestQuality()).toBe('https://example.com/official.png');
    });
  });
});
