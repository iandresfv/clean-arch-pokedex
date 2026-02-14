import { PokemonType, TypeEffectivenessService } from '@/domain/pokemon';

describe('TypeEffectivenessService', () => {
  const service = new TypeEffectivenessService();

  describe('getAttackMultiplier', () => {
    it('should return 2 for super effective (fire → grass)', () => {
      expect(service.getAttackMultiplier(new PokemonType('fire'), new PokemonType('grass'))).toBe(
        2
      );
    });

    it('should return 0.5 for not very effective (fire → water)', () => {
      expect(service.getAttackMultiplier(new PokemonType('fire'), new PokemonType('water'))).toBe(
        0.5
      );
    });

    it('should return 0 for no effect (normal → ghost)', () => {
      expect(service.getAttackMultiplier(new PokemonType('normal'), new PokemonType('ghost'))).toBe(
        0
      );
    });

    it('should return 1 for neutral (fire → normal)', () => {
      expect(service.getAttackMultiplier(new PokemonType('fire'), new PokemonType('normal'))).toBe(
        1
      );
    });

    it('should handle electric → ground (no effect)', () => {
      expect(
        service.getAttackMultiplier(new PokemonType('electric'), new PokemonType('ground'))
      ).toBe(0);
    });

    it('should handle dragon → fairy (no effect)', () => {
      expect(service.getAttackMultiplier(new PokemonType('dragon'), new PokemonType('fairy'))).toBe(
        0
      );
    });
  });

  describe('getDefenseMultiplier', () => {
    it('should calculate multiplier against single type', () => {
      const multiplier = service.getDefenseMultiplier(new PokemonType('fire'), [
        new PokemonType('grass'),
      ]);
      expect(multiplier).toBe(2);
    });

    it('should calculate combined multiplier against dual type (fire → grass/poison = 2 * 1 = 2)', () => {
      const multiplier = service.getDefenseMultiplier(new PokemonType('fire'), [
        new PokemonType('grass'),
        new PokemonType('poison'),
      ]);
      expect(multiplier).toBe(2);
    });

    it('should calculate 4x effectiveness (ground → fire/steel = 2 * 2 = 4)', () => {
      const multiplier = service.getDefenseMultiplier(new PokemonType('ground'), [
        new PokemonType('fire'),
        new PokemonType('steel'),
      ]);
      expect(multiplier).toBe(4);
    });

    it('should calculate 0.25x resistance (fire → fire/water = 0.5 * 0.5 = 0.25)', () => {
      const multiplier = service.getDefenseMultiplier(new PokemonType('fire'), [
        new PokemonType('fire'),
        new PokemonType('water'),
      ]);
      expect(multiplier).toBe(0.25);
    });

    it('should calculate 0x when one type is immune (normal → ghost/dark = 0 * 0.5 = 0)', () => {
      const multiplier = service.getDefenseMultiplier(new PokemonType('normal'), [
        new PokemonType('ghost'),
        new PokemonType('dark'),
      ]);
      expect(multiplier).toBe(0);
    });
  });

  describe('getWeaknesses', () => {
    it('should return weaknesses for a single fire type', () => {
      const weaknesses = service.getWeaknesses([new PokemonType('fire')]);
      expect(weaknesses).toContain('water');
      expect(weaknesses).toContain('ground');
      expect(weaknesses).toContain('rock');
      expect(weaknesses).not.toContain('grass');
    });

    it('should return weaknesses for dual type grass/poison', () => {
      const weaknesses = service.getWeaknesses([
        new PokemonType('grass'),
        new PokemonType('poison'),
      ]);
      expect(weaknesses).toContain('fire');
      expect(weaknesses).toContain('psychic');
      expect(weaknesses).toContain('flying');
      expect(weaknesses).toContain('ice');
    });
  });

  describe('getResistances', () => {
    it('should return resistances for a single fire type', () => {
      const resistances = service.getResistances([new PokemonType('fire')]);
      expect(resistances).toContain('fire');
      expect(resistances).toContain('grass');
      expect(resistances).toContain('ice');
      expect(resistances).toContain('bug');
      expect(resistances).toContain('steel');
      expect(resistances).toContain('fairy');
    });

    it('should return resistances for steel/fairy (many resistances)', () => {
      const resistances = service.getResistances([
        new PokemonType('steel'),
        new PokemonType('fairy'),
      ]);
      expect(resistances.length).toBeGreaterThan(5);
    });
  });

  describe('getImmunities', () => {
    it('should return immunities for ghost type', () => {
      const immunities = service.getImmunities([new PokemonType('ghost')]);
      expect(immunities).toContain('normal');
      expect(immunities).toContain('fighting');
    });

    it('should return empty for types with no immunities', () => {
      const immunities = service.getImmunities([new PokemonType('fire')]);
      expect(immunities).toHaveLength(0);
    });

    it('should return immunities for dual type (normal/ghost)', () => {
      const immunities = service.getImmunities([
        new PokemonType('normal'),
        new PokemonType('ghost'),
      ]);
      // ghost is immune to normal and fighting; normal is immune to ghost
      expect(immunities).toContain('normal');
      expect(immunities).toContain('fighting');
      expect(immunities).toContain('ghost');
    });
  });
});
