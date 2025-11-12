import type {
  StorageAdapter,
  
  StorageQuota,
} from './interface.js';
import { StorageError, StorageErrorCode } from './interface.js';

/**
 * IndexedDB storage adapter.
 * Provides persistent storage with large capacity using IndexedDB API.
 */
export class IndexedDBAdapter implements StorageAdapter {
  private dbName: string;
  private storeName = 'storage';
  private db: IDBDatabase | null = null;
  private initPromise: Promise<void> | null = null;

  constructor(name: string = 'browser-git') {
    this.dbName = `bg-${name}`;
  }

  /**
   * Initialize the IndexedDB database.
   */
  private async init(): Promise<void> {
    if (this.db) {
      return;
    }

    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = new Promise<void>((resolve, reject) => {
      if (typeof indexedDB === 'undefined') {
        reject(
          new StorageError(
            'IndexedDB is not available in this environment',
            StorageErrorCode.NOT_SUPPORTED
          )
        );
        return;
      }

      const request = indexedDB.open(this.dbName, 1);

      request.onerror = () => {
        reject(
          new StorageError(
            'Failed to open IndexedDB',
            StorageErrorCode.OPERATION_FAILED,
            request.error
          )
        );
      };

      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;

        // Create object store if it doesn't exist
        if (!db.objectStoreNames.contains(this.storeName)) {
          db.createObjectStore(this.storeName);
        }
      };
    });

    return this.initPromise;
  }

  /**
   * Get a transaction for the object store.
   */
  private async getTransaction(
    mode: IDBTransactionMode
  ): Promise<IDBObjectStore> {
    await this.init();

    if (!this.db) {
      throw new StorageError(
        'Database not initialized',
        StorageErrorCode.OPERATION_FAILED
      );
    }

    const transaction = this.db.transaction([this.storeName], mode);
    return transaction.objectStore(this.storeName);
  }

  async get(key: string): Promise<Uint8Array | null> {
    try {
      const store = await this.getTransaction('readonly');
      const request = store.get(key);

      return new Promise<Uint8Array | null>((resolve, reject) => {
        request.onsuccess = () => {
          const result = request.result;
          if (result === undefined) {
            resolve(null);
          } else {
            resolve(new Uint8Array(result));
          }
        };

        request.onerror = () => {
          reject(
            new StorageError(
              `Failed to get key: ${key}`,
              StorageErrorCode.OPERATION_FAILED,
              request.error
            )
          );
        };
      });
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
      const store = await this.getTransaction('readwrite');
      const request = store.put(value.buffer, key);

      return new Promise<void>((resolve, reject) => {
        request.onsuccess = () => {
          resolve();
        };

        request.onerror = () => {
          // Check for quota exceeded error
          if (
            request.error?.name === 'QuotaExceededError' ||
            request.error?.name === 'NS_ERROR_DOM_QUOTA_REACHED'
          ) {
            reject(
              new StorageError(
                'Storage quota exceeded',
                StorageErrorCode.QUOTA_EXCEEDED,
                request.error
              )
            );
          } else {
            reject(
              new StorageError(
                `Failed to set key: ${key}`,
                StorageErrorCode.OPERATION_FAILED,
                request.error
              )
            );
          }
        };
      });
    } catch (error) {
      if (error instanceof StorageError) {
        throw error;
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
      const store = await this.getTransaction('readwrite');
      const request = store.delete(key);

      return new Promise<void>((resolve, reject) => {
        request.onsuccess = () => {
          resolve();
        };

        request.onerror = () => {
          reject(
            new StorageError(
              `Failed to delete key: ${key}`,
              StorageErrorCode.OPERATION_FAILED,
              request.error
            )
          );
        };
      });
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
      const store = await this.getTransaction('readonly');
      const request = store.getAllKeys();

      return new Promise<string[]>((resolve, reject) => {
        request.onsuccess = () => {
          const keys = request.result as string[];

          if (prefix) {
            resolve(keys.filter((key) => key.startsWith(prefix)));
          } else {
            resolve(keys);
          }
        };

        request.onerror = () => {
          reject(
            new StorageError(
              'Failed to list keys',
              StorageErrorCode.OPERATION_FAILED,
              request.error
            )
          );
        };
      });
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
      const store = await this.getTransaction('readonly');
      const request = store.getKey(key);

      return new Promise<boolean>((resolve, reject) => {
        request.onsuccess = () => {
          resolve(request.result !== undefined);
        };

        request.onerror = () => {
          reject(
            new StorageError(
              `Failed to check existence of key: ${key}`,
              StorageErrorCode.OPERATION_FAILED,
              request.error
            )
          );
        };
      });
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
      const store = await this.getTransaction('readwrite');
      const request = store.clear();

      return new Promise<void>((resolve, reject) => {
        request.onsuccess = () => {
          resolve();
        };

        request.onerror = () => {
          reject(
            new StorageError(
              'Failed to clear storage',
              StorageErrorCode.OPERATION_FAILED,
              request.error
            )
          );
        };
      });
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
    if (!navigator.storage || !navigator.storage.estimate) {
      return null;
    }

    try {
      const estimate = await navigator.storage.estimate();
      const usage = estimate.usage || 0;
      const quota = estimate.quota || 0;

      return {
        usage,
        quota,
        percentage: quota > 0 ? (usage / quota) * 100 : 0,
      };
    } catch (error) {
      // Return null if quota estimation fails
      return null;
    }
  }

  /**
   * Close the database connection.
   * Should be called when done using the adapter.
   */
  close(): void {
    if (this.db) {
      this.db.close();
      this.db = null;
      this.initPromise = null;
    }
  }
}
