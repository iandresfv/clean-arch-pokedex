import { InvalidSpeciesError } from '@/domain/errors';
import type { SpeciesProps } from '@/domain/pokemon';
import { Species } from '@/domain/pokemon';

const validProps: SpeciesProps = {
  id: 1,
  name: 'bulbasaur',
  generation: 1,
  flavorText: 'A strange seed was planted on its back at birth.',
  habitat: 'grassland',
  isLegendary: false,
  isMythical: false,
};

describe('Species Entity', () => {
  describe('create', () => {
    it('should create a valid Species entity', () => {
      const species = Species.create(validProps);

      expect(species.id).toBe(1);
      expect(species.name).toBe('bulbasaur');
      expect(species.generation).toBe(1);
      expect(species.flavorText).toBe('A strange seed was planted on its back at birth.');
      expect(species.habitat).toBe('grassland');
      expect(species.isLegendary).toBe(false);
      expect(species.isMythical).toBe(false);
      expect(species.isSpecial).toBe(false);
    });

    it('should create a legendary Species', () => {
      const species = Species.create({
        ...validProps,
        id: 150,
        name: 'mewtwo',
        generation: 1,
        isLegendary: true,
      });

      expect(species.isLegendary).toBe(true);
      expect(species.isSpecial).toBe(true);
    });

    it('should create a mythical Species', () => {
      const species = Species.create({
        ...validProps,
        id: 151,
        name: 'mew',
        generation: 1,
        isMythical: true,
      });

      expect(species.isMythical).toBe(true);
      expect(species.isSpecial).toBe(true);
    });

    it('should allow null habitat', () => {
      const species = Species.create({ ...validProps, habitat: null });
      expect(species.habitat).toBeNull();
    });

    it('should trim name and flavorText', () => {
      const species = Species.create({
        ...validProps,
        name: '  bulbasaur  ',
        flavorText: '  A strange seed.  ',
      });

      expect(species.name).toBe('bulbasaur');
      expect(species.flavorText).toBe('A strange seed.');
    });

    it('should support all valid generations (1-9)', () => {
      for (let gen = 1; gen <= 9; gen++) {
        const species = Species.create({ ...validProps, generation: gen });
        expect(species.generation).toBe(gen);
      }
    });
  });

  describe('validation', () => {
    it('should reject id of 0', () => {
      expect(() => Species.create({ ...validProps, id: 0 })).toThrow(InvalidSpeciesError);
    });

    it('should reject negative id', () => {
      expect(() => Species.create({ ...validProps, id: -1 })).toThrow(InvalidSpeciesError);
    });

    it('should reject non-integer id', () => {
      expect(() => Species.create({ ...validProps, id: 1.5 })).toThrow(InvalidSpeciesError);
    });

    it('should reject empty name', () => {
      expect(() => Species.create({ ...validProps, name: '' })).toThrow(InvalidSpeciesError);
    });

    it('should reject whitespace-only name', () => {
      expect(() => Species.create({ ...validProps, name: '   ' })).toThrow(InvalidSpeciesError);
    });

    it('should reject generation 0', () => {
      expect(() => Species.create({ ...validProps, generation: 0 })).toThrow(InvalidSpeciesError);
    });

    it('should reject generation 10', () => {
      expect(() => Species.create({ ...validProps, generation: 10 })).toThrow(InvalidSpeciesError);
    });

    it('should reject negative generation', () => {
      expect(() => Species.create({ ...validProps, generation: -1 })).toThrow(InvalidSpeciesError);
    });

    it('should reject non-integer generation', () => {
      expect(() => Species.create({ ...validProps, generation: 1.5 })).toThrow(InvalidSpeciesError);
    });
  });

  describe('equals', () => {
    it('should return true for Species with the same id', () => {
      const species1 = Species.create(validProps);
      const species2 = Species.create({ ...validProps, name: 'different' });
      expect(species1.equals(species2)).toBe(true);
    });

    it('should return false for Species with different ids', () => {
      const species1 = Species.create(validProps);
      const species2 = Species.create({ ...validProps, id: 2 });
      expect(species1.equals(species2)).toBe(false);
    });
  });
});
