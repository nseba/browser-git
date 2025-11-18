/**
 * Tests for Mock Storage Adapter
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { MockAdapter } from '../src/mock';

describe('MockAdapter', () => {
  let adapter: MockAdapter;

  beforeEach(() => {
    adapter = new MockAdapter();
  });

  describe('Basic operations', () => {
    it('should get and set values', async () => {
      const key = 'test-key';
      const value = new Uint8Array([1, 2, 3]);

      await adapter.set(key, value);
      const result = await adapter.get(key);

      expect(result).toEqual(value);
    });

    it('should return null for non-existent keys', async () => {
      const result = await adapter.get('non-existent');
      expect(result).toBeNull();
    });

    it('should delete keys', async () => {
      const key = 'test-key';
      const value = new Uint8Array([1, 2, 3]);

      await adapter.set(key, value);
      await adapter.delete(key);

      const result = await adapter.get(key);
      expect(result).toBeNull();
    });

    it('should check if keys exist', async () => {
      const key = 'test-key';
      const value = new Uint8Array([1, 2, 3]);

      expect(await adapter.has(key)).toBe(false);

      await adapter.set(key, value);
      expect(await adapter.has(key)).toBe(true);

      await adapter.delete(key);
      expect(await adapter.has(key)).toBe(false);
    });

    it('should list all keys', async () => {
      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));
      await adapter.set('key3', new Uint8Array([3]));

      const keys = await adapter.keys();
      expect(keys).toHaveLength(3);
      expect(keys).toContain('key1');
      expect(keys).toContain('key2');
      expect(keys).toContain('key3');
    });

    it('should clear all data', async () => {
      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));

      await adapter.clear();

      const keys = await adapter.keys();
      expect(keys).toHaveLength(0);
    });
  });

  describe('Delay simulation', () => {
    it('should simulate delays', async () => {
      adapter.setDelay(50);

      const start = Date.now();
      await adapter.get('test');
      const duration = Date.now() - start;

      // Allow small timing tolerance (2ms) for setTimeout precision
      expect(duration).toBeGreaterThanOrEqual(48);
    });

    it('should apply delay to all operations', async () => {
      adapter.setDelay(20);

      const operations = [
        adapter.set('key', new Uint8Array([1])),
        adapter.get('key'),
        adapter.has('key'),
        adapter.delete('key'),
        adapter.keys(),
      ];

      const start = Date.now();
      await Promise.all(operations);
      const duration = Date.now() - start;

      // Should take at least 100ms (5 operations × 20ms each, running in parallel)
      expect(duration).toBeGreaterThanOrEqual(20);
    });
  });

  describe('Failure simulation', () => {
    it('should fail operations for specific keys', async () => {
      adapter.addFailKey('fail-key');

      await expect(adapter.get('fail-key')).rejects.toThrow('Mock failure for key: fail-key');
      await expect(adapter.set('fail-key', new Uint8Array([1]))).rejects.toThrow();
      await expect(adapter.delete('fail-key')).rejects.toThrow();
      await expect(adapter.has('fail-key')).rejects.toThrow();
    });

    it('should allow removing fail keys', async () => {
      adapter.addFailKey('fail-key');
      adapter.removeFailKey('fail-key');

      await expect(adapter.get('fail-key')).resolves.toBeNull();
    });

    it('should clear all fail keys', async () => {
      adapter.addFailKey('fail1');
      adapter.addFailKey('fail2');
      adapter.clearFailKeys();

      await expect(adapter.get('fail1')).resolves.toBeNull();
      await expect(adapter.get('fail2')).resolves.toBeNull();
    });

    it('should use custom error for failures', async () => {
      const customError = new Error('Custom error message');
      adapter.setFailError(customError);
      adapter.addFailKey('fail-key');

      await expect(adapter.get('fail-key')).rejects.toThrow('Custom error message');
    });
  });

  describe('Quota simulation', () => {
    it('should enforce maximum keys limit', async () => {
      adapter.setMaxKeys(2);

      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));

      await expect(adapter.set('key3', new Uint8Array([3]))).rejects.toThrow(
        'QuotaExceededError'
      );
    });

    it('should allow updating existing keys without checking quota', async () => {
      adapter.setMaxKeys(2);

      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));

      // Update existing key should work
      await expect(adapter.set('key1', new Uint8Array([1, 2, 3]))).resolves.toBeUndefined();
    });

    it('should report quota information', async () => {
      adapter.setMaxKeys(10);

      await adapter.set('key1', new Uint8Array([1, 2, 3]));
      await adapter.set('key2', new Uint8Array([4, 5, 6, 7]));

      const quota = await adapter.getQuota();

      expect(quota.used).toBeGreaterThan(0);
      expect(quota.available).toBeGreaterThan(0);
    });
  });

  describe('Call tracking', () => {
    beforeEach(() => {
      adapter.enableCallTracking();
    });

    it('should track method calls', async () => {
      await adapter.get('key1');
      await adapter.set('key2', new Uint8Array([1]));
      await adapter.delete('key3');

      const calls = adapter.getCalls();
      expect(calls).toHaveLength(3);
      expect(calls[0].method).toBe('get');
      expect(calls[0].args).toEqual(['key1']);
      expect(calls[1].method).toBe('set');
      expect(calls[2].method).toBe('delete');
    });

    it('should clear tracked calls', async () => {
      await adapter.get('key1');
      await adapter.get('key2');

      adapter.clearCalls();

      const calls = adapter.getCalls();
      expect(calls).toHaveLength(0);
    });

    it('should track timestamps', async () => {
      const before = Date.now();
      await adapter.get('key');
      const after = Date.now();

      const calls = adapter.getCalls();
      expect(calls[0].timestamp).toBeGreaterThanOrEqual(before);
      expect(calls[0].timestamp).toBeLessThanOrEqual(after);
    });

    it('should not track calls when disabled', async () => {
      adapter.disableCallTracking();

      await adapter.get('key');
      await adapter.set('key', new Uint8Array([1]));

      const calls = adapter.getCalls();
      expect(calls).toHaveLength(0);
    });
  });

  describe('Transaction support', () => {
    it('should execute transactions', async () => {
      const result = await adapter.transaction(async () => {
        await adapter.set('key1', new Uint8Array([1]));
        await adapter.set('key2', new Uint8Array([2]));
        return 'success';
      });

      expect(result).toBe('success');
      expect(await adapter.get('key1')).toEqual(new Uint8Array([1]));
      expect(await adapter.get('key2')).toEqual(new Uint8Array([2]));
    });

    it('should rollback on transaction error', async () => {
      await adapter.set('key1', new Uint8Array([1]));

      await expect(
        adapter.transaction(async () => {
          await adapter.set('key1', new Uint8Array([2]));
          await adapter.set('key2', new Uint8Array([3]));
          throw new Error('Transaction failed');
        })
      ).rejects.toThrow('Transaction failed');

      // Should be rolled back
      expect(await adapter.get('key1')).toEqual(new Uint8Array([1]));
      expect(await adapter.get('key2')).toBeNull();
    });
  });

  describe('Snapshot and restore', () => {
    it('should create snapshots', async () => {
      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));

      const snapshot = adapter.snapshot();

      expect(snapshot.size).toBe(2);
      expect(snapshot.get('key1')).toEqual(new Uint8Array([1]));
      expect(snapshot.get('key2')).toEqual(new Uint8Array([2]));
    });

    it('should restore from snapshots', async () => {
      await adapter.set('key1', new Uint8Array([1]));
      const snapshot = adapter.snapshot();

      await adapter.set('key2', new Uint8Array([2]));
      await adapter.delete('key1');

      adapter.restore(snapshot);

      expect(await adapter.get('key1')).toEqual(new Uint8Array([1]));
      expect(await adapter.get('key2')).toBeNull();
    });

    it('should get storage contents', async () => {
      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));

      const contents = adapter.getStorageContents();

      expect(contents.size).toBe(2);
      expect(contents.get('key1')).toEqual(new Uint8Array([1]));
    });

    it('should set storage contents', async () => {
      const contents = new Map<string, Uint8Array>();
      contents.set('key1', new Uint8Array([1]));
      contents.set('key2', new Uint8Array([2]));

      adapter.setStorageContents(contents);

      expect(await adapter.get('key1')).toEqual(new Uint8Array([1]));
      expect(await adapter.get('key2')).toEqual(new Uint8Array([2]));
    });
  });

  describe('Utility methods', () => {
    it('should get storage size', async () => {
      expect(adapter.getSize()).toBe(0);

      await adapter.set('key1', new Uint8Array([1]));
      expect(adapter.getSize()).toBe(1);

      await adapter.set('key2', new Uint8Array([2]));
      expect(adapter.getSize()).toBe(2);

      await adapter.delete('key1');
      expect(adapter.getSize()).toBe(1);
    });

    it('should reset adapter state', async () => {
      adapter.enableCallTracking();
      adapter.addFailKey('fail');
      adapter.setDelay(100);
      await adapter.set('key', new Uint8Array([1]));

      adapter.reset();

      expect(adapter.getSize()).toBe(0);
      expect(adapter.getCalls()).toHaveLength(0);
      await expect(adapter.get('fail')).resolves.toBeNull();
    });
  });

  describe('Constructor options', () => {
    it('should accept initial delay', async () => {
      const adapter = new MockAdapter({ delay: 50 });

      const start = Date.now();
      await adapter.get('test');
      const duration = Date.now() - start;

      // Allow small timing tolerance (2ms) for setTimeout precision
      expect(duration).toBeGreaterThanOrEqual(48);
    });

    it('should accept initial fail keys', async () => {
      const failKeys = new Set(['fail1', 'fail2']);
      const adapter = new MockAdapter({ failKeys });

      await expect(adapter.get('fail1')).rejects.toThrow();
      await expect(adapter.get('fail2')).rejects.toThrow();
    });

    it('should accept maxKeys option', async () => {
      const adapter = new MockAdapter({ maxKeys: 2 });

      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));

      await expect(adapter.set('key3', new Uint8Array([3]))).rejects.toThrow(
        'QuotaExceededError'
      );
    });

    it('should accept trackCalls option', async () => {
      const adapter = new MockAdapter({ trackCalls: true });

      await adapter.get('key');

      const calls = adapter.getCalls();
      expect(calls).toHaveLength(1);
    });
  });
});
