import { InvalidPhysicalMeasurementError } from '@/domain/errors'

export type MeasurementUnit = 'dm' | 'm' | 'hg' | 'kg'

export class PhysicalMeasurement {
  private readonly _value: number
  private readonly _unit: MeasurementUnit

  constructor(value: number, unit: MeasurementUnit) {
    if (!Number.isFinite(value)) {
      throw new InvalidPhysicalMeasurementError('Measurement value must be a finite number')
    }

    if (value <= 0) {
      throw new InvalidPhysicalMeasurementError('Measurement value must be positive')
    }

    this._value = value
    this._unit = unit
  }

  get value(): number {
    return this._value
  }

  get unit(): MeasurementUnit {
    return this._unit
  }

  toMeters(): number {
    if (this._unit === 'm') {
      return this._value
    }

    if (this._unit === 'dm') {
      return this._value / 10
    }

    throw new InvalidPhysicalMeasurementError('Cannot convert weight unit to meters')
  }

  toKilograms(): number {
    if (this._unit === 'kg') {
      return this._value
    }

    if (this._unit === 'hg') {
      return this._value / 10
    }

    throw new InvalidPhysicalMeasurementError('Cannot convert length unit to kilograms')
  }

  toDisplayString(): string {
    if (this._unit === 'dm' || this._unit === 'm') {
      return `${this.toMeters().toFixed(2)} m`
    }

    if (this._unit === 'hg' || this._unit === 'kg') {
      return `${this.toKilograms().toFixed(2)} kg`
    }

    return `${this._value} ${this._unit}`
  }
}
