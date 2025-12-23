/**
 * Node.js-like file system API for browser storage
 */

import type { StorageAdapter } from "@browser-git/storage-adapters";
import type {
  Stats,
  MkdirOptions,
  RmdirOptions,
  ReadFileOptions,
  WriteFileOptions,
  Encoding,
  FSWatcher,
  FSWatchCallback,
  FSChangeEvent,
} from "../types/fs.js";
import { normalize, dirname, join } from "./path.js";

/**
 * File system errors
 */
export class FSError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly path?: string,
  ) {
    super(message);
    this.name = "FSError";
  }

  static ENOENT(path: string): FSError {
    return new FSError(
      `ENOENT: no such file or directory, '${path}'`,
      "ENOENT",
      path,
    );
  }

  static EEXIST(path: string): FSError {
    return new FSError(
      `EEXIST: file already exists, '${path}'`,
      "EEXIST",
      path,
    );
  }

  static ENOTDIR(path: string): FSError {
    return new FSError(`ENOTDIR: not a directory, '${path}'`, "ENOTDIR", path);
  }

  static EISDIR(path: string): FSError {
    return new FSError(
      `EISDIR: illegal operation on a directory, '${path}'`,
      "EISDIR",
      path,
    );
  }

  static ENOTEMPTY(path: string): FSError {
    return new FSError(
      `ENOTEMPTY: directory not empty, '${path}'`,
      "ENOTEMPTY",
      path,
    );
  }

  static EINVAL(message: string): FSError {
    return new FSError(`EINVAL: ${message}`, "EINVAL");
  }
}

/**
 * Metadata stored for each file/directory
 */
interface FileMetadata {
  path: string;
  isDirectory: boolean;
  size: number;
  ctimeMs: number;
  mtimeMs: number;
  atimeMs: number;
}

/**
 * Internal storage keys
 */
const METADATA_PREFIX = "__meta__/";
const DATA_PREFIX = "__data__/";
const DIR_MARKER = "__dir__";

/**
 * File system implementation using storage adapter
 */
export class FileSystem {
  private watchers: Map<string, Set<FSWatchCallback>> = new Map();

  constructor(private storage: StorageAdapter) {}

  /**
   * Read file contents
   */
  async readFile(path: string): Promise<Uint8Array>;
  async readFile(
    path: string,
    options: { encoding: Encoding },
  ): Promise<string>;
  async readFile(
    path: string,
    options?: ReadFileOptions,
  ): Promise<Uint8Array | string> {
    const normalizedPath = normalize(path);
    const meta = await this.getMetadata(normalizedPath);

    if (!meta) {
      throw FSError.ENOENT(normalizedPath);
    }

    if (meta.isDirectory) {
      throw FSError.EISDIR(normalizedPath);
    }

    const dataKey = DATA_PREFIX + normalizedPath;
    const data = await this.storage.get(dataKey);

    if (!data) {
      throw FSError.ENOENT(normalizedPath);
    }

    // Update access time
    meta.atimeMs = Date.now();
    await this.setMetadata(normalizedPath, meta);

    if (options?.encoding) {
      return this.decode(data, options.encoding);
    }

    return data;
  }

  /**
   * Write file contents
   */
  async writeFile(
    path: string,
    data: string | Uint8Array,
    options?: WriteFileOptions,
  ): Promise<void> {
    const normalizedPath = normalize(path);
    const dirPath = dirname(normalizedPath);

    // Create parent directories if recursive option is set
    if (options?.recursive && dirPath !== "." && dirPath !== "/") {
      await this.mkdir(dirPath, { recursive: true });
    }

    // Check if parent directory exists
    if (dirPath !== "." && dirPath !== "/") {
      const parentMeta = await this.getMetadata(dirPath);
      if (!parentMeta) {
        throw FSError.ENOENT(dirPath);
      }
      if (!parentMeta.isDirectory) {
        throw FSError.ENOTDIR(dirPath);
      }
    }

    // Check if path is a directory
    const existingMeta = await this.getMetadata(normalizedPath);
    if (existingMeta?.isDirectory) {
      throw FSError.EISDIR(normalizedPath);
    }

    // Encode data if string
    const bytes =
      typeof data === "string"
        ? this.encode(data, options?.encoding || "utf8")
        : data;

    // Write data
    const dataKey = DATA_PREFIX + normalizedPath;
    await this.storage.set(dataKey, bytes);

    // Update or create metadata
    const now = Date.now();
    const meta: FileMetadata = existingMeta
      ? { ...existingMeta, size: bytes.length, mtimeMs: now, atimeMs: now }
      : {
          path: normalizedPath,
          isDirectory: false,
          size: bytes.length,
          ctimeMs: now,
          mtimeMs: now,
          atimeMs: now,
        };

    await this.setMetadata(normalizedPath, meta);

    // Emit change event
    this.emitChange({
      type: existingMeta ? "modify" : "create",
      path: normalizedPath,
      timestamp: now,
    });
  }

