import type { GetPokemonByIdUseCase } from './GetPokemonByIdUseCase';

import type { PokemonDetailDTO, SpeciesDTO } from '@/application/dto';
import type { Logger, PokemonRepository } from '@/application/ports';
import type { Pokemon, Species } from '@/domain/pokemon';

export class GetPokemonByIdUseCaseImpl implements GetPokemonByIdUseCase {
  private readonly pokemonRepository: PokemonRepository;
  private readonly logger: Logger;

  constructor(pokemonRepository: PokemonRepository, logger: Logger) {
    this.pokemonRepository = pokemonRepository;
    this.logger = logger;
  }

  async execute(id: number): Promise<PokemonDetailDTO> {
    this.logger.info('GetPokemonByIdUseCase.execute', { id });

    const pokemon = await this.pokemonRepository.getById(id);

    let speciesDTO: SpeciesDTO | null = null;
    try {
      const species = await this.pokemonRepository.getSpeciesById(id);
      speciesDTO = this.speciesToDTO(species);
    } catch (error: unknown) {
      this.logger.warn('Failed to fetch species data', { id, error });
    }

    return this.toDTO(pokemon, speciesDTO);
  }

  private toDTO(pokemon: Pokemon, species: SpeciesDTO | null): PokemonDetailDTO {
    return {
      id: pokemon.id,
      name: pokemon.name,
      types: pokemon.types.map((t) => t.value),
      stats: pokemon.stats.toObject(),
      height: pokemon.height.toDisplayString(),
      weight: pokemon.weight.toDisplayString(),
      spriteUrl: pokemon.sprites.getBestQuality(),
      officialArtworkUrl: pokemon.sprites.officialArtwork,
      sprites: {
        frontDefault: pokemon.sprites.frontDefault,
        frontShiny: pokemon.sprites.frontShiny,
        backDefault: pokemon.sprites.backDefault,
        backShiny: pokemon.sprites.backShiny,
      },
      baseExperience: pokemon.baseExperience,
      species,
    };
  }

  private speciesToDTO(species: Species): SpeciesDTO {
    return {
      id: species.id,
      name: species.name,
      generation: species.generation,
      flavorText: species.flavorText,
      habitat: species.habitat,
      isLegendary: species.isLegendary,
      isMythical: species.isMythical,
    };
  }
}
