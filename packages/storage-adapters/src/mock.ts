/**
 * Mock Storage Adapter for deterministic testing
 *
 * Provides controllable behavior for testing edge cases, errors, and timing
 */

import { StorageAdapter, StorageAdapterOptions } from './types';

export interface MockStorageOptions extends StorageAdapterOptions {
  /**
   * Simulate delays for operations (in milliseconds)
   */
  delay?: number;

  /**
   * Fail operations with specific keys
   */
  failKeys?: Set<string>;

  /**
   * Throw specific error for failed operations
   */
  failError?: Error;

  /**
   * Maximum number of keys before throwing quota exceeded
   */
  maxKeys?: number;

  /**
   * Track operation calls for testing
   */
  trackCalls?: boolean;
}

export interface MockCall {
  method: string;
  args: any[];
  timestamp: number;
}

/**
 * Mock storage adapter for testing with controllable behavior
 */
export class MockAdapter implements StorageAdapter {
  private storage: Map<string, Uint8Array> = new Map();
  private options: MockStorageOptions;
  private calls: MockCall[] = [];

  readonly name = 'mock';
  readonly supportsTransactions = true;
  readonly maxKeySize = 1024;
  readonly maxValueSize = 10 * 1024 * 1024; // 10MB

  constructor(options: MockStorageOptions = {}) {
    this.options = {
      delay: 0,
      failKeys: new Set(),
      maxKeys: Infinity,
      trackCalls: false,
      ...options,
    };
  }

  /**
   * Track a method call
   */
  private trackCall(method: string, ...args: any[]): void {
    if (this.options.trackCalls) {
      this.calls.push({
        method,
        args,
        timestamp: Date.now(),
      });
    }
  }

  /**
   * Get tracked calls
   */
  getCalls(): MockCall[] {
    return [...this.calls];
  }

  /**
   * Clear tracked calls
   */
  clearCalls(): void {
    this.calls = [];
  }

  /**
   * Simulate delay if configured
   */
  private async simulateDelay(): Promise<void> {
    if (this.options.delay && this.options.delay > 0) {
      await new Promise((resolve) => setTimeout(resolve, this.options.delay));
    }
  }

  /**
   * Check if operation should fail
   */
  private checkFailure(key: string): void {
    if (this.options.failKeys?.has(key)) {
      throw this.options.failError || new Error(`Mock failure for key: ${key}`);
    }
  }

  /**
   * Check quota
   */
  private checkQuota(): void {
    if (this.options.maxKeys && this.storage.size >= this.options.maxKeys) {
      throw new Error('QuotaExceededError: Storage quota exceeded');
    }
  }

  async get(key: string): Promise<Uint8Array | null> {
    this.trackCall('get', key);
    await this.simulateDelay();
    this.checkFailure(key);

    return this.storage.get(key) || null;
  }

  async set(key: string, value: Uint8Array): Promise<void> {
    this.trackCall('set', key, value);
    await this.simulateDelay();
    this.checkFailure(key);

    if (!this.storage.has(key)) {
      this.checkQuota();
    }

    this.storage.set(key, new Uint8Array(value));
  }

  async delete(key: string): Promise<void> {
    this.trackCall('delete', key);
    await this.simulateDelay();
    this.checkFailure(key);

    this.storage.delete(key);
  }

  async has(key: string): Promise<boolean> {
    this.trackCall('has', key);
    await this.simulateDelay();
    this.checkFailure(key);

    return this.storage.has(key);
  }

  async keys(): Promise<string[]> {
    this.trackCall('keys');
    await this.simulateDelay();

    return Array.from(this.storage.keys());
  }

  async clear(): Promise<void> {
    this.trackCall('clear');
    await this.simulateDelay();

    this.storage.clear();
  }

  async getQuota(): Promise<{ used: number; available: number }> {
    this.trackCall('getQuota');
    await this.simulateDelay();

    let used = 0;
    for (const [key, value] of this.storage.entries()) {
      used += key.length + value.length;
    }

    const maxKeys = this.options.maxKeys || Infinity;
    const available =
      maxKeys === Infinity ? Infinity : (maxKeys - this.storage.size) * 1024;

    return { used, available };
  }

  async transaction<T>(fn: () => Promise<T>): Promise<T> {
    this.trackCall('transaction');
    await this.simulateDelay();

    // Simple transaction: save state, run function, restore on error
    const backup = new Map(this.storage);

    try {
      return await fn();
    } catch (error) {
      this.storage = backup;
      throw error;
    }
  }

  /**
   * Mock-specific methods for testing
   */

  /**
   * Set delay for all operations
   */
  setDelay(delay: number): void {
    this.options.delay = delay;
  }

  /**
   * Add keys that should fail
   */
  addFailKey(key: string): void {
    if (!this.options.failKeys) {
      this.options.failKeys = new Set();
    }
    this.options.failKeys.add(key);
  }

  /**
   * Remove keys that should fail
   */
  removeFailKey(key: string): void {
    this.options.failKeys?.delete(key);
  }

  /**
   * Clear all fail keys
   */
  clearFailKeys(): void {
    this.options.failKeys?.clear();
  }

  /**
   * Set custom error for failures
   */
  setFailError(error: Error): void {
    this.options.failError = error;
  }

  /**
   * Set maximum number of keys
   */
  setMaxKeys(maxKeys: number): void {
    this.options.maxKeys = maxKeys;
  }

  /**
   * Enable call tracking
   */
  enableCallTracking(): void {
    this.options.trackCalls = true;
  }

  /**
   * Disable call tracking
   */
  disableCallTracking(): void {
    this.options.trackCalls = false;
  }

  /**
   * Get current storage size
   */
  getSize(): number {
    return this.storage.size;
  }

  /**
   * Get storage contents for inspection
   */
  getStorageContents(): Map<string, Uint8Array> {
    return new Map(this.storage);
  }

  /**
   * Restore storage contents
   */
  setStorageContents(contents: Map<string, Uint8Array>): void {
    this.storage = new Map(contents);
  }

  /**
   * Create a snapshot of current state
   */
  snapshot(): Map<string, Uint8Array> {
    return new Map(this.storage);
  }

  /**
   * Restore from a snapshot
   */
  restore(snapshot: Map<string, Uint8Array>): void {
    this.storage = new Map(snapshot);
  }

  /**
   * Reset adapter to initial state
   */
  reset(): void {
    this.storage.clear();
    this.calls = [];
    this.options.failKeys?.clear();
    this.options.failError = undefined;
  }
}
