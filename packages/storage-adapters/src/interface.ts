/**
 * Common storage interface that all storage adapters must implement.
 * Provides a unified API for storing and retrieving binary data.
 */
export interface StorageAdapter {
  /**
   * Retrieve a value by key.
   *
   * @param key - The key to retrieve
   * @returns The stored value as a Uint8Array, or null if not found
   */
  get(key: string): Promise<Uint8Array | null>;

  /**
   * Store a value with the given key.
   *
   * @param key - The key to store under
   * @param value - The binary data to store
   */
  set(key: string, value: Uint8Array): Promise<void>;

  /**
   * Delete a value by key.
   *
   * @param key - The key to delete
   */
  delete(key: string): Promise<void>;

  /**
   * List all keys, optionally filtered by prefix.
   *
   * @param prefix - Optional prefix to filter keys
   * @returns Array of matching keys
   */
  list(prefix?: string): Promise<string[]>;

  /**
   * Check if a key exists in storage.
   *
   * @param key - The key to check
   * @returns True if the key exists, false otherwise
   */
  exists(key: string): Promise<boolean>;

  /**
   * Clear all data from storage.
   * WARNING: This will delete all data stored by this adapter.
   */
  clear(): Promise<void>;

  /**
   * Get the current storage quota information (if supported).
   *
   * @returns Storage quota information or null if not supported
   */
  getQuota?(): Promise<StorageQuota | null>;
}

/**
 * Storage quota information.
 */
export interface StorageQuota {
  /**
   * Current usage in bytes
   */
  usage: number;

  /**
   * Total available quota in bytes
   */
  quota: number;

  /**
   * Percentage of quota used (0-100)
   */
  percentage: number;
}

/**
 * Configuration options for storage adapters.
 */
export interface StorageAdapterOptions {
  /**
   * Name/prefix for this storage instance.
   * Used to namespace data and avoid conflicts.
   */
  name: string;

  /**
   * Optional custom serialization for special data types.
   */
  serializer?: {
    serialize: (data: Uint8Array) => unknown;
    deserialize: (data: unknown) => Uint8Array;
  };
}

/**
 * Error thrown when storage operations fail.
 */
export class StorageError extends Error {
  constructor(
    message: string,
    public readonly code: StorageErrorCode,
    public readonly cause?: unknown
  ) {
    super(message);
    this.name = 'StorageError';
  }
}

/**
 * Error codes for storage operations.
 */
export enum StorageErrorCode {
  NOT_FOUND = 'NOT_FOUND',
  QUOTA_EXCEEDED = 'QUOTA_EXCEEDED',
  NOT_SUPPORTED = 'NOT_SUPPORTED',
  PERMISSION_DENIED = 'PERMISSION_DENIED',
  OPERATION_FAILED = 'OPERATION_FAILED',
  INVALID_KEY = 'INVALID_KEY',
  INVALID_VALUE = 'INVALID_VALUE',
}
