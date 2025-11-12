import type {
  StorageAdapter,
  StorageAdapterOptions,
  StorageQuota,
} from './interface.js';
import { StorageError, StorageErrorCode } from './interface.js';

/**
 * OPFS (Origin Private File System) storage adapter.
 * Uses the File System Access API for best performance in modern browsers.
 */
export class OPFSAdapter implements StorageAdapter {
  private rootDir: FileSystemDirectoryHandle | null = null;
  private initPromise: Promise<void> | null = null;
  private dirName: string;

  constructor(name: string = 'browser-git') {
    this.dirName = `bg-${name}`;
  }

  /**
   * Initialize the OPFS root directory.
   */
  private async init(): Promise<void> {
    if (this.rootDir) {
      return;
    }

    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = (async () => {
      try {
        // Check if OPFS is supported
        if (!navigator.storage || !navigator.storage.getDirectory) {
          throw new StorageError(
            'OPFS is not supported in this browser',
            StorageErrorCode.NOT_SUPPORTED
          );
        }

        // Get the origin private file system root
        const root = await navigator.storage.getDirectory();

        // Create or get our app directory
        this.rootDir = await root.getDirectoryHandle(this.dirName, {
          create: true,
        });
      } catch (error) {
        if (error instanceof StorageError) {
          throw error;
        }
        throw new StorageError(
          'Failed to initialize OPFS',
          StorageErrorCode.OPERATION_FAILED,
          error
        );
      }
    })();

    return this.initPromise;
  }

  /**
   * Get a file handle for a given key.
   */
  private async getFileHandle(
    key: string,
    create: boolean = false
  ): Promise<FileSystemFileHandle> {
    await this.init();

    if (!this.rootDir) {
      throw new StorageError(
        'OPFS not initialized',
        StorageErrorCode.OPERATION_FAILED
      );
    }

    try {
      // Handle nested paths (e.g., "dir/subdir/file.txt")
      const parts = key.split('/');
      const fileName = parts.pop()!;

      let currentDir = this.rootDir;

      // Navigate/create directory structure
      for (const part of parts) {
        if (part) {
          currentDir = await currentDir.getDirectoryHandle(part, {
            create,
          });
        }
      }

      // Get or create the file
      return await currentDir.getFileHandle(fileName, { create });
    } catch (error: unknown) {
      if ((error as {name?: string}).name === 'NotFoundError') {
        throw new StorageError(
          `Key not found: ${key}`,
          StorageErrorCode.NOT_FOUND,
          error
        );
      }
      throw new StorageError(
        `Failed to access file: ${key}`,
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async get(key: string): Promise<Uint8Array | null> {
    try {
      const fileHandle = await this.getFileHandle(key, false);
      const file = await fileHandle.getFile();
      const arrayBuffer = await file.arrayBuffer();
      return new Uint8Array(arrayBuffer);
    } catch (error) {
      if (
        error instanceof StorageError &&
        error.code === StorageErrorCode.NOT_FOUND
      ) {
        return null;
      }
      throw error;
    }
  }

  async set(key: string, value: Uint8Array): Promise<void> {
    try {
      const fileHandle = await this.getFileHandle(key, true);
      const writable = await fileHandle.createWritable();

      try {
        await writable.write(value);
        await writable.close();
      } catch (writeError) {
        // Try to close the writable even if write failed
        try {
          await writable.close();
        } catch {
          // Ignore close error
        }

        // Check for quota exceeded
        if ((writeError as {name?: string}).name === 'QuotaExceededError') {
          throw new StorageError(
            'Storage quota exceeded',
            StorageErrorCode.QUOTA_EXCEEDED,
            writeError
          );
        }

        throw writeError;
      }
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
    await this.init();

    if (!this.rootDir) {
      throw new StorageError(
        'OPFS not initialized',
        StorageErrorCode.OPERATION_FAILED
      );
    }

    try {
      const parts = key.split('/');
      const fileName = parts.pop()!;

      let currentDir = this.rootDir;

      // Navigate to parent directory
      for (const part of parts) {
        if (part) {
          currentDir = await currentDir.getDirectoryHandle(part, {
            create: false,
          });
        }
      }

      // Remove the file
      await currentDir.removeEntry(fileName);
    } catch (error) {
      if ((error as {name?: string}).name === 'NotFoundError') {
        // File doesn't exist, consider it deleted
        return;
      }
      throw new StorageError(
        `Failed to delete key: ${key}`,
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  async list(prefix?: string): Promise<string[]> {
    await this.init();

    if (!this.rootDir) {
      throw new StorageError(
        'OPFS not initialized',
        StorageErrorCode.OPERATION_FAILED
      );
    }

    const keys: string[] = [];

    try {
      await this.listRecursive(this.rootDir, '', keys, prefix);
      return keys;
    } catch (error) {
      throw new StorageError(
        'Failed to list keys',
        StorageErrorCode.OPERATION_FAILED,
        error
      );
    }
  }

  /**
   * Recursively list all files in a directory.
   */
  private async listRecursive(
    dir: FileSystemDirectoryHandle,
    path: string,
    keys: string[],
    prefix?: string
  ): Promise<void> {
    for await (const [name, handle] of dir.entries()) {
      const fullPath = path ? `${path}/${name}` : name;

      if (handle.kind === 'file') {
        if (!prefix || fullPath.startsWith(prefix)) {
          keys.push(fullPath);
        }
      } else if (handle.kind === 'directory') {
        // Recursively list subdirectory
        await this.listRecursive(handle, fullPath, keys, prefix);
      }
    }
  }

  async exists(key: string): Promise<boolean> {
    try {
      await this.getFileHandle(key, false);
      return true;
    } catch (error) {
      if (
        error instanceof StorageError &&
        error.code === StorageErrorCode.NOT_FOUND
      ) {
        return false;
      }
      throw error;
    }
  }

  async clear(): Promise<void> {
    await this.init();

    if (!this.rootDir) {
      throw new StorageError(
        'OPFS not initialized',
        StorageErrorCode.OPERATION_FAILED
      );
    }

    try {
      // Remove all entries in the root directory
      for await (const [name] of this.rootDir.entries()) {
        await this.rootDir.removeEntry(name, { recursive: true });
      }
    } catch (error) {
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
      return null;
    }
  }

  /**
   * Check if OPFS is supported in the current environment.
   */
  static isSupported(): boolean {
    return (
      typeof navigator !== 'undefined' &&
      'storage' in navigator &&
      typeof (navigator.storage as {getDirectory?: unknown}).getDirectory === 'function'
    );
  }
}
