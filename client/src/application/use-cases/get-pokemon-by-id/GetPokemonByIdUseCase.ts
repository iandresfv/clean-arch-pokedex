import type { PokemonDetailDTO } from '@/application/dto';

export interface GetPokemonByIdUseCase {
  execute(id: number): Promise<PokemonDetailDTO>;
}
