/**
 * Repository API for browser-git
 *
 * High-level TypeScript API for Git repository operations
 */

import type { AuthConfig } from './types/auth.js';

/**
 * Progress callback function
 */
export type ProgressCallback = (message: string) => void;

/**
 * Clone options
 */
export interface CloneOptions {
  /**
   * Create a bare repository (no working directory)
   * @default false
   */
  bare?: boolean;

  /**
   * Depth for shallow clone (0 for full clone)
   * @default 0
   */
  depth?: number;

  /**
   * Specific branch to clone (empty for default branch)
   * @default ""
   */
  branch?: string;

  /**
   * Name of the remote
   * @default "origin"
   */
  remote?: string;

  /**
   * Authentication configuration
   */
  auth?: AuthConfig;

  /**
   * Progress callback
   */
  onProgress?: ProgressCallback;
}

/**
 * Fetch options
 */
export interface FetchOptions {
  /**
   * Name of the remote to fetch from
   * @default "origin"
   */
  remote?: string;

  /**
   * Refspecs to fetch (default: fetch all branches)
   * @default []
   */
  refspecs?: string[];

  /**
   * Prune remote tracking branches that no longer exist
   * @default false
   */
  prune?: boolean;

  /**
   * Force non-fast-forward updates
   * @default false
   */
  force?: boolean;

  /**
   * Depth for shallow fetch (0 for full fetch)
   * @default 0
   */
  depth?: number;

  /**
   * Authentication configuration
   */
  auth?: AuthConfig;

  /**
   * Progress callback
   */
  onProgress?: ProgressCallback;
}

/**
 * Pull options
 */
export interface PullOptions {
  /**
   * Name of the remote to pull from
   * @default "origin"
   */
  remote?: string;

  /**
   * Specific branch to pull (empty for current branch's upstream)
   * @default ""
   */
  branch?: string;

  /**
   * Rebase instead of merge
   * @default false
   */
  rebase?: boolean;

  /**
   * Only allow fast-forward merges
   * @default false
   */
  fastForwardOnly?: boolean;

  /**
   * Force non-fast-forward updates during fetch
   * @default false
   */
  force?: boolean;

  /**
   * Authentication configuration
   */
  auth?: AuthConfig;

  /**
   * Progress callback
   */
  onProgress?: ProgressCallback;
}

/**
 * Push options
 */
export interface PushOptions {
  /**
   * Name of the remote to push to
   * @default "origin"
   */
  remote?: string;

  /**
   * Refspecs to push (e.g., "refs/heads/main:refs/heads/main")
   * If empty, pushes current branch to remote
   * @default []
   */
  refSpecs?: string[];

  /**
   * Allow non-fast-forward updates
   * @default false
   */
  force?: boolean;

  /**
   * Authentication configuration
   */
  auth?: AuthConfig;

  /**
   * Progress callback
   */
  onProgress?: ProgressCallback;
}

/**
 * Reference update information
 */
export interface RefUpdate {
  /** Reference name */
  refName: string;
  /** Previous hash (empty if new) */
  oldHash: string;
  /** New hash (empty if deleted) */
  newHash: string;
  /** Whether this was a forced update */
  forced: boolean;
}

/**
 * Fetch result
 */
export interface FetchResult {
  /** Updated references */
  updatedRefs: Record<string, RefUpdate>;
  /** Pruned references */
  prunedRefs: string[];
  /** Number of objects fetched */
  objectCount: number;
}

/**
 * Pull result
 */
export interface PullResult {
  /** Fetch operation result */
  fetchResult: FetchResult;
  /** Merge operation result (if any) */
  mergeResult?: any;
  /** Whether this was a fast-forward update */
  fastForward: boolean;
  /** Whether already up to date */
  alreadyUpToDate: boolean;
}

/**
 * Repository init options
 */
export interface InitOptions {
  /**
   * Create a bare repository
   * @default false
   */
  bare?: boolean;

  /**
   * Initial branch name
   * @default "main"
   */
  initialBranch?: string;

  /**
   * Hash algorithm to use
   * @default "sha1"
   */
  hashAlgorithm?: 'sha1' | 'sha256';
}

/**
 * Repository class
 */
export class Repository {
  private wasmInstance: any;

  constructor(private path: string, wasmInstance: any) {
    this.wasmInstance = wasmInstance;
  }

