/**
 * @browser-git/storage-adapters
 *
 * Storage adapter implementations for BrowserGit.
 * Provides unified interface for IndexedDB, OPFS, LocalStorage, and in-memory storage.
 */

export type {
  StorageAdapter,
  StorageAdapterOptions,
  StorageQuota,
} from "./interface.js";

export { StorageError, StorageErrorCode } from "./interface.js";

// Storage adapter implementations
export { IndexedDBAdapter } from "./indexeddb.js";
export { OPFSAdapter } from "./opfs.js";
export { LocalStorageAdapter } from "./localstorage.js";
export { MemoryAdapter } from "./memory.js";
export { MockAdapter } from "./mock.js";
export type { MockStorageOptions, MockCall } from "./mock.js";

// Utilities
export * from "./utils/quota.js";
export * from "./utils/serialization.js";
