import { PhysicalMeasurement } from './PhysicalMeasurement';
import type { PokemonProps } from './pokemon.types';
import { PokemonType } from './PokemonType';
import { Sprites } from './Sprites';
import { Stats } from './Stats';

import { InvalidPokemonEntityError } from '@/domain/errors';

export class Pokemon {
  private readonly _id: number;
  private readonly _name: string;
  private readonly _types: readonly PokemonType[];
  private readonly _stats: Stats;
  private readonly _height: PhysicalMeasurement;
  private readonly _weight: PhysicalMeasurement;
  private readonly _sprites: Sprites;
  private readonly _order: number;
  private readonly _baseExperience: number | null;

  private constructor(
    id: number,
    name: string,
    types: readonly PokemonType[],
    stats: Stats,
    height: PhysicalMeasurement,
    weight: PhysicalMeasurement,
    sprites: Sprites,
    order: number,
    baseExperience: number | null
  ) {
    this._id = id;
    this._name = name;
    this._types = types;
    this._stats = stats;
    this._height = height;
    this._weight = weight;
    this._sprites = sprites;
    this._order = order;
    this._baseExperience = baseExperience;
  }

  static create(props: PokemonProps): Pokemon {
    Pokemon.validateId(props.id);
    Pokemon.validateName(props.name);
    Pokemon.validateTypes(props.types);
    Pokemon.validateOrder(props.order);

    const types = props.types.map((t) => new PokemonType(t));
    const stats = new Stats(props.stats);
    const height = new PhysicalMeasurement(props.height.value, props.height.unit);
    const weight = new PhysicalMeasurement(props.weight.value, props.weight.unit);
    const sprites = new Sprites(props.sprites);

    return new Pokemon(
      props.id,
      props.name.trim(),
      Object.freeze([...types]),
      stats,
      height,
      weight,
      sprites,
      props.order,
      props.baseExperience
    );
  }

  private static validateId(id: number): void {
    if (!Number.isInteger(id) || id <= 0) {
      throw new InvalidPokemonEntityError(
        `Pokemon ID must be a positive integer, got ${String(id)}`
      );
    }
  }

  private static validateName(name: string): void {
    if (typeof name !== 'string' || name.trim() === '') {
      throw new InvalidPokemonEntityError('Pokemon name cannot be empty');
    }
  }

  private static validateTypes(types: string[]): void {
    if (!Array.isArray(types) || types.length === 0) {
      throw new InvalidPokemonEntityError('Pokemon must have at least one type');
    }

    if (types.length > 2) {
      throw new InvalidPokemonEntityError('Pokemon cannot have more than two types');
    }
  }

  private static validateOrder(order: number): void {
    if (!Number.isInteger(order)) {
      throw new InvalidPokemonEntityError(`Pokemon order must be an integer, got ${String(order)}`);
    }
  }

  get id(): number {
    return this._id;
  }

  get name(): string {
    return this._name;
  }

  get types(): readonly PokemonType[] {
    return this._types;
  }

  get primaryType(): PokemonType {
    return this._types[0];
  }

  get secondaryType(): PokemonType | null {
    return this._types[1] ?? null;
  }

  get isDualType(): boolean {
    return this._types.length === 2;
  }

  get stats(): Stats {
    return this._stats;
  }

  get height(): PhysicalMeasurement {
    return this._height;
  }

  get weight(): PhysicalMeasurement {
    return this._weight;
  }

  get sprites(): Sprites {
    return this._sprites;
  }

  get order(): number {
    return this._order;
  }

  get baseExperience(): number | null {
    return this._baseExperience;
  }

  hasType(typeName: string): boolean {
    return this._types.some((t) => t.value === typeName.toLowerCase().trim());
  }

  equals(other: Pokemon): boolean {
    return this._id === other._id;
  }
}
