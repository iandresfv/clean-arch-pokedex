import { createContainer } from '@/di/container';

describe('DI Container', () => {
  it('should create a container with all dependencies', () => {
    const container = createContainer();

    expect(container.pokemonRepository).toBeDefined();
    expect(container.cacheService).toBeDefined();
    expect(container.logger).toBeDefined();
    expect(container.listPokemonUseCase).toBeDefined();
    expect(container.getPokemonByIdUseCase).toBeDefined();
    expect(container.searchPokemonUseCase).toBeDefined();
  });

  it('should create a new container instance each time', () => {
    const container1 = createContainer();
    const container2 = createContainer();

    expect(container1).not.toBe(container2);
    expect(container1.pokemonRepository).not.toBe(container2.pokemonRepository);
  });

  it('should provide use cases that implement their interfaces', () => {
    const container = createContainer();

    expect(typeof container.listPokemonUseCase.execute).toBe('function');
    expect(typeof container.getPokemonByIdUseCase.execute).toBe('function');
    expect(typeof container.searchPokemonUseCase.execute).toBe('function');
  });

  it('should provide a cache service with full CRUD', () => {
    const container = createContainer();
    const { cacheService } = container;

    expect(typeof cacheService.get).toBe('function');
    expect(typeof cacheService.set).toBe('function');
    expect(typeof cacheService.has).toBe('function');
    expect(typeof cacheService.delete).toBe('function');
    expect(typeof cacheService.clear).toBe('function');
  });

  it('should provide a logger with all levels', () => {
    const container = createContainer();
    const { logger } = container;

    expect(typeof logger.info).toBe('function');
    expect(typeof logger.warn).toBe('function');
    expect(typeof logger.error).toBe('function');
    expect(typeof logger.debug).toBe('function');
  });
});
