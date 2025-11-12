import type {
  StorageAdapter,
  StorageQuota,
} from './interface.js';

/**
 * In-memory storage adapter.
 * Fast, ephemeral storage for testing and temporary operations.
 * Data is lost when the adapter is garbage collected or the page is refreshed.
 */
export class MemoryAdapter implements StorageAdapter {
  private store: Map<string, Uint8Array> = new Map();

  async get(key: string): Promise<Uint8Array | null> {
    const value = this.store.get(key);
    return value ? new Uint8Array(value) : null; // Return a copy
  }

  async set(key: string, value: Uint8Array): Promise<void> {
    // Store a copy to prevent external modifications
    this.store.set(key, new Uint8Array(value));
  }

  async delete(key: string): Promise<void> {
    this.store.delete(key);
  }

  async list(prefix?: string): Promise<string[]> {
    const keys = Array.from(this.store.keys());
    if (prefix) {
      return keys.filter((key) => key.startsWith(prefix));
    }
    return keys;
  }

  async exists(key: string): Promise<boolean> {
    return this.store.has(key);
  }

  async clear(): Promise<void> {
    this.store.clear();
  }

  async getQuota(): Promise<StorageQuota | null> {
    // Calculate memory usage
    let usage = 0;
    for (const [key, value] of this.store) {
      // Rough estimate: key length + value length
      usage += key.length * 2 + value.byteLength; // 2 bytes per character for UTF-16 keys
    }

    // Memory storage has no hard quota, but we can report usage
    return {
      usage,
      quota: Number.MAX_SAFE_INTEGER,
      percentage: 0, // Effectively unlimited
    };
  }

  /**
   * Get the number of items stored.
   */
  size(): number {
    return this.store.size;
  }

  /**
   * Get all keys (synchronous convenience method for testing).
   */
  keys(): string[] {
    return Array.from(this.store.keys());
  }

  /**
   * Get all entries (synchronous convenience method for testing).
   */
  entries(): Array<[string, Uint8Array]> {
    return Array.from(this.store.entries()).map(([key, value]) => [
      key,
      new Uint8Array(value),
    ]);
  }
}
