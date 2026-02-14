export abstract class ApplicationError extends Error {
  readonly code: string;

  constructor(message: string, code: string) {
    super(message);
    this.code = code;
    this.name = this.constructor.name;
    Object.setPrototypeOf(this, new.target.prototype);
  }
}

export class PokemonNotFoundError extends ApplicationError {
  constructor(id: number) {
    super(`Pokemon with ID ${String(id)} not found`, 'POKEMON_NOT_FOUND');
  }
}

export class SpeciesNotFoundError extends ApplicationError {
  constructor(id: number) {
    super(`Species with ID ${String(id)} not found`, 'SPECIES_NOT_FOUND');
  }
}

export class RepositoryError extends ApplicationError {
  constructor(message: string) {
    super(message, 'REPOSITORY_ERROR');
  }
}
