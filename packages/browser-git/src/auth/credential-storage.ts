/**
 * Credential storage implementation using various browser storage mechanisms
 */

import type { AuthConfig, StoredCredentials } from '../types/auth';
import { AuthMethod } from '../types/auth';

/**
 * Storage backend interface
 */
interface StorageBackend {
  get(key: string): Promise<string | null>;
  set(key: string, value: string): Promise<void>;
  remove(key: string): Promise<void>;
}

/**
 * In-memory storage backend
 */
class MemoryStorage implements StorageBackend {
  private store = new Map<string, string>();

  async get(key: string): Promise<string | null> {
    return this.store.get(key) ?? null;
  }

  async set(key: string, value: string): Promise<void> {
    this.store.set(key, value);
  }

  async remove(key: string): Promise<void> {
    this.store.delete(key);
  }
}

/**
 * Session storage backend
 */
class SessionStorageBackend implements StorageBackend {
  async get(key: string): Promise<string | null> {
    try {
      return sessionStorage.getItem(key);
    } catch {
      return null;
    }
  }

  async set(key: string, value: string): Promise<void> {
    try {
      sessionStorage.setItem(key, value);
    } catch (e) {
      console.warn('Failed to store credentials in sessionStorage:', e);
    }
  }

  async remove(key: string): Promise<void> {
    try {
      sessionStorage.removeItem(key);
    } catch {
      // Ignore errors
    }
  }
}

/**
 * Local storage backend
 */
class LocalStorageBackend implements StorageBackend {
  async get(key: string): Promise<string | null> {
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  }

  async set(key: string, value: string): Promise<void> {
    try {
      localStorage.setItem(key, value);
    } catch (e) {
      console.warn('Failed to store credentials in localStorage:', e);
    }
  }

  async remove(key: string): Promise<void> {
    try {
      localStorage.removeItem(key);
    } catch {
      // Ignore errors
    }
  }
}

/**
 * Credential Management API backend (for password/token storage)
 * Note: This API is not widely supported and has limitations
 */
class CredentialManagerBackend implements StorageBackend {
  private fallback = new MemoryStorage();

  async get(key: string): Promise<string | null> {
    // Credential Management API doesn't support arbitrary key-value storage
    // Fall back to memory storage
    return this.fallback.get(key);
  }

  async set(key: string, value: string): Promise<void> {
    // Store in memory fallback
    await this.fallback.set(key, value);

    // Try to use Credential Management API for password credentials
    if ('credentials' in navigator && 'PasswordCredential' in window) {
      try {
        const data = JSON.parse(value) as StoredCredentials;
        if (data.username && (data.method === AuthMethod.Basic || data.method === 'basic')) {
          // Note: PasswordCredential doesn't store the password in a retrievable way
          // This is just for browser's autofill
          const credential = new (window as any).PasswordCredential({
            id: data.username,
            password: '', // Can't store actual password
            name: data.username,
          });
          await navigator.credentials.store(credential);
        }
      } catch (e) {
        console.warn('Failed to store credentials using Credential Manager:', e);
      }
    }
  }

  async remove(key: string): Promise<void> {
    await this.fallback.remove(key);
  }
}

/**
 * Credential storage manager
 */
export class CredentialStorage {
  private backend: StorageBackend;
  private prefix = 'browsergit:credentials:';

  constructor(type: 'memory' | 'session' | 'local' | 'credential-manager' = 'memory') {
    switch (type) {
      case 'session':
        this.backend = new SessionStorageBackend();
        break;
      case 'local':
        this.backend = new LocalStorageBackend();
        break;
      case 'credential-manager':
        this.backend = new CredentialManagerBackend();
        break;
      case 'memory':
      default:
        this.backend = new MemoryStorage();
        break;
    }
  }

  /**
   * Stores credentials for a given key (typically repository URL)
   */
  async store(key: string, config: AuthConfig, expiresIn?: number): Promise<void> {
    const credentials: StoredCredentials = {
      method: config.method as AuthMethod,
      storedAt: Date.now(),
    };

    if (expiresIn) {
      credentials.expiresAt = Date.now() + expiresIn;
    }

    // Extract credentials based on auth method
    switch (config.method) {
      case AuthMethod.Basic:
        credentials.username = config.username;
        // Note: Storing passwords in browser storage is not secure
        // This is a convenience feature and should be used with caution
        break;

      case AuthMethod.Token:
        credentials.token = config.token;
        break;

      case AuthMethod.OAuth:
        credentials.accessToken = config.accessToken;
        credentials.refreshToken = config.refreshToken || undefined;
        break;

      default:
        // Don't store custom or none auth
        return;
    }

    const storageKey = this.prefix + this.sanitizeKey(key);
    await this.backend.set(storageKey, JSON.stringify(credentials));
  }

  /**
   * Retrieves stored credentials for a given key
   */
  async retrieve(key: string): Promise<StoredCredentials | null> {
    const storageKey = this.prefix + this.sanitizeKey(key);
    const data = await this.backend.get(storageKey);

    if (!data) {
      return null;
    }

    try {
      const credentials = JSON.parse(data) as StoredCredentials;

      // Check if credentials have expired
      if (credentials.expiresAt && Date.now() > credentials.expiresAt) {
        await this.remove(key);
        return null;
      }

      return credentials;
    } catch {
      // Invalid JSON, remove it
      await this.remove(key);
      return null;
    }
  }

  /**
   * Removes stored credentials for a given key
   */
  async remove(key: string): Promise<void> {
    const storageKey = this.prefix + this.sanitizeKey(key);
    await this.backend.remove(storageKey);
  }

  /**
   * Clears all stored credentials
   */
  async clear(): Promise<void> {
    // This is limited - we can only clear what we know about
    // For a full implementation, we'd need to iterate over all keys
    console.warn('CredentialStorage.clear() has limited functionality');
  }

  /**
   * Converts stored credentials to auth config
   */
  credentialsToAuthConfig(credentials: StoredCredentials): AuthConfig | null {
    switch (credentials.method) {
      case AuthMethod.Basic:
      case 'basic':
        if (!credentials.username) return null;
        return {
          method: AuthMethod.Basic,
          username: credentials.username,
          password: '', // Password is not retrievable
        };

      case AuthMethod.Token:
      case 'token':
        if (!credentials.token) return null;
        return {
          method: AuthMethod.Token,
          token: credentials.token,
        };

      case AuthMethod.OAuth:
      case 'oauth':
        if (!credentials.accessToken) return null;
        return {
          method: AuthMethod.OAuth,
          accessToken: credentials.accessToken,
          refreshToken: credentials.refreshToken,
        };

      default:
        return null;
    }
  }

  /**
   * Sanitizes a key for storage
   */
  private sanitizeKey(key: string): string {
    // Remove special characters and limit length
    return key
      .replace(/[^a-zA-Z0-9-_.]/g, '_')
      .substring(0, 200);
  }
}

/**
 * Default credential storage instance
 */
export const defaultCredentialStorage = new CredentialStorage('memory');

/**
 * Creates a credential storage instance
 */
export function createCredentialStorage(
  type: 'memory' | 'session' | 'local' | 'credential-manager' = 'memory'
): CredentialStorage {
  return new CredentialStorage(type);
}
