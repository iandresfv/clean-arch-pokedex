import { describe, expect, it } from 'vitest';

import { InvalidSpriteUrlError } from '@/domain/errors';
import { Sprites } from '@/domain/pokemon';

describe('Sprites', () => {
  const validSprites = {
    frontDefault: 'https://example.com/front.png',
    frontShiny: 'https://example.com/front-shiny.png',
    backDefault: 'https://example.com/back.png',
    backShiny: 'https://example.com/back-shiny.png',
    officialArtwork: 'https://example.com/official.png',
  };

  describe('constructor', () => {
    it('should create Sprites with valid URLs', () => {
      const sprites = new Sprites(validSprites);

      expect(sprites.frontDefault).toBe(validSprites.frontDefault);
      expect(sprites.frontShiny).toBe(validSprites.frontShiny);
      expect(sprites.backDefault).toBe(validSprites.backDefault);
      expect(sprites.backShiny).toBe(validSprites.backShiny);
      expect(sprites.officialArtwork).toBe(validSprites.officialArtwork);
    });

    it('should allow null sprites', () => {
      const spritesWithNull = {
        frontDefault: 'https://example.com/front.png',
        frontShiny: null,
        backDefault: null,
        backShiny: null,
        officialArtwork: null,
      };

      const sprites = new Sprites(spritesWithNull);

      expect(sprites.frontDefault).toBe(spritesWithNull.frontDefault);
      expect(sprites.frontShiny).toBeNull();
      expect(sprites.backDefault).toBeNull();
      expect(sprites.backShiny).toBeNull();
      expect(sprites.officialArtwork).toBeNull();
    });

    it('should handle missing officialArtwork', () => {
      const spritesWithoutOfficial = {
        frontDefault: 'https://example.com/front.png',
        frontShiny: null,
        backDefault: null,
        backShiny: null,
      };

      const sprites = new Sprites(spritesWithoutOfficial);

      expect(sprites.officialArtwork).toBeNull();
    });

    it('should throw InvalidSpriteUrlError for empty string URL', () => {
      const invalidSprites = {
        ...validSprites,
        frontDefault: '',
      };

      expect(() => new Sprites(invalidSprites)).toThrow(InvalidSpriteUrlError);
    });

    it('should throw InvalidSpriteUrlError for invalid URL format', () => {
      const invalidSprites = {
        ...validSprites,
        frontShiny: 'not-a-valid-url',
      };

      expect(() => new Sprites(invalidSprites)).toThrow(InvalidSpriteUrlError);
    });

    it('should throw InvalidSpriteUrlError for whitespace-only URL', () => {
      const invalidSprites = {
        ...validSprites,
        backDefault: '   ',
      };

      expect(() => new Sprites(invalidSprites)).toThrow(InvalidSpriteUrlError);
    });
  });

  describe('getBestQuality', () => {
    it('should return official artwork when available', () => {
      const sprites = new Sprites(validSprites);

      expect(sprites.getBestQuality()).toBe(validSprites.officialArtwork);
    });

    it('should fallback to frontDefault when no official artwork', () => {
      const spritesNoOfficial = {
        ...validSprites,
        officialArtwork: null,
      };

      const sprites = new Sprites(spritesNoOfficial);

      expect(sprites.getBestQuality()).toBe(validSprites.frontDefault);
    });

    it('should fallback to frontShiny when no official or frontDefault', () => {
      const spritesOnlyShiny = {
        frontDefault: null,
        frontShiny: 'https://example.com/shiny.png',
        backDefault: null,
        backShiny: null,
        officialArtwork: null,
      };

      const sprites = new Sprites(spritesOnlyShiny);

      expect(sprites.getBestQuality()).toBe(spritesOnlyShiny.frontShiny);
    });

    it('should return null when no sprites available', () => {
      const noSprites = {
        frontDefault: null,
        frontShiny: null,
        backDefault: null,
        backShiny: null,
        officialArtwork: null,
      };

      const sprites = new Sprites(noSprites);

      expect(sprites.getBestQuality()).toBeNull();
    });
  });

  describe('getShinySprite', () => {
    it('should return frontShiny when available', () => {
      const sprites = new Sprites(validSprites);

      expect(sprites.getShinySprite()).toBe(validSprites.frontShiny);
    });

    it('should fallback to backShiny when no frontShiny', () => {
      const spritesNoFrontShiny = {
        ...validSprites,
        frontShiny: null,
      };

      const sprites = new Sprites(spritesNoFrontShiny);

      expect(sprites.getShinySprite()).toBe(validSprites.backShiny);
    });

    it('should return null when no shiny sprites', () => {
      const noShiny = {
        frontDefault: 'https://example.com/front.png',
        frontShiny: null,
        backDefault: null,
        backShiny: null,
      };

      const sprites = new Sprites(noShiny);

      expect(sprites.getShinySprite()).toBeNull();
    });
  });

  describe('hasAnySprite', () => {
    it('should return true when has any sprite', () => {
      const sprites = new Sprites(validSprites);

      expect(sprites.hasAnySprite()).toBe(true);
    });

    it('should return true when has only one sprite', () => {
      const onlyFront = {
        frontDefault: 'https://example.com/front.png',
        frontShiny: null,
        backDefault: null,
        backShiny: null,
      };

      const sprites = new Sprites(onlyFront);

      expect(sprites.hasAnySprite()).toBe(true);
    });

    it('should return false when no sprites', () => {
      const noSprites = {
        frontDefault: null,
        frontShiny: null,
        backDefault: null,
        backShiny: null,
      };

      const sprites = new Sprites(noSprites);

      expect(sprites.hasAnySprite()).toBe(false);
    });
  });
});
