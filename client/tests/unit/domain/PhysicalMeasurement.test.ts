import { describe, it, expect } from 'vitest'
import { PhysicalMeasurement } from '@/domain/pokemon'
import { InvalidPhysicalMeasurementError } from '@/domain/errors'

describe('PhysicalMeasurement', () => {
  describe('constructor', () => {
    it('should create measurement with valid data', () => {
      const measurement = new PhysicalMeasurement(7, 'dm')

      expect(measurement.value).toBe(7)
      expect(measurement.unit).toBe('dm')
    })

    it('should throw InvalidPhysicalMeasurementError for negative value', () => {
      expect(() => new PhysicalMeasurement(-5, 'dm')).toThrow(InvalidPhysicalMeasurementError)
      expect(() => new PhysicalMeasurement(-5, 'dm')).toThrow('must be positive')
    })

    it('should throw InvalidPhysicalMeasurementError for zero value', () => {
      expect(() => new PhysicalMeasurement(0, 'kg')).toThrow(InvalidPhysicalMeasurementError)
    })

    it('should throw InvalidPhysicalMeasurementError for non-finite value', () => {
      expect(() => new PhysicalMeasurement(Infinity, 'dm')).toThrow(InvalidPhysicalMeasurementError)
      expect(() => new PhysicalMeasurement(NaN, 'kg')).toThrow('finite number')
    })

    it('should accept all valid units', () => {
      expect(() => new PhysicalMeasurement(10, 'dm')).not.toThrow()
      expect(() => new PhysicalMeasurement(10, 'm')).not.toThrow()
      expect(() => new PhysicalMeasurement(10, 'hg')).not.toThrow()
      expect(() => new PhysicalMeasurement(10, 'kg')).not.toThrow()
    })
  })

  describe('toMeters', () => {
    it('should convert decimeters to meters', () => {
      const measurement = new PhysicalMeasurement(7, 'dm')

      expect(measurement.toMeters()).toBe(0.7)
    })

    it('should return same value for meters', () => {
      const measurement = new PhysicalMeasurement(1.5, 'm')

      expect(measurement.toMeters()).toBe(1.5)
    })

    it('should throw when converting weight to meters', () => {
      const measurement = new PhysicalMeasurement(69, 'hg')

      expect(() => measurement.toMeters()).toThrow(InvalidPhysicalMeasurementError)
      expect(() => measurement.toMeters()).toThrow('Cannot convert weight')
    })
  })

  describe('toKilograms', () => {
    it('should convert hectograms to kilograms', () => {
      const measurement = new PhysicalMeasurement(69, 'hg')

      expect(measurement.toKilograms()).toBe(6.9)
    })

    it('should return same value for kilograms', () => {
      const measurement = new PhysicalMeasurement(10.5, 'kg')

      expect(measurement.toKilograms()).toBe(10.5)
    })

    it('should throw when converting length to kilograms', () => {
      const measurement = new PhysicalMeasurement(7, 'dm')

      expect(() => measurement.toKilograms()).toThrow(InvalidPhysicalMeasurementError)
      expect(() => measurement.toKilograms()).toThrow('Cannot convert length')
    })
  })

  describe('toDisplayString', () => {
    it('should display decimeters as meters', () => {
      const measurement = new PhysicalMeasurement(7, 'dm')

      expect(measurement.toDisplayString()).toBe('0.70 m')
    })

    it('should display meters', () => {
      const measurement = new PhysicalMeasurement(1.83, 'm')

      expect(measurement.toDisplayString()).toBe('1.83 m')
    })

    it('should display hectograms as kilograms', () => {
      const measurement = new PhysicalMeasurement(69, 'hg')

      expect(measurement.toDisplayString()).toBe('6.90 kg')
    })

    it('should display kilograms', () => {
      const measurement = new PhysicalMeasurement(100.5, 'kg')

      expect(measurement.toDisplayString()).toBe('100.50 kg')
    })

    it('should format with 2 decimal places', () => {
      const measurement = new PhysicalMeasurement(123, 'hg')

      expect(measurement.toDisplayString()).toBe('12.30 kg')
    })
  })
})
