import { LocalStorageCacheService } from '@/infrastructure/cache/LocalStorageCacheService';

describe('LocalStorageCacheService', () => {
  let cache: LocalStorageCacheService;

  beforeEach(() => {
    localStorage.clear();
    cache = new LocalStorageCacheService('test_');
  });

  describe('set and get', () => {
    it('should store and retrieve a value', () => {
      cache.set('key1', { name: 'Pikachu' }, 60000);

      const result = cache.get('key1');
      expect(result).toEqual({ name: 'Pikachu' });
    });

    it('should return null for non-existent key', () => {
      expect(cache.get('nonexistent')).toBeNull();
    });

    it('should return null for expired entry', () => {
      // Negative TTL ensures the entry is already expired
      cache.set('expired', 'value', -1);
      expect(cache.get('expired')).toBeNull();
    });

    it('should store primitive values', () => {
      cache.set('number', 42, 60000);
      cache.set('string', 'hello', 60000);
      cache.set('boolean', true, 60000);

      expect(cache.get('number')).toBe(42);
      expect(cache.get('string')).toBe('hello');
      expect(cache.get('boolean')).toBe(true);
    });

    it('should store arrays', () => {
      cache.set('arr', [1, 2, 3], 60000);
      expect(cache.get('arr')).toEqual([1, 2, 3]);
    });
  });

  describe('has', () => {
    it('should return true for existing key', () => {
      cache.set('exists', 'value', 60000);
      expect(cache.has('exists')).toBe(true);
    });

    it('should return false for non-existent key', () => {
      expect(cache.has('nope')).toBe(false);
    });
  });

  describe('delete', () => {
    it('should remove a cached entry', () => {
      cache.set('key', 'value', 60000);
      cache.delete('key');
      expect(cache.get('key')).toBeNull();
    });
  });

  describe('clear', () => {
    it('should remove only prefixed entries', () => {
      cache.set('a', 1, 60000);
      cache.set('b', 2, 60000);
      localStorage.setItem('other_key', 'should remain');

      cache.clear();

      expect(cache.get('a')).toBeNull();
      expect(cache.get('b')).toBeNull();
      expect(localStorage.getItem('other_key')).toBe('should remain');
    });
  });

  describe('corrupted data', () => {
    it('should return null and clean up corrupted entries', () => {
      localStorage.setItem('test_corrupted', 'not-json');
      expect(cache.get('corrupted')).toBeNull();
    });
  });
});
