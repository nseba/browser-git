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
    clone: async (url: string, path: string, opts: any, progress?: ProgressCallback) => {
      // This will be implemented when we integrate with actual WASM
      throw new Error('WASM not yet integrated');
    },
    init: async (path: string, opts: any) => {
      throw new Error('WASM not yet integrated');
    },
    open: async (path: string) => {
      throw new Error('WASM not yet integrated');
    },
    getCurrentBranch: async (path: string) => {
      throw new Error('WASM not yet integrated');
    },
    listBranches: async (path: string) => {
      throw new Error('WASM not yet integrated');
    },
    createBranch: async (path: string, name: string, startPoint?: string) => {
      throw new Error('WASM not yet integrated');
    },
    checkout: async (path: string, target: string) => {
      throw new Error('WASM not yet integrated');
    },
  };

  return wasmInstance;
}
