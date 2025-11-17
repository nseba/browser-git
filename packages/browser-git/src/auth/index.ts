/**
 * Authentication module for browser-git
 * Provides authentication management and credential storage
 */

export { AuthMethod } from '../types/auth';
export type {
  AuthConfig,
  BasicAuthConfig,
  TokenAuthConfig,
  OAuthConfig,
  CustomAuthConfig,
  NoneAuthConfig,
  AuthOptions,
  CredentialStorageOptions,
  StoredCredentials,
} from '../types/auth';

export {
  AuthenticationError,
  createAuthConfig,
  validateAuthConfig,
  applyAuthToRequest,
} from '../types/auth';

export {
  CredentialStorage,
  defaultCredentialStorage,
  createCredentialStorage,
} from './credential-storage';

import type { AuthConfig, AuthOptions, CredentialStorageOptions } from '../types/auth';
import { createAuthConfig, validateAuthConfig, applyAuthToRequest } from '../types/auth';
import { CredentialStorage, defaultCredentialStorage } from './credential-storage';

/**
 * Authentication manager for a repository
 */
export class AuthManager {
  private config: AuthConfig | null = null;
  private credentialStorage: CredentialStorage;
  private storageOptions: CredentialStorageOptions;
  private repositoryUrl: string;

  constructor(
    repositoryUrl: string,
    storage?: CredentialStorage,
    storageOptions?: CredentialStorageOptions
  ) {
    this.repositoryUrl = repositoryUrl;
    this.credentialStorage = storage || defaultCredentialStorage;
    this.storageOptions = storageOptions || { store: false };
  }

  /**
   * Sets authentication configuration
   */
  async setAuth(config: AuthConfig | AuthOptions): Promise<void> {
    // Convert options to config if needed
    const authConfig = 'method' in config ? config : createAuthConfig(config);

    // Validate the configuration
    validateAuthConfig(authConfig);

    this.config = authConfig;

    // Store credentials if requested
    if (this.storageOptions.store) {
      const key = this.storageOptions.key || this.repositoryUrl;
      await this.credentialStorage.store(key, authConfig);
    }
  }

  /**
   * Gets the current authentication configuration
   */
  getAuth(): AuthConfig | null {
    return this.config;
  }

  /**
   * Clears authentication
   */
  async clearAuth(): Promise<void> {
    this.config = null;

    if (this.storageOptions.store) {
      const key = this.storageOptions.key || this.repositoryUrl;
      await this.credentialStorage.remove(key);
    }
  }

  /**
   * Loads stored credentials
   */
  async loadStoredCredentials(): Promise<boolean> {
    const key = this.storageOptions.key || this.repositoryUrl;
    const stored = await this.credentialStorage.retrieve(key);

    if (!stored) {
      return false;
    }

    const authConfig = this.credentialStorage.credentialsToAuthConfig(stored);
    if (authConfig) {
      this.config = authConfig;
      return true;
    }

    return false;
  }

  /**
   * Applies authentication to a fetch request
   */
  async applyToRequest(request: Request): Promise<Request> {
    if (!this.config) {
      return request;
    }

    return applyAuthToRequest(request, this.config);
  }

  /**
   * Sets storage options
   */
  setStorageOptions(options: CredentialStorageOptions): void {
    this.storageOptions = { ...this.storageOptions, ...options };
  }

  /**
   * Gets storage options
   */
  getStorageOptions(): CredentialStorageOptions {
    return { ...this.storageOptions };
  }

  /**
   * Checks if authentication is configured
   */
  hasAuth(): boolean {
    return this.config !== null && this.config.method !== 'none';
  }

  /**
   * Gets the authentication method
   */
  getAuthMethod(): string | null {
    return this.config?.method || null;
  }
}

/**
 * Creates an authentication manager
 */
export function createAuthManager(
  repositoryUrl: string,
  options?: {
    storage?: CredentialStorage;
    storageOptions?: CredentialStorageOptions;
  }
): AuthManager {
  return new AuthManager(
    repositoryUrl,
    options?.storage,
    options?.storageOptions
  );
}
