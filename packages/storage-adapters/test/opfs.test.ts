import { describe, it, expect, beforeEach, vi } from 'vitest';
import { OPFSAdapter } from '../src/opfs.js';
import { StorageError, StorageErrorCode } from '../src/interface.js';

// Mock OPFS for testing since it's not available in jsdom
class MockFileSystemFileHandle {
  private data: ArrayBuffer = new ArrayBuffer(0);

  async getFile() {
    return {
      arrayBuffer: async () => this.data,
    };
  }

  async createWritable() {
    const self = this;
    return {
      data: new Uint8Array(0),
      async write(value: Uint8Array) {
        self.data = value.buffer;
      },
      async close() {
        // noop
      },
    };
  }
}

class MockFileSystemDirectoryHandle {
  private files = new Map<string, MockFileSystemFileHandle>();
  private dirs = new Map<string, MockFileSystemDirectoryHandle>();

  async getFileHandle(name: string, options?: { create?: boolean }) {
    if (!this.files.has(name)) {
      if (options?.create) {
        const handle = new MockFileSystemFileHandle();
        this.files.set(name, handle);
        return handle;
      }
      const error = new Error('NotFoundError');
      (error as {name?: string}).name = 'NotFoundError';
      throw error;
    }
    return this.files.get(name)!;
  }

  async getDirectoryHandle(name: string, options?: { create?: boolean }) {
    if (!this.dirs.has(name)) {
      if (options?.create) {
        const handle = new MockFileSystemDirectoryHandle();
        this.dirs.set(name, handle);
        return handle;
      }
      const error = new Error('NotFoundError');
      (error as {name?: string}).name = 'NotFoundError';
      throw error;
    }
    return this.dirs.get(name)!;
  }

  async removeEntry(name: string, options?: { recursive?: boolean }) {
    this.files.delete(name);
    this.dirs.delete(name);
  }

  async *entries(): AsyncIterableIterator<[string, MockFileSystemFileHandle | MockFileSystemDirectoryHandle]> {
    for (const [name, handle] of this.files) {
      yield [name, Object.assign(handle, { kind: 'file' })];
    }
    for (const [name, handle] of this.dirs) {
      yield [name, Object.assign(handle, { kind: 'directory' })];
    }
  }
}

describe('OPFSAdapter', () => {
  let adapter: OPFSAdapter;
  let mockRoot: MockFileSystemDirectoryHandle;

  beforeEach(() => {
    // Mock navigator.storage.getDirectory
    mockRoot = new MockFileSystemDirectoryHandle();

    globalThis.navigator = globalThis.navigator || {} as Navigator;
    (globalThis.navigator as {storage?: {getDirectory?: () => Promise<MockFileSystemDirectoryHandle>}}).storage = {
      getDirectory: vi.fn(async () => mockRoot),
    };

    adapter = new OPFSAdapter('test-db');
  });

  describe('isSupported', () => {
    it('should check if OPFS is supported', () => {
      // Should be true since we mocked it
      expect(OPFSAdapter.isSupported()).toBe(true);
    });
  });

  describe('basic operations', () => {
    it('should store and retrieve data', async () => {
      const key = 'test-key';
      const value = new Uint8Array([1, 2, 3, 4, 5]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });

    it('should return null for non-existent key', async () => {
      const result = await adapter.get('non-existent');
      expect(result).toBeNull();
    });

    it('should delete data', async () => {
      const key = 'test-key';
      const value = new Uint8Array([1, 2, 3]);

      await adapter.set(key, value);
      await adapter.delete(key);

      const retrieved = await adapter.get(key);
      expect(retrieved).toBeNull();
    });

    it('should check if key exists', async () => {
      const key = 'test-key';
      const value = new Uint8Array([1, 2, 3]);

      expect(await adapter.exists(key)).toBe(false);

      await adapter.set(key, value);
      expect(await adapter.exists(key)).toBe(true);

      await adapter.delete(key);
      expect(await adapter.exists(key)).toBe(false);
    });
  });

  describe('nested paths', () => {
    it('should handle nested directory paths', async () => {
      const key = 'dir1/dir2/file.txt';
      const value = new Uint8Array([1, 2, 3]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });

    it('should list files in nested directories', async () => {
      await adapter.set('file1.txt', new Uint8Array([1]));
      await adapter.set('dir1/file2.txt', new Uint8Array([2]));
      await adapter.set('dir1/dir2/file3.txt', new Uint8Array([3]));

      const keys = await adapter.list();
      expect(keys).toContain('file1.txt');
      expect(keys).toContain('dir1/file2.txt');
      expect(keys).toContain('dir1/dir2/file3.txt');
    });

    it('should delete nested files', async () => {
      const key = 'dir/subdir/file.txt';
      await adapter.set(key, new Uint8Array([1]));

      expect(await adapter.exists(key)).toBe(true);

      await adapter.delete(key);
      expect(await adapter.exists(key)).toBe(false);
    });
  });

  describe('list operations', () => {
    beforeEach(async () => {
      await adapter.set('file1.txt', new Uint8Array([1]));
      await adapter.set('file2.txt', new Uint8Array([2]));
      await adapter.set('dir/file3.txt', new Uint8Array([3]));
      await adapter.set('other.dat', new Uint8Array([4]));
    });

    it('should list all keys', async () => {
      const keys = await adapter.list();
      expect(keys.length).toBeGreaterThanOrEqual(4);
      expect(keys).toContain('file1.txt');
      expect(keys).toContain('file2.txt');
      expect(keys).toContain('dir/file3.txt');
      expect(keys).toContain('other.dat');
    });

    it('should list keys with prefix', async () => {
      const keys = await adapter.list('file');
      expect(keys).toContain('file1.txt');
      expect(keys).toContain('file2.txt');
      expect(keys).not.toContain('other.dat');
    });

    it('should list keys with directory prefix', async () => {
      const keys = await adapter.list('dir/');
      expect(keys).toContain('dir/file3.txt');
      expect(keys).not.toContain('file1.txt');
    });
  });

  describe('clear operation', () => {
    it('should clear all data', async () => {
      await adapter.set('key1', new Uint8Array([1]));
      await adapter.set('key2', new Uint8Array([2]));
      await adapter.set('dir/key3', new Uint8Array([3]));

      await adapter.clear();

      const keys = await adapter.list();
      expect(keys).toHaveLength(0);
    });
  });

  describe('binary data', () => {
    it('should handle empty Uint8Array', async () => {
      const key = 'empty';
      const value = new Uint8Array([]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
      expect(retrieved?.length).toBe(0);
    });

    it('should handle large binary data', async () => {
      const key = 'large';
      const value = new Uint8Array(10000);
      for (let i = 0; i < value.length; i++) {
        value[i] = i % 256;
      }

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });

    it('should preserve binary data integrity', async () => {
      const key = 'binary';
      const value = new Uint8Array([0, 255, 128, 1, 254, 127]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });
  });
});
