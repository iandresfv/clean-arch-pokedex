/**
 * Response types of the project's own Go API.
 *
 * These mirror the server's JSON exactly and exist only at this boundary; the
 * mappers turn them into domain entities so nothing above infrastructure ever
 * sees a wire type.
 */

export interface PokedexSpritesResponse {
  frontDefault: string | null;
  frontShiny: string | null;
  backDefault: string | null;
  backShiny: string | null;
  officialArtwork: string | null;
}

export interface PokedexStatsResponse {
  hp: number;
  attack: number;
  defense: number;
  specialAttack: number;
  specialDefense: number;
  speed: number;
}

export interface PokedexSpeciesResponse {
  id: number;
  name: string;
  generation: number;
  flavorText: string;
  habitat: string | null;
  color: string | null;
  shape: string | null;
  isLegendary: boolean;
  isMythical: boolean;
}

export interface PokedexPokemonResponse {
  id: number;
  name: string;
  pokedexOrder: number;
  types: string[];
  stats: PokedexStatsResponse;
  heightDm: number;
  weightHg: number;
  baseExperience: number | null;
  sprites: PokedexSpritesResponse;
  species: PokedexSpeciesResponse | null;
}

export interface PokedexListItemResponse {
  id: number;
  name: string;
  pokedexOrder: number;
  types: string[];
  stats: PokedexStatsResponse;
  heightDm: number;
  weightHg: number;
  baseExperience: number | null;
  sprites: PokedexSpritesResponse;
  spriteUrl: string | null;
}

/**
 * Field names match the client's own PaginatedResult<T>, which is why the
 * adapter needs no renaming step.
 */
export interface PokedexPaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
  hasNextPage: boolean;
  hasPreviousPage: boolean;
}

export interface PokedexTypeResponse {
  id: number;
  name: string;
}

export interface PokedexMatchupsResponse {
  type: string;
  weaknesses: string[];
  resistances: string[];
  immunities: string[];
}