  /**
   * Clone a remote repository
   *
   * @param url - Repository URL to clone from
   * @param path - Local path to clone into
   * @param options - Clone options
   * @returns Promise resolving to a Repository instance
   *
   * @example
   * ```ts
   * const repo = await Repository.clone('https://github.com/user/repo.git', './my-repo', {
   *   auth: {
   *     method: AuthMethod.Token,
   *     token: 'ghp_xxxxxxxxxxxx'
   *   },
   *   onProgress: (msg) => console.log(msg)
   * });
   * ```
   */
  static async clone(
    url: string,
    path: string,
    options: CloneOptions = {}
  ): Promise<Repository> {
    // Load WASM if not already loaded
    const wasm = await loadWASM();

    // Prepare clone options
    const cloneOpts = {
      bare: options.bare ?? false,
      depth: options.depth ?? 0,
      branch: options.branch ?? '',
      remote: options.remote ?? 'origin',
      auth: options.auth,
    };

    // Create progress callback wrapper
    const progressCallback = options.onProgress
      ? (msg: string) => options.onProgress!(msg)
      : undefined;

    // Call WASM clone function
    try {
      await wasm.clone(url, path, cloneOpts, progressCallback);
      return new Repository(path, wasm);
    } catch (error) {
      throw new CloneError(
        `Failed to clone repository: ${error instanceof Error ? error.message : String(error)}`,
        url,
        error
      );
    }
  }

