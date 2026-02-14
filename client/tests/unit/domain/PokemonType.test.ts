import { describe, expect, it } from 'vitest';

import { InvalidPokemonTypeError } from '@/domain/errors';
import { PokemonType, VALID_POKEMON_TYPES } from '@/domain/pokemon';

describe('PokemonType', () => {
  describe('constructor', () => {
    it('should create valid Pokemon type', () => {
      const type = new PokemonType('fire');

      expect(type.value).toBe('fire');
    });

    it('should create type from all valid types', () => {
      VALID_POKEMON_TYPES.forEach((validType) => {
        const type = new PokemonType(validType);
        expect(type.value).toBe(validType);
      });
    });

    it('should normalize type to lowercase', () => {
      const type = new PokemonType('FIRE');

      expect(type.value).toBe('fire');
    });

    it('should trim whitespace', () => {
      const type = new PokemonType('  water  ');

      expect(type.value).toBe('water');
    });

    it('should throw InvalidPokemonTypeError for invalid type', () => {
      expect(() => new PokemonType('banana')).toThrow(InvalidPokemonTypeError);
    });

    it('should throw for empty string', () => {
      expect(() => new PokemonType('')).toThrow(InvalidPokemonTypeError);
    });

    it('should throw for numeric string', () => {
      expect(() => new PokemonType('123')).toThrow(InvalidPokemonTypeError);
    });
  });

  describe('equals', () => {
    it('should return true for same type', () => {
      const type1 = new PokemonType('fire');
      const type2 = new PokemonType('fire');

      expect(type1.equals(type2)).toBe(true);
    });

    it('should return true regardless of input case', () => {
      const type1 = new PokemonType('FIRE');
      const type2 = new PokemonType('fire');

      expect(type1.equals(type2)).toBe(true);
    });

    it('should return false for different types', () => {
      const type1 = new PokemonType('fire');
      const type2 = new PokemonType('water');

      expect(type1.equals(type2)).toBe(false);
    });
  });

  describe('toString', () => {
    it('should return type as string', () => {
      const type = new PokemonType('electric');

      expect(type.toString()).toBe('electric');
    });
  });
});
