import type { PokemonTypeName } from '@/domain/pokemon';

export interface SearchCriteria {
  name?: string;
  type?: PokemonTypeName;
}