  /**
   * Initialize a new repository
   *
   * @param path - Path to initialize repository in
   * @param options - Init options
   * @returns Promise resolving to a Repository instance
   *
   * @example
   * ```ts
   * const repo = await Repository.init('./my-repo', {
   *   initialBranch: 'main'
   * });
   * ```
   */
  static async init(
    path: string,
    options: InitOptions = {}
  ): Promise<Repository> {
    const wasm = await loadWASM();

    const initOpts = {
      bare: options.bare ?? false,
      initialBranch: options.initialBranch ?? 'main',
      hashAlgorithm: options.hashAlgorithm ?? 'sha1',
    };

    try {
      await wasm.init(path, initOpts);
      return new Repository(path, wasm);
    } catch (error) {
      throw new GitError(
        `Failed to initialize repository: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Open an existing repository
   *
   * @param path - Path to the repository
   * @returns Promise resolving to a Repository instance
   *
   * @example
   * ```ts
   * const repo = await Repository.open('./my-repo');
   * ```
   */
  static async open(path: string): Promise<Repository> {
    const wasm = await loadWASM();

    try {
      await wasm.open(path);
      return new Repository(path, wasm);
    } catch (error) {
      throw new GitError(
        `Failed to open repository: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Get the repository path
   */
  getPath(): string {
    return this.path;
  }

  /**
   * Get the current branch name
   */
  async getCurrentBranch(): Promise<string> {
    try {
      return await this.wasmInstance.getCurrentBranch(this.path);
    } catch (error) {
      throw new GitError(
        `Failed to get current branch: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * List all branches
   */
  async listBranches(): Promise<string[]> {
    try {
      return await this.wasmInstance.listBranches(this.path);
    } catch (error) {
      throw new GitError(
        `Failed to list branches: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Create a new branch
   */
  async createBranch(name: string, startPoint?: string): Promise<void> {
    try {
      await this.wasmInstance.createBranch(this.path, name, startPoint);
    } catch (error) {
      throw new GitError(
        `Failed to create branch: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Checkout a branch or commit
   */
  async checkout(target: string): Promise<void> {
    try {
      await this.wasmInstance.checkout(this.path, target);
    } catch (error) {
      throw new GitError(
        `Failed to checkout: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Fetch objects and refs from a remote repository
   *
   * @param options - Fetch options
   * @returns Promise resolving to fetch result
   *
   * @example
   * ```ts
   * const result = await repo.fetch({
   *   remote: 'origin',
   *   prune: true,
   *   auth: {
   *     method: AuthMethod.Token,
   *     token: 'ghp_xxxxxxxxxxxx'
   *   },
   *   onProgress: (msg) => console.log(msg)
   * });
   * console.log(`Fetched ${result.objectCount} objects`);
   * ```
   */
  async fetch(options: FetchOptions = {}): Promise<FetchResult> {
    const fetchOpts = {
      remote: options.remote ?? 'origin',
      refspecs: options.refspecs ?? [],
      prune: options.prune ?? false,
      force: options.force ?? false,
      depth: options.depth ?? 0,
      auth: options.auth,
    };

    // Create progress callback wrapper
    const progressCallback = options.onProgress
      ? (msg: string) => options.onProgress!(msg)
      : undefined;

    try {
      return await this.wasmInstance.fetch(this.path, fetchOpts, progressCallback);
    } catch (error) {
      throw new GitError(
        `Failed to fetch: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Pull changes from a remote repository and integrate them into the current branch
   *
   * @param options - Pull options
   * @returns Promise resolving to pull result
   *
   * @example
   * ```ts
   * const result = await repo.pull({
   *   remote: 'origin',
   *   branch: 'main',
   *   auth: {
   *     method: AuthMethod.Token,
   *     token: 'ghp_xxxxxxxxxxxx'
   *   },
   *   onProgress: (msg) => console.log(msg)
   * });
   *
   * if (result.alreadyUpToDate) {
   *   console.log('Already up to date');
   * } else if (result.fastForward) {
   *   console.log('Fast-forwarded');
   * } else {
   *   console.log('Merged changes');
   * }
   * ```
   */
  async pull(options: PullOptions = {}): Promise<PullResult> {
    const pullOpts = {
      remote: options.remote ?? 'origin',
      branch: options.branch ?? '',
      rebase: options.rebase ?? false,
      fastForwardOnly: options.fastForwardOnly ?? false,
      force: options.force ?? false,
      auth: options.auth,
    };

    // Create progress callback wrapper
    const progressCallback = options.onProgress
      ? (msg: string) => options.onProgress!(msg)
      : undefined;

    try {
      return await this.wasmInstance.pull(this.path, pullOpts, progressCallback);
    } catch (error) {
      throw new GitError(
        `Failed to pull: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }

  /**
   * Push local commits to a remote repository
   *
   * @param options - Push options
   *
   * @example
   * ```ts
   * // Push current branch to origin
   * await repo.push({
   *   auth: {
   *     method: AuthMethod.Token,
   *     token: 'ghp_xxxxxxxxxxxx'
   *   },
   *   onProgress: (msg) => console.log(msg)
   * });
   *
   * // Force push a specific branch
   * await repo.push({
   *   refSpecs: ['refs/heads/feature:refs/heads/feature'],
   *   force: true,
   *   auth: { method: AuthMethod.Token, token: 'ghp_xxxxxxxxxxxx' }
   * });
   *
   * // Delete a remote branch
   * await repo.push({
   *   refSpecs: [':refs/heads/old-branch'],
   *   auth: { method: AuthMethod.Token, token: 'ghp_xxxxxxxxxxxx' }
   * });
   * ```
   */
  async push(options: PushOptions = {}): Promise<void> {
    try {
      // Prepare push options
      const pushOpts = {
        remote: options.remote ?? 'origin',
        refSpecs: options.refSpecs ?? [],
        force: options.force ?? false,
        auth: options.auth,
      };

      // Create progress callback wrapper
      const progressCallback = options.onProgress
        ? (msg: string) => options.onProgress!(msg)
        : undefined;

      // Call WASM push function
      await this.wasmInstance.push(this.path, pushOpts, progressCallback);
    } catch (error) {
      throw new PushError(
        `Failed to push: ${error instanceof Error ? error.message : String(error)}`,
        error
      );
    }
  }
}

/**
 * Base Git error class
 */
export class GitError extends Error {
  constructor(
    message: string,
    public readonly cause?: unknown
  ) {
    super(message);
    this.name = 'GitError';
  }
}

/**
 * Clone-specific error
 */
export class CloneError extends GitError {
  constructor(
    message: string,
    public readonly url: string,
    cause?: unknown
  ) {
    super(message, cause);
    this.name = 'CloneError';
  }
}

/**
 * Push-specific error
 */
export class PushError extends GitError {
  constructor(
    message: string,
    cause?: unknown
  ) {
    super(message, cause);
    this.name = 'PushError';
  }
}

/**
 * WASM loader (placeholder - to be implemented)
 */
let wasmInstance: any = null;

async function loadWASM(): Promise<any> {
  if (wasmInstance) {
    return wasmInstance;
  }

  // TODO: Implement actual WASM loading
  // For now, return a mock instance
  wasmInstance = {
    clone: async (_url: string, _path: string, _opts: any, _progress?: ProgressCallback) => {
      // This will be implemented when we integrate with actual WASM
      throw new Error('WASM not yet integrated');
    },
    init: async (_path: string, _opts: any) => {
      throw new Error('WASM not yet integrated');
    },
    open: async (_path: string) => {
      throw new Error('WASM not yet integrated');
    },
    getCurrentBranch: async (_path: string) => {
      throw new Error('WASM not yet integrated');
    },
    listBranches: async (_path: string) => {
      throw new Error('WASM not yet integrated');
    },
    createBranch: async (_path: string, _name: string, _startPoint?: string) => {
      throw new Error('WASM not yet integrated');
    },
    checkout: async (_path: string, _target: string) => {
      throw new Error('WASM not yet integrated');
    },
    fetch: async (_path: string, _opts: any, _progress?: ProgressCallback) => {
      throw new Error('WASM not yet integrated');
    },
    pull: async (_path: string, _opts: any, _progress?: ProgressCallback) => {
      throw new Error('WASM not yet integrated');
    },
    push: async (_path: string, _opts: any, _progress?: ProgressCallback) => {
      throw new Error('WASM not yet integrated');
    },
  };

  return wasmInstance;
}
