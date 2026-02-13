import { describe, expect,it } from 'vitest'

import { InvalidStatsError } from '@/domain/errors'
import { Stats } from '@/domain/pokemon'

describe('Stats', () => {
  const validStats = {
    hp: 45,
    attack: 49,
    defense: 49,
    specialAttack: 65,
    specialDefense: 65,
    speed: 45,
  }

  describe('constructor', () => {
    it('should create Stats with valid data', () => {
      const stats = new Stats(validStats)

      expect(stats.hp).toBe(45)
      expect(stats.attack).toBe(49)
      expect(stats.defense).toBe(49)
      expect(stats.specialAttack).toBe(65)
      expect(stats.specialDefense).toBe(65)
      expect(stats.speed).toBe(45)
    })

    it('should allow zero values', () => {
      const statsWithZero = {
        ...validStats,
        hp: 0,
      }

      const stats = new Stats(statsWithZero)

      expect(stats.hp).toBe(0)
    })

    it('should throw InvalidStatsError for negative HP', () => {
      const invalidStats = { ...validStats, hp: -1 }

      expect(() => new Stats(invalidStats)).toThrow(InvalidStatsError)
      expect(() => new Stats(invalidStats)).toThrow('HP cannot be negative')
    })

    it('should throw InvalidStatsError for negative Attack', () => {
      const invalidStats = { ...validStats, attack: -10 }

      expect(() => new Stats(invalidStats)).toThrow(InvalidStatsError)
    })

    it('should throw InvalidStatsError for non-integer value', () => {
      const invalidStats = { ...validStats, defense: 49.5 }

      expect(() => new Stats(invalidStats)).toThrow(InvalidStatsError)
      expect(() => new Stats(invalidStats)).toThrow('Defense must be an integer')
    })

    it('should throw InvalidStatsError for float HP', () => {
      const invalidStats = { ...validStats, hp: 45.7 }

      expect(() => new Stats(invalidStats)).toThrow('HP must be an integer')
    })

    it('should validate all stats', () => {
      const negativeDefense = { ...validStats, defense: -5 }
      expect(() => new Stats(negativeDefense)).toThrow('Defense cannot be negative')

      const negativeSpecialAttack = { ...validStats, specialAttack: -1 }
      expect(() => new Stats(negativeSpecialAttack)).toThrow('Special Attack cannot be negative')

      const negativeSpecialDefense = { ...validStats, specialDefense: -1 }
      expect(() => new Stats(negativeSpecialDefense)).toThrow('Special Defense cannot be negative')

      const negativeSpeed = { ...validStats, speed: -1 }
      expect(() => new Stats(negativeSpeed)).toThrow('Speed cannot be negative')
    })
  })

  describe('total', () => {
    it('should calculate total stats correctly', () => {
      const stats = new Stats(validStats)

      expect(stats.total).toBe(318) // 45 + 49 + 49 + 65 + 65 + 45
    })

    it('should calculate total for zero stats', () => {
      const zeroStats = {
        hp: 0,
        attack: 0,
        defense: 0,
        specialAttack: 0,
        specialDefense: 0,
        speed: 0,
      }

      const stats = new Stats(zeroStats)

      expect(stats.total).toBe(0)
    })
  })

  describe('average', () => {
    it('should calculate average stats correctly', () => {
      const stats = new Stats(validStats)

      expect(stats.average).toBe(53) // 318 / 6
    })

    it('should handle decimal averages', () => {
      const unevenStats = {
        hp: 50,
        attack: 50,
        defense: 50,
        specialAttack: 50,
        specialDefense: 50,
        speed: 51,
      }

      const stats = new Stats(unevenStats)

      expect(stats.average).toBeCloseTo(50.16666666666667)
    })
  })

  describe('toObject', () => {
    it('should return stats as plain object', () => {
      const stats = new Stats(validStats)

      const obj = stats.toObject()

      expect(obj).toEqual(validStats)
    })

    it('should return independent copy', () => {
      const stats = new Stats(validStats)

      const obj = stats.toObject()
      obj.hp = 999

      expect(stats.hp).toBe(45)
    })
  })
})
