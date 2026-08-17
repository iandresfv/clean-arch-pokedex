/**
 * Typed access to build-time environment variables.
 *
 * Reading `import.meta.env` directly at call sites produces `undefined` for a
 * missing variable, which silently becomes the string "undefined" inside a URL
 * and surfaces as a confusing 404 at runtime. Resolving everything here means a
 * misconfigured build fails immediately, with a message naming the variable.
 */

export type DataSource = 'pokeapi' | 'pokedex-api';

interface Env {
  /** Base URL of the Pokedex API, without a trailing slash. */
  apiUrl: string;
  /** Which repository adapter the composition root should wire. */
  dataSource: DataSource;
  /** Per-request timeout in milliseconds. */
  requestTimeoutMs: number;
}

function readString(key: string, fallback: string): string {
  const raw = import.meta.env[key] as string | undefined;
  return raw !== undefined && raw !== '' ? raw : fallback;
}

function readDataSource(): DataSource {
  const raw = readString('VITE_DATA_SOURCE', 'pokedex-api');

  if (raw !== 'pokeapi' && raw !== 'pokedex-api') {
    throw new Error(
      `VITE_DATA_SOURCE must be "pokeapi" or "pokedex-api", got "${raw}". ` +
        'Check client/.env.local.'
    );
  }
  return raw;
}

function readTimeout(): number {
  const raw = readString('VITE_REQUEST_TIMEOUT_MS', '10000');
  const parsed = Number.parseInt(raw, 10);

  if (Number.isNaN(parsed) || parsed <= 0) {
    throw new Error(`VITE_REQUEST_TIMEOUT_MS must be a positive integer, got "${raw}".`);
  }
  return parsed;
}

export const env: Env = {
  // Trailing slashes are stripped so path concatenation cannot produce "//".
  apiUrl: readString('VITE_API_URL', 'http://localhost:8080').replace(/\/+$/, ''),
  dataSource: readDataSource(),
  requestTimeoutMs: readTimeout(),
};
