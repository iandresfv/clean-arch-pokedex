import type { CacheService } from '@/application/ports';

interface CacheEntry {
  value: unknown;
  expiresAt: number;
}

export class LocalStorageCacheService implements CacheService {
  private readonly prefix: string;

  constructor(prefix = 'pokedex_cache_') {
    this.prefix = prefix;
  }

  get(key: string): unknown {
    const raw = localStorage.getItem(this.prefix + key);

    if (!raw) {
      return null;
    }

    try {
      const entry = JSON.parse(raw) as CacheEntry;

      if (Date.now() > entry.expiresAt) {
        this.delete(key);
        return null;
      }

      return entry.value;
    } catch {
      this.delete(key);
      return null;
    }
  }

  set(key: string, value: unknown, ttlMs: number): void {
    const entry: CacheEntry = {
      value,
      expiresAt: Date.now() + ttlMs,
    };

    localStorage.setItem(this.prefix + key, JSON.stringify(entry));
  }

  has(key: string): boolean {
    return this.get(key) !== null;
  }

  delete(key: string): void {
    localStorage.removeItem(this.prefix + key);
  }

  clear(): void {
    const keysToRemove: string[] = [];

    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key?.startsWith(this.prefix)) {
        keysToRemove.push(key);
      }
    }

    for (const key of keysToRemove) {
      localStorage.removeItem(key);
    }
  }
}
