import { InvalidPokemonTypeError } from '@/domain/errors'

export const VALID_POKEMON_TYPES = [
  'normal',
  'fire',
  'water',
  'electric',
  'grass',
  'ice',
  'fighting',
  'poison',
  'ground',
  'flying',
  'psychic',
  'bug',
  'rock',
  'ghost',
  'dragon',
  'dark',
  'steel',
  'fairy',
] as const

export type PokemonTypeName = (typeof VALID_POKEMON_TYPES)[number]

export class PokemonType {
  private readonly _value: PokemonTypeName

  constructor(value: string) {
    const normalizedValue = value.toLowerCase().trim()

    if (!this.isValidType(normalizedValue)) {
      throw new InvalidPokemonTypeError(value)
    }

    this._value = normalizedValue
  }

  private isValidType(value: string): value is PokemonTypeName {
    return VALID_POKEMON_TYPES.includes(value as PokemonTypeName)
  }

  get value(): PokemonTypeName {
    return this._value
  }

  equals(other: PokemonType): boolean {
    return this._value === other._value
  }

  toString(): string {
    return this._value
  }
}