  /**
   * Create directory
   */
  async mkdir(path: string, options?: MkdirOptions): Promise<void> {
    const normalizedPath = normalize(path);

    // Check if already exists
    const existingMeta = await this.getMetadata(normalizedPath);
    if (existingMeta) {
      if (!options?.recursive) {
        throw FSError.EEXIST(normalizedPath);
      }
      return;
    }

    // Create parent directories if recursive
    if (options?.recursive) {
      const parentPath = dirname(normalizedPath);
      if (
        parentPath !== "." &&
        parentPath !== "/" &&
        parentPath !== normalizedPath
      ) {
        const parentMeta = await this.getMetadata(parentPath);
        if (!parentMeta) {
          await this.mkdir(parentPath, { recursive: true });
        }
      }
    } else {
      // Check parent exists
      const parentPath = dirname(normalizedPath);
      if (parentPath !== "." && parentPath !== "/") {
        const parentMeta = await this.getMetadata(parentPath);
        if (!parentMeta) {
          throw FSError.ENOENT(parentPath);
        }
        if (!parentMeta.isDirectory) {
          throw FSError.ENOTDIR(parentPath);
        }
      }
    }

    // Create directory marker
    const dataKey = DATA_PREFIX + normalizedPath + "/" + DIR_MARKER;
    await this.storage.set(dataKey, new Uint8Array(0));

    // Create metadata
    const now = Date.now();
    const meta: FileMetadata = {
      path: normalizedPath,
      isDirectory: true,
      size: 0,
      ctimeMs: now,
      mtimeMs: now,
      atimeMs: now,
    };

    await this.setMetadata(normalizedPath, meta);

    // Emit change event
    this.emitChange({
      type: "create",
      path: normalizedPath,
      timestamp: now,
    });
  }

  /**
   * Read directory contents
   */
  async readdir(path: string): Promise<string[]> {
    const normalizedPath = normalize(path);
    const meta = await this.getMetadata(normalizedPath);

    if (!meta) {
      throw FSError.ENOENT(normalizedPath);
    }

    if (!meta.isDirectory) {
      throw FSError.ENOTDIR(normalizedPath);
    }

    // List all keys with data prefix
    const prefix = DATA_PREFIX + normalizedPath + "/";
    const allKeys = await this.storage.list(DATA_PREFIX);

    // Filter keys that are direct children
    const children = new Set<string>();
    for (const key of allKeys) {
      if (key.startsWith(prefix)) {
        const relative = key.substring(prefix.length);
        // Skip directory markers
        if (relative === DIR_MARKER) continue;

        // Get first path segment
        const parts = relative.split("/");
        if (parts[0]) {
          children.add(parts[0]);
        }
      }
    }

    // Update access time
    meta.atimeMs = Date.now();
    await this.setMetadata(normalizedPath, meta);

    return Array.from(children).sort();
  }

  /**
   * Delete file
   */
  async unlink(path: string): Promise<void> {
    const normalizedPath = normalize(path);
    const meta = await this.getMetadata(normalizedPath);

    if (!meta) {
      throw FSError.ENOENT(normalizedPath);
    }

    if (meta.isDirectory) {
      throw FSError.EISDIR(normalizedPath);
    }

    // Delete data and metadata
    const dataKey = DATA_PREFIX + normalizedPath;
    const metaKey = METADATA_PREFIX + normalizedPath;

    await this.storage.delete(dataKey);
    await this.storage.delete(metaKey);

    // Emit change event
    this.emitChange({
      type: "delete",
      path: normalizedPath,
      timestamp: Date.now(),
    });
  }

  /**
   * Remove directory
   */
  async rmdir(path: string, options?: RmdirOptions): Promise<void> {
    const normalizedPath = normalize(path);
    const meta = await this.getMetadata(normalizedPath);

    if (!meta) {
      throw FSError.ENOENT(normalizedPath);
    }

    if (!meta.isDirectory) {
      throw FSError.ENOTDIR(normalizedPath);
    }

    // Check if directory is empty (unless recursive)
    if (!options?.recursive) {
      const children = await this.readdir(normalizedPath);
      if (children.length > 0) {
        throw FSError.ENOTEMPTY(normalizedPath);
      }
    } else {
      // Remove all children recursively
      const children = await this.readdir(normalizedPath);
      for (const child of children) {
        const childPath = join(normalizedPath, child);
        const childMeta = await this.getMetadata(childPath);
        if (childMeta?.isDirectory) {
          await this.rmdir(childPath, { recursive: true });
        } else {
          await this.unlink(childPath);
        }
      }
    }

    // Delete directory marker and metadata
    const dataKey = DATA_PREFIX + normalizedPath + "/" + DIR_MARKER;
    const metaKey = METADATA_PREFIX + normalizedPath;

    await this.storage.delete(dataKey);
    await this.storage.delete(metaKey);

    // Emit change event
    this.emitChange({
      type: "delete",
      path: normalizedPath,
      timestamp: Date.now(),
    });
  }

