/**
 * File system type definitions
 */

/**
 * Encoding types supported for text operations
 */
export type Encoding = 'utf8' | 'utf-8' | 'base64' | 'hex' | 'binary';

/**
 * File or directory statistics
 */
export interface Stats {
  /** Path to the file or directory */
  path: string;
  /** File size in bytes (0 for directories) */
  size: number;
  /** True if this is a directory */
  isDirectory: boolean;
  /** True if this is a file */
  isFile: boolean;
  /** Creation timestamp (ms since epoch) */
  ctimeMs: number;
  /** Last modification timestamp (ms since epoch) */
  mtimeMs: number;
  /** Last access timestamp (ms since epoch) */
  atimeMs: number;
}

/**
 * Directory entry information
 */
export interface Dirent {
  /** Entry name (without path) */
  name: string;
  /** True if this is a directory */
  isDirectory: boolean;
  /** True if this is a file */
  isFile: boolean;
}

/**
 * Options for mkdir operation
 */
export interface MkdirOptions {
  /** Create parent directories if they don't exist */
  recursive?: boolean;
}

/**
 * Options for rmdir operation
 */
export interface RmdirOptions {
  /** Remove directory and its contents recursively */
  recursive?: boolean;
}

/**
 * Options for readFile operation
 */
export interface ReadFileOptions {
  /** Text encoding (if omitted, returns Uint8Array) */
  encoding?: Encoding;
}

/**
 * Options for writeFile operation
 */
export interface WriteFileOptions {
  /** Text encoding (default: 'utf8') */
  encoding?: Encoding;
  /** Create parent directories if they don't exist */
  recursive?: boolean;
}

/**
 * File system change types
 */
export type ChangeType = 'create' | 'modify' | 'delete';

/**
 * File system change event
 */
export interface FSChangeEvent {
  /** Type of change */
  type: ChangeType;
  /** Path that changed */
  path: string;
  /** Timestamp of the change */
  timestamp: number;
}

/**
 * File system watcher callback
 */
export type FSWatchCallback = (event: FSChangeEvent) => void;

/**
 * File system watcher
 */
export interface FSWatcher {
  /** Stop watching for changes */
  close(): void;
}
