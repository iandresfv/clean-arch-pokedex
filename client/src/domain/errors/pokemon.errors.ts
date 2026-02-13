import { DomainError } from './DomainError'

export class InvalidPokemonTypeError extends DomainError {
  constructor(type: string) {
    super(
      `Invalid Pokemon type: "${type}". Must be one of the 18 official Pokemon types.`,
      'INVALID_POKEMON_TYPE'
    )
  }
}

export class InvalidStatsError extends DomainError {
  constructor(message: string) {
    super(message, 'INVALID_STATS')
  }
}

export class InvalidPhysicalMeasurementError extends DomainError {
  constructor(message: string) {
    super(message, 'INVALID_PHYSICAL_MEASUREMENT')
  }
}

export class InvalidSpriteUrlError extends DomainError {
  constructor(url: string) {
    super(`Invalid sprite URL: "${url}"`, 'INVALID_SPRITE_URL')
  }
}
