export interface CacheService {
  get(key: string): unknown;
  set(key: string, value: unknown, ttlMs: number): void;
  has(key: string): boolean;
  delete(key: string): void;
  clear(): void;
}
