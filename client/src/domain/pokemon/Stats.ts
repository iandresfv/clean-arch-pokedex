import { InvalidStatsError } from '@/domain/errors'

export interface StatsData {
  hp: number
  attack: number
  defense: number
  specialAttack: number
  specialDefense: number
  speed: number
}

export class Stats {
  private readonly _hp: number
  private readonly _attack: number
  private readonly _defense: number
  private readonly _specialAttack: number
  private readonly _specialDefense: number
  private readonly _speed: number

  constructor(data: StatsData) {
    this.validateStat('HP', data.hp)
    this.validateStat('Attack', data.attack)
    this.validateStat('Defense', data.defense)
    this.validateStat('Special Attack', data.specialAttack)
    this.validateStat('Special Defense', data.specialDefense)
    this.validateStat('Speed', data.speed)

    this._hp = data.hp
    this._attack = data.attack
    this._defense = data.defense
    this._specialAttack = data.specialAttack
    this._specialDefense = data.specialDefense
    this._speed = data.speed
  }

  private validateStat(name: string, value: number): void {
    if (!Number.isInteger(value)) {
      throw new InvalidStatsError(`${name} must be an integer, got ${String(value)}`)
    }

    if (value < 0) {
      throw new InvalidStatsError(`${name} cannot be negative, got ${String(value)}`)
    }
  }

  get hp(): number {
    return this._hp
  }

  get attack(): number {
    return this._attack
  }

  get defense(): number {
    return this._defense
  }

  get specialAttack(): number {
    return this._specialAttack
  }

  get specialDefense(): number {
    return this._specialDefense
  }

  get speed(): number {
    return this._speed
  }

  get total(): number {
    return (
      this._hp +
      this._attack +
      this._defense +
      this._specialAttack +
      this._specialDefense +
      this._speed
    )
  }

  get average(): number {
    return this.total / 6
  }

  toObject(): StatsData {
    return {
      hp: this._hp,
      attack: this._attack,
      defense: this._defense,
      specialAttack: this._specialAttack,
      specialDefense: this._specialDefense,
      speed: this._speed,
    }
  }
}
