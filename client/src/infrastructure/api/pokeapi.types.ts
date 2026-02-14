export interface PokeAPIPokemonResponse {
  id: number;
  name: string;
  order: number;
  base_experience: number | null;
  height: number;
  weight: number;
  types: PokeAPITypeSlot[];
  stats: PokeAPIStatSlot[];
  sprites: PokeAPISpriteSet;
}

export interface PokeAPITypeSlot {
  slot: number;
  type: {
    name: string;
    url: string;
  };
}

export interface PokeAPIStatSlot {
  base_stat: number;
  effort: number;
  stat: {
    name: string;
    url: string;
  };
}

export interface PokeAPISpriteSet {
  front_default: string | null;
  front_shiny: string | null;
  back_default: string | null;
  back_shiny: string | null;
  other?: {
    'official-artwork'?: {
      front_default: string | null;
    };
  };
}

export interface PokeAPISpeciesResponse {
  id: number;
  name: string;
  generation: {
    name: string;
    url: string;
  };
  flavor_text_entries: PokeAPIFlavorTextEntry[];
  habitat: {
    name: string;
    url: string;
  } | null;
  is_legendary: boolean;
  is_mythical: boolean;
}

export interface PokeAPIFlavorTextEntry {
  flavor_text: string;
  language: {
    name: string;
    url: string;
  };
  version: {
    name: string;
    url: string;
  };
}

export interface PokeAPIListResponse {
  count: number;
  next: string | null;
  previous: string | null;
  results: PokeAPIListItem[];
}

export interface PokeAPIListItem {
  name: string;
  url: string;
}