  /**
   * Get file/directory stats
   */
  async stat(path: string): Promise<Stats> {
    const normalizedPath = normalize(path);
    const meta = await this.getMetadata(normalizedPath);

    if (!meta) {
      throw FSError.ENOENT(normalizedPath);
    }

    // Update access time
    meta.atimeMs = Date.now();
    await this.setMetadata(normalizedPath, meta);

    return {
      path: meta.path,
      size: meta.size,
      isDirectory: meta.isDirectory,
      isFile: !meta.isDirectory,
      ctimeMs: meta.ctimeMs,
      mtimeMs: meta.mtimeMs,
      atimeMs: meta.atimeMs,
    };
  }

  /**
   * Check if file/directory exists
   */
  async exists(path: string): Promise<boolean> {
    const normalizedPath = normalize(path);
    const meta = await this.getMetadata(normalizedPath);
    return meta !== null;
  }

  /**
   * Watch for file changes
   */
  watch(path: string, callback: FSWatchCallback): FSWatcher {
    const normalizedPath = normalize(path);

    if (!this.watchers.has(normalizedPath)) {
      this.watchers.set(normalizedPath, new Set());
    }

    this.watchers.get(normalizedPath)!.add(callback);

    return {
      close: () => {
        const callbacks = this.watchers.get(normalizedPath);
        if (callbacks) {
          callbacks.delete(callback);
          if (callbacks.size === 0) {
            this.watchers.delete(normalizedPath);
          }
        }
      },
    };
  }

  /**
   * Get metadata for a path
   */
  private async getMetadata(path: string): Promise<FileMetadata | null> {
    const metaKey = METADATA_PREFIX + path;
    const data = await this.storage.get(metaKey);

    if (!data) {
      return null;
    }

    const json = new TextDecoder().decode(data);
    return JSON.parse(json) as FileMetadata;
  }

  /**
   * Set metadata for a path
   */
  private async setMetadata(path: string, meta: FileMetadata): Promise<void> {
    const metaKey = METADATA_PREFIX + path;
    const json = JSON.stringify(meta);
    const data = new TextEncoder().encode(json);
    await this.storage.set(metaKey, data);
  }

  /**
   * Emit change event to watchers
   */
  private emitChange(event: FSChangeEvent): void {
    const callbacks = this.watchers.get(event.path);
    if (callbacks) {
      for (const callback of callbacks) {
        try {
          callback(event);
        } catch (error) {
          console.error("Error in file watcher callback:", error);
        }
      }
    }
  }

  /**
   * Encode string to bytes
   */
  private encode(str: string, encoding: Encoding): Uint8Array {
    switch (encoding) {
      case "utf8":
      case "utf-8":
        return new TextEncoder().encode(str);

      case "base64": {
        const binary = atob(str);
        const bytes = new Uint8Array(binary.length);
        for (let i = 0; i < binary.length; i++) {
          bytes[i] = binary.charCodeAt(i);
        }
        return bytes;
      }

      case "hex": {
        const bytes = new Uint8Array(str.length / 2);
        for (let i = 0; i < bytes.length; i++) {
          bytes[i] = parseInt(str.substring(i * 2, i * 2 + 2), 16);
        }
        return bytes;
      }

      case "binary": {
        const bytes = new Uint8Array(str.length);
        for (let i = 0; i < str.length; i++) {
          bytes[i] = str.charCodeAt(i) & 0xff;
        }
        return bytes;
      }

      default:
        throw FSError.EINVAL(`Unknown encoding: ${encoding}`);
    }
  }

  /**
   * Decode bytes to string
   */
  private decode(bytes: Uint8Array, encoding: Encoding): string {
    switch (encoding) {
      case "utf8":
      case "utf-8":
        return new TextDecoder().decode(bytes);

      case "base64": {
        const binary = Array.from(bytes)
          .map((byte) => String.fromCharCode(byte))
          .join("");
        return btoa(binary);
      }

      case "hex":
        return Array.from(bytes)
          .map((byte) => byte.toString(16).padStart(2, "0"))
          .join("");

      case "binary":
        return Array.from(bytes)
          .map((byte) => String.fromCharCode(byte))
          .join("");

      default:
        throw FSError.EINVAL(`Unknown encoding: ${encoding}`);
    }
  }
}
