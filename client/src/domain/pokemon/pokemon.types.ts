import type { MeasurementUnit } from './PhysicalMeasurement';
import type { SpritesData } from './Sprites';
import type { StatsData } from './Stats';

export interface PokemonProps {
  id: number;
  name: string;
  types: string[];
  stats: StatsData;
  height: { value: number; unit: MeasurementUnit };
  weight: { value: number; unit: MeasurementUnit };
  sprites: SpritesData;
  order: number;
  baseExperience: number | null;
}

export interface SpeciesProps {
  id: number;
  name: string;
  generation: number;
  flavorText: string;
  habitat: string | null;
  isLegendary: boolean;
  isMythical: boolean;
}
