/**
 * Performance benchmarks for storage adapters
 */

import { bench, describe } from 'vitest';
import { MemoryAdapter, IndexedDBAdapter, LocalStorageAdapter } from '@browser-git/storage-adapters';

const SMALL_DATA = new Uint8Array(100); // 100 bytes
const MEDIUM_DATA = new Uint8Array(10 * 1024); // 10KB
const LARGE_DATA = new Uint8Array(100 * 1024); // 100KB

describe('Storage Adapters - Write Performance', () => {
  bench('MemoryAdapter - write small data', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', SMALL_DATA);
  });

  bench('MemoryAdapter - write medium data', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', MEDIUM_DATA);
  });

  bench('MemoryAdapter - write large data', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', LARGE_DATA);
  });

  bench('IndexedDBAdapter - write small data', async () => {
    const adapter = new IndexedDBAdapter('bench-idb-small');
    await adapter.set('test-key', SMALL_DATA);
    await adapter.clear();
  });

  bench('IndexedDBAdapter - write medium data', async () => {
    const adapter = new IndexedDBAdapter('bench-idb-medium');
    await adapter.set('test-key', MEDIUM_DATA);
    await adapter.clear();
  });

  bench('LocalStorageAdapter - write small data', async () => {
    const adapter = new LocalStorageAdapter('bench-ls-small-');
    await adapter.set('test-key', SMALL_DATA);
    await adapter.clear();
  });

  bench('LocalStorageAdapter - write medium data', async () => {
    const adapter = new LocalStorageAdapter('bench-ls-medium-');
    await adapter.set('test-key', MEDIUM_DATA);
    await adapter.clear();
  });
});

describe('Storage Adapters - Read Performance', () => {
  bench('MemoryAdapter - read small data', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', SMALL_DATA);
    await adapter.get('test-key');
  });

  bench('MemoryAdapter - read medium data', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', MEDIUM_DATA);
    await adapter.get('test-key');
  });

  bench('MemoryAdapter - read large data', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', LARGE_DATA);
    await adapter.get('test-key');
  });

  bench('IndexedDBAdapter - read small data', async () => {
    const adapter = new IndexedDBAdapter('bench-idb-read-small');
    await adapter.set('test-key', SMALL_DATA);
    await adapter.get('test-key');
    await adapter.clear();
  });

  bench('IndexedDBAdapter - read medium data', async () => {
    const adapter = new IndexedDBAdapter('bench-idb-read-medium');
    await adapter.set('test-key', MEDIUM_DATA);
    await adapter.get('test-key');
    await adapter.clear();
  });

  bench('LocalStorageAdapter - read small data', async () => {
    const adapter = new LocalStorageAdapter('bench-ls-read-small-');
    await adapter.set('test-key', SMALL_DATA);
    await adapter.get('test-key');
    await adapter.clear();
  });

  bench('LocalStorageAdapter - read medium data', async () => {
    const adapter = new LocalStorageAdapter('bench-ls-read-medium-');
    await adapter.set('test-key', MEDIUM_DATA);
    await adapter.get('test-key');
    await adapter.clear();
  });
});

describe('Storage Adapters - Bulk Operations', () => {
  bench('MemoryAdapter - write 100 keys', async () => {
    const adapter = new MemoryAdapter();
    for (let i = 0; i < 100; i++) {
      await adapter.set(`key-${i}`, SMALL_DATA);
    }
  });

  bench('MemoryAdapter - read 100 keys', async () => {
    const adapter = new MemoryAdapter();
    for (let i = 0; i < 100; i++) {
      await adapter.set(`key-${i}`, SMALL_DATA);
    }
    for (let i = 0; i < 100; i++) {
      await adapter.get(`key-${i}`);
    }
  });

  bench('IndexedDBAdapter - write 100 keys', async () => {
    const adapter = new IndexedDBAdapter('bench-idb-bulk-write');
    for (let i = 0; i < 100; i++) {
      await adapter.set(`key-${i}`, SMALL_DATA);
    }
    await adapter.clear();
  });

  bench('LocalStorageAdapter - write 100 keys', async () => {
    const adapter = new LocalStorageAdapter('bench-ls-bulk-write-');
    for (let i = 0; i < 100; i++) {
      await adapter.set(`key-${i}`, SMALL_DATA);
    }
    await adapter.clear();
  });
});

describe('Storage Adapters - Delete Performance', () => {
  bench('MemoryAdapter - delete key', async () => {
    const adapter = new MemoryAdapter();
    await adapter.set('test-key', SMALL_DATA);
    await adapter.delete('test-key');
  });

  bench('IndexedDBAdapter - delete key', async () => {
    const adapter = new IndexedDBAdapter('bench-idb-delete');
    await adapter.set('test-key', SMALL_DATA);
    await adapter.delete('test-key');
    await adapter.clear();
  });

  bench('LocalStorageAdapter - delete key', async () => {
    const adapter = new LocalStorageAdapter('bench-ls-delete-');
    await adapter.set('test-key', SMALL_DATA);
    await adapter.delete('test-key');
    await adapter.clear();
  });
});
