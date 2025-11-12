import type {
  StorageAdapter,
  StorageAdapterOptions,
  StorageQuota,
} from './interface.js';
import { StorageError, StorageErrorCode } from './interface.js';

/**
 * LocalStorage adapter.
 * Provides fallback storage with ~5-10MB capacity using localStorage API.
 * Note: LocalStorage is synchronous but we maintain async interface for consistency.
 */
export class LocalStorageAdapter implements StorageAdapter {
  private prefix: string;
  private readonly maxSize = 5 * 1024 * 1024; // 5MB approximate limit

  constructor(name: string = 'browser-git') {
    this.prefix = `bg-${name}:`;
  }

  /**
   * Get the full key with prefix.
   */
  private getKey(key: string): string {
    return `${this.prefix}${key}`;
  }

  /**
   * Estimate the size of stored data in bytes.
   */
  private getUsageEstimate(): number {
    let total = 0;
    try {
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && key.startsWith(this.prefix)) {
          const value = localStorage.getItem(key);
          if (value) {
            // Rough estimate: 2 bytes per character (UTF-16)
            total += (key.length + value.length) * 2;
          }
        }
      }
    } catch {
      // Ignore errors
    }
    return total;
  }

  /**
   * Encode Uint8Array to base64 string for localStorage.
   */
  private encode(data: Uint8Array): string {
    const binary = Array.from(data)
      .map((byte) => String.fromCharCode(byte))
      .join('');
    return btoa(binary);
  }

  /**
   * Decode base64 string back to Uint8Array.
   */
  private decode(encoded: string): Uint8Array {
    const binary = atob(encoded);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }

  async get(key: string): Promise<Uint8Array | null> {
    try {
      if (typeof localStorage === 'undefined') {
        throw new StorageError(
          'localStorage is not available',
          StorageErrorCode.NOT_SUPPORTED
        );
      }

      const fullKey = this.getKey(key);
      const value = localStorage.getItem(fullKey);

      if (value === null) {
        return null;
      }

      return this.decode(value);
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
      }
      throw new StorageError(
        `Failed to get key: ${key}`,
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async set(key: string, value: Uint8Array): Promise<void> {
    try {
      if (typeof localStorage === 'undefined') {
        throw new StorageError(
          'localStorage is not available',
          StorageErrorCode.NOT_SUPPORTED
        );
      }

      // Check size before storing
      const encoded = this.encode(value);
      const estimatedSize = (this.getKey(key).length + encoded.length) * 2;

      if (this.getUsageEstimate() + estimatedSize > this.maxSize) {
        throw new StorageError(
          'Storage quota would be exceeded',
          StorageErrorCode.QUOTA_EXCEEDED
        );
      }

      const fullKey = this.getKey(key);
      localStorage.setItem(fullKey, encoded);
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
      }

      // Check for quota exceeded errors
      if (
        (error as {name?: string}).name === 'QuotaExceededError' ||
        (error as {name?: string}).name === 'NS_ERROR_DOM_QUOTA_REACHED'
      ) {
        throw new StorageError(
          'Storage quota exceeded',
          StorageErrorCode.QUOTA_EXCEEDED,
          error
        );
      }

      throw new StorageError(
        `Failed to set key: ${key}`,
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async delete(key: string): Promise<void> {
    try {
      if (typeof localStorage === 'undefined') {
        throw new StorageError(
          'localStorage is not available',
          StorageErrorCode.NOT_SUPPORTED
        );
      }

      const fullKey = this.getKey(key);
      localStorage.removeItem(fullKey);
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
      }
      throw new StorageError(
        `Failed to delete key: ${key}`,
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async list(prefix?: string): Promise<string[]> {
    try {
      if (typeof localStorage === 'undefined') {
        throw new StorageError(
          'localStorage is not available',
          StorageErrorCode.NOT_SUPPORTED
        );
      }

      const keys: string[] = [];
      const prefixLen = this.prefix.length;

      for (let i = 0; i < localStorage.length; i++) {
        const fullKey = localStorage.key(i);
        if (fullKey && fullKey.startsWith(this.prefix)) {
          const key = fullKey.substring(prefixLen);
          if (!prefix || key.startsWith(prefix)) {
            keys.push(key);
          }
        }
      }

      return keys;
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
      }
      throw new StorageError(
        'Failed to list keys',
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async exists(key: string): Promise<boolean> {
    try {
      if (typeof localStorage === 'undefined') {
        throw new StorageError(
          'localStorage is not available',
          StorageErrorCode.NOT_SUPPORTED
        );
      }

      const fullKey = this.getKey(key);
      return localStorage.getItem(fullKey) !== null;
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
      }
      throw new StorageError(
        `Failed to check existence of key: ${key}`,
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async clear(): Promise<void> {
    try {
      if (typeof localStorage === 'undefined') {
        throw new StorageError(
          'localStorage is not available',
          StorageErrorCode.NOT_SUPPORTED
        );
      }

      // Collect keys to delete
      const keysToDelete: string[] = [];
      for (let i = 0; i < localStorage.length; i++) {
        const key = localStorage.key(i);
        if (key && key.startsWith(this.prefix)) {
          keysToDelete.push(key);
        }
      }

      // Delete all collected keys
      for (const key of keysToDelete) {
        localStorage.removeItem(key);
      }
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
      }
      throw new StorageError(
        'Failed to clear storage',
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async getQuota(): Promise<StorageQuota | null> {
    // LocalStorage doesn't provide reliable quota information
    // Return estimated usage with assumed max size
    const usage = this.getUsageEstimate();
    const quota = this.maxSize;

    return {
      usage,
      quota,
      percentage: (usage / quota) * 100,
    };
  }

  /**
   * Check if localStorage is available in the current environment.
   */
  static isSupported(): boolean {
    try {
      if (typeof localStorage === 'undefined') {
        return false;
      }

      // Test if we can actually write to localStorage
      const testKey = '__bg_test__';
      localStorage.setItem(testKey, 'test');
      localStorage.removeItem(testKey);
      return true;
    } catch {
      return false;
    }
  }
}
