import { describe, it, expect } from 'vitest'
import {
  DomainError,
  InvalidPokemonTypeError,
  InvalidStatsError,
  InvalidPhysicalMeasurementError,
  InvalidSpriteUrlError,
} from '@/domain/errors'

describe('DomainError', () => {
  describe('InvalidPokemonTypeError', () => {
    it('should create error with correct message and code', () => {
      const error = new InvalidPokemonTypeError('banana')

      expect(error).toBeInstanceOf(Error)
      expect(error).toBeInstanceOf(DomainError)
      expect(error.message).toContain('Invalid Pokemon type')
      expect(error.message).toContain('banana')
      expect(error.code).toBe('INVALID_POKEMON_TYPE')
      expect(error.name).toBe('InvalidPokemonTypeError')
    })
  })

  describe('InvalidStatsError', () => {
    it('should create error with custom message', () => {
      const error = new InvalidStatsError('HP cannot be negative')

      expect(error).toBeInstanceOf(DomainError)
      expect(error.message).toBe('HP cannot be negative')
      expect(error.code).toBe('INVALID_STATS')
    })
  })

  describe('InvalidPhysicalMeasurementError', () => {
    it('should create error with custom message', () => {
      const error = new InvalidPhysicalMeasurementError('Height must be positive')

      expect(error).toBeInstanceOf(DomainError)
      expect(error.message).toBe('Height must be positive')
      expect(error.code).toBe('INVALID_PHYSICAL_MEASUREMENT')
    })
  })

  describe('InvalidSpriteUrlError', () => {
    it('should create error with URL in message', () => {
      const error = new InvalidSpriteUrlError('not-a-url')

      expect(error).toBeInstanceOf(DomainError)
      expect(error.message).toContain('not-a-url')
      expect(error.code).toBe('INVALID_SPRITE_URL')
    })
  })
})
