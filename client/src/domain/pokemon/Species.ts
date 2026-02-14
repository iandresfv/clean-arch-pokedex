import type { SpeciesProps } from './pokemon.types';

import { InvalidSpeciesError } from '@/domain/errors';

const MIN_GENERATION = 1;
const MAX_GENERATION = 9;

export class Species {
  private readonly _id: number;
  private readonly _name: string;
  private readonly _generation: number;
  private readonly _flavorText: string;
  private readonly _habitat: string | null;
  private readonly _isLegendary: boolean;
  private readonly _isMythical: boolean;

  private constructor(
    id: number,
    name: string,
    generation: number,
    flavorText: string,
    habitat: string | null,
    isLegendary: boolean,
    isMythical: boolean
  ) {
    this._id = id;
    this._name = name;
    this._generation = generation;
    this._flavorText = flavorText;
    this._habitat = habitat;
    this._isLegendary = isLegendary;
    this._isMythical = isMythical;
  }

  static create(props: SpeciesProps): Species {
    Species.validateId(props.id);
    Species.validateName(props.name);
    Species.validateGeneration(props.generation);

    return new Species(
      props.id,
      props.name.trim(),
      props.generation,
      props.flavorText.trim(),
      props.habitat?.trim() ?? null,
      props.isLegendary,
      props.isMythical
    );
  }

  private static validateId(id: number): void {
    if (!Number.isInteger(id) || id <= 0) {
      throw new InvalidSpeciesError(`Species ID must be a positive integer, got ${String(id)}`);
    }
  }

  private static validateName(name: string): void {
    if (typeof name !== 'string' || name.trim() === '') {
      throw new InvalidSpeciesError('Species name cannot be empty');
    }
  }

  private static validateGeneration(generation: number): void {
    if (!Number.isInteger(generation)) {
      throw new InvalidSpeciesError(`Generation must be an integer, got ${String(generation)}`);
    }

    if (generation < MIN_GENERATION || generation > MAX_GENERATION) {
      throw new InvalidSpeciesError(
        `Generation must be between ${String(MIN_GENERATION)} and ${String(MAX_GENERATION)}, got ${String(generation)}`
      );
    }
  }

  get id(): number {
    return this._id;
  }

  get name(): string {
    return this._name;
  }

  get generation(): number {
    return this._generation;
  }

  get flavorText(): string {
    return this._flavorText;
  }

  get habitat(): string | null {
    return this._habitat;
  }

  get isLegendary(): boolean {
    return this._isLegendary;
  }

  get isMythical(): boolean {
    return this._isMythical;
  }

  get isSpecial(): boolean {
    return this._isLegendary || this._isMythical;
  }

  equals(other: Species): boolean {
    return this._id === other._id;
  }
}
