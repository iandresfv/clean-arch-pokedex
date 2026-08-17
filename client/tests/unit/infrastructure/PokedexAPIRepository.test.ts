import type { Mock } from 'vitest';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { HttpClient, HttpError } from '@/infrastructure/http/HttpClient';
import { PokedexAPIRepository } from '@/infrastructure/repositories/PokedexAPIRepository';

/**
 * Builds a list item exactly as the Go API serialises one, so the test fails if
 * either side of the contract drifts.
 */
function apiListItem(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    id: 1,
    name: 'bulbasaur',
    pokedexOrder: 1,
    types: ['grass', 'poison'],
    stats: { hp: 45, attack: 49, defense: 49, specialAttack: 65, specialDefense: 65, speed: 45 },
    heightDm: 7,
    weightHg: 69,
    baseExperience: 64,
    sprites: {
      frontDefault: 'https://example.test/1.png',
      frontShiny: 'https://example.test/1-shiny.png',
      backDefault: null,
      backShiny: null,
      officialArtwork: 'https://example.test/1-art.png',
    },
    spriteUrl: 'https://example.test/1-art.png',
    ...overrides,
  };
}

function apiPage(items: unknown[], total = 1302) {
  return {
    data: items,
    total,
    page: 1,
    limit: 20,
    totalPages: Math.ceil(total / 20),
    hasNextPage: total > 20,
    hasPreviousPage: false,
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('PokedexAPIRepository', () => {
  let fetchMock: Mock;
  let repository: PokedexAPIRepository;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    repository = new PokedexAPIRepository(new HttpClient('http://api.test', 5000));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  describe('getAll', () => {
    it('fetches a page in a single request', async () => {
      fetchMock.mockResolvedValue(jsonResponse(apiPage([apiListItem()])));

      const result = await repository.getAll({ page: 1, limit: 20 });

      // The whole point of the migration: one request instead of one index
      // call plus one detail call per item.
      expect(fetchMock).toHaveBeenCalledTimes(1);
      expect(result.data).toHaveLength(1);
    });

    it('passes pagination through as query parameters', async () => {
      fetchMock.mockResolvedValue(jsonResponse(apiPage([])));

      await repository.getAll({ page: 3, limit: 50 });

      const url = fetchMock.mock.calls[0]?.[0] as string;
      expect(url).toContain('page=3');
      expect(url).toContain('limit=50');
    });

    it('uses the pagination metadata the server computed', async () => {
      fetchMock.mockResolvedValue(jsonResponse(apiPage([apiListItem()], 1302)));

      const result = await repository.getAll({ page: 1, limit: 20 });

      expect(result.total).toBe(1302);
      expect(result.totalPages).toBe(66);
      expect(result.hasNextPage).toBe(true);
      expect(result.hasPreviousPage).toBe(false);
    });

    it('maps responses into domain entities', async () => {
      fetchMock.mockResolvedValue(jsonResponse(apiPage([apiListItem()])));

      const { data } = await repository.getAll({ page: 1, limit: 20 });
      const pokemon = data[0];

      expect(pokemon).toBeDefined();
      expect(pokemon.name).toBe('bulbasaur');
      // Slot order is meaningful: the UI renders the primary type first.
      expect(pokemon.types.map((t) => t.value)).toEqual(['grass', 'poison']);
      expect(pokemon.stats.hp).toBe(45);
      expect(pokemon.sprites.getBestQuality()).toBe('https://example.test/1-art.png');
    });
  });

  describe('searchByName', () => {
    it('delegates filtering to the server', async () => {
      fetchMock.mockResolvedValue(jsonResponse(apiPage([apiListItem()], 3)));

      const result = await repository.searchByName('char', { page: 1, limit: 20 });

      // The PokeAPI adapter had to download all 1302 names and filter locally.
      expect(fetchMock).toHaveBeenCalledTimes(1);
      const url = fetchMock.mock.calls[0]?.[0] as string;
      expect(url).toContain('q=char');
      expect(result.total).toBe(3);
    });

    it('encodes terms that would otherwise break the query string', async () => {
      fetchMock.mockResolvedValue(jsonResponse(apiPage([])));

      await repository.searchByName('mr. mime & co', { page: 1, limit: 20 });

      const url = fetchMock.mock.calls[0]?.[0] as string;
      expect(url).not.toContain(' ');
      expect(url).toContain('q=mr.+mime+%26+co');
    });
  });

  describe('getById', () => {
    it('returns a fully populated entity including species', async () => {
      fetchMock.mockResolvedValue(
        jsonResponse({
          ...apiListItem(),
          species: {
            id: 1,
            name: 'bulbasaur',
            generation: 1,
            flavorText: 'A strange seed was planted on its back at birth.',
            habitat: 'grassland',
            color: 'green',
            shape: 'quadruped',
            isLegendary: false,
            isMythical: false,
          },
        })
      );

      const pokemon = await repository.getById(1);

      expect(pokemon.id).toBe(1);
      expect(pokemon.stats.attack).toBe(49);
    });

    it('surfaces a 404 as an HttpError carrying the correlation id', async () => {
      fetchMock.mockResolvedValue(
        new Response(
          JSON.stringify({
            type: 'https://example.test/errors/not-found',
            title: 'Pokemon not found',
            status: 404,
            detail: 'no pokemon exists with the requested identifier',
            requestId: '8f14e45fceea167a',
          }),
          { status: 404, headers: { 'Content-Type': 'application/problem+json' } }
        )
      );

      await expect(repository.getById(9999)).rejects.toBeInstanceOf(HttpError);

      // Rejected twice on purpose: the second call inspects the error object.
      fetchMock.mockResolvedValue(
        new Response(
          JSON.stringify({ title: 'Pokemon not found', status: 404, requestId: 'abc' }),
          {
            status: 404,
            headers: { 'Content-Type': 'application/problem+json' },
          }
        )
      );
      await repository.getById(9999).catch((error: unknown) => {
        expect(error).toBeInstanceOf(HttpError);
        const httpError = error as HttpError;
        expect(httpError.isNotFound).toBe(true);
        expect(httpError.requestId).toBe('abc');
      });
    });
  });

  describe('getSpeciesById', () => {
    it('maps a species response into a domain entity', async () => {
      fetchMock.mockResolvedValue(
        jsonResponse({
          id: 25,
          name: 'pikachu',
          generation: 1,
          flavorText: 'It keeps its tail raised to monitor its surroundings.',
          habitat: 'forest',
          color: 'yellow',
          shape: 'quadruped',
          isLegendary: false,
          isMythical: false,
        })
      );

      const species = await repository.getSpeciesById(25);

      expect(species.name).toBe('pikachu');
      expect(species.generation).toBe(1);
      expect(species.isLegendary).toBe(false);
    });
  });

  describe('error handling', () => {
    it('reports a 500 without leaking the body as a success', async () => {
      fetchMock.mockResolvedValue(
        new Response(JSON.stringify({ title: 'Internal Server Error', status: 500 }), {
          status: 500,
          headers: { 'Content-Type': 'application/problem+json' },
        })
      );

      await expect(repository.getAll({ page: 1, limit: 20 })).rejects.toBeInstanceOf(HttpError);
    });

    it('does not crash when the error body is not JSON', async () => {
      // A proxy or load balancer in front of the API returns HTML, and parsing
      // it must not mask the real status.
      fetchMock.mockResolvedValue(
        new Response('<html>502 Bad Gateway</html>', {
          status: 502,
          headers: { 'Content-Type': 'text/html' },
        })
      );

      await expect(repository.getAll({ page: 1, limit: 20 })).rejects.toThrow(/502/);
    });
  });
});
