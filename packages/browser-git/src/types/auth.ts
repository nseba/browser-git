/**
 * Authentication types for browser-git
 */

/**
 * Authentication method types
 */
export enum AuthMethod {
  /** No authentication */
  None = 'none',
  /** HTTP Basic Authentication (username/password) */
  Basic = 'basic',
  /** Token-based authentication (e.g., Personal Access Token) */
  Token = 'token',
  /** OAuth 2.0 authentication */
  OAuth = 'oauth',
  /** SSH key-based authentication (not supported in browser) */
  SSH = 'ssh',
  /** Custom authentication handler */
  Custom = 'custom',
}

/**
 * Basic authentication configuration
 */
export interface BasicAuthConfig {
  method: AuthMethod.Basic;
  username: string;
  password: string;
}

/**
 * Token authentication configuration
 */
export interface TokenAuthConfig {
  method: AuthMethod.Token;
  /** Personal Access Token or API token */
  token: string;
}

/**
 * OAuth authentication configuration
 */
export interface OAuthConfig {
  method: AuthMethod.OAuth;
  /** OAuth access token */
  accessToken: string;
  /** OAuth refresh token (optional) */
  refreshToken?: string;
}

/**
 * Custom authentication configuration
 */
export interface CustomAuthConfig {
  method: AuthMethod.Custom;
  /** Custom headers to add to requests */
  headers?: Record<string, string>;
  /** Custom authentication handler */
  handler?: (request: Request) => Promise<void> | void;
}

/**
 * No authentication configuration
 */
export interface NoneAuthConfig {
  method: AuthMethod.None;
}

/**
 * Union type of all authentication configurations
 */
export type AuthConfig =
  | NoneAuthConfig
  | BasicAuthConfig
  | TokenAuthConfig
  | OAuthConfig
  | CustomAuthConfig;

/**
 * Convenience type for creating auth configs without specifying method
 */
export interface AuthOptions {
  /** Username for basic auth */
  username?: string;
  /** Password for basic auth */
  password?: string;
  /** Token for token-based auth */
  token?: string;
  /** OAuth access token */
  accessToken?: string;
  /** OAuth refresh token */
  refreshToken?: string;
  /** Custom headers */
  headers?: Record<string, string>;
  /** Custom handler */
  handler?: (request: Request) => Promise<void> | void;
}

/**
 * Credential storage options
 */
export interface CredentialStorageOptions {
  /** Whether to store credentials */
  store?: boolean;
  /** Storage key (defaults to repository URL) */
  key?: string;
  /** Storage mechanism to use */
  storage?: 'memory' | 'session' | 'local' | 'credential-manager';
}

/**
 * Stored credential data
 */
export interface StoredCredentials {
  method: AuthMethod;
  username?: string;
  token?: string;
  accessToken?: string;
  refreshToken?: string;
  /** Timestamp when credentials were stored */
  storedAt: number;
  /** Optional expiration timestamp */
  expiresAt?: number;
}

/**
 * Authentication error types
 */
export class AuthenticationError extends Error {
  constructor(
    message: string,
    public readonly statusCode?: number,
    public readonly hint?: string
  ) {
    super(message);
    this.name = 'AuthenticationError';
  }
}

/**
 * Helper function to create auth config from options
 */
export function createAuthConfig(options: AuthOptions): AuthConfig {
  // Determine the method based on provided options
  if (options.token) {
    return {
      method: AuthMethod.Token,
      token: options.token,
    };
  }

  if (options.accessToken) {
    const config: AuthConfig = {
      method: AuthMethod.OAuth,
      accessToken: options.accessToken,
    };
    if (options.refreshToken) {
      (config as any).refreshToken = options.refreshToken;
    }
    return config;
  }

  if (options.username && options.password) {
    return {
      method: AuthMethod.Basic,
      username: options.username,
      password: options.password,
    };
  }

  if (options.headers || options.handler) {
    const config: AuthConfig = {
      method: AuthMethod.Custom,
    };
    if (options.headers) {
      (config as any).headers = options.headers;
    }
    if (options.handler) {
      (config as any).handler = options.handler;
    }
    return config;
  }

  return {
    method: AuthMethod.None,
  };
}

/**
 * Validates that an auth config has the required fields
 */
export function validateAuthConfig(config: AuthConfig): void {
  switch (config.method) {
    case AuthMethod.Basic:
      if (!config.username || !config.password) {
        throw new AuthenticationError(
          'Basic authentication requires username and password'
        );
      }
      break;

    case AuthMethod.Token:
      if (!config.token) {
        throw new AuthenticationError('Token authentication requires a token');
      }
      break;

    case AuthMethod.OAuth:
      if (!config.accessToken) {
        throw new AuthenticationError(
          'OAuth authentication requires an access token'
        );
      }
      break;

    case AuthMethod.Custom:
      if (!config.headers && !config.handler) {
        throw new AuthenticationError(
          'Custom authentication requires headers or handler'
        );
      }
      break;

    case AuthMethod.None:
      // No validation needed
      break;

    default:
      throw new AuthenticationError(
        `Unsupported authentication method: ${(config as any).method}`
      );
  }
}

/**
 * Applies authentication to a Fetch API Request
 */
export async function applyAuthToRequest(
  request: Request,
  config: AuthConfig
): Promise<Request> {
  validateAuthConfig(config);

  // Create a new request with modified headers
  const headers = new Headers(request.headers);

  switch (config.method) {
    case AuthMethod.Basic: {
      const credentials = btoa(`${config.username}:${config.password}`);
      headers.set('Authorization', `Basic ${credentials}`);
      break;
    }

    case AuthMethod.Token:
      headers.set('Authorization', `Bearer ${config.token}`);
      break;

    case AuthMethod.OAuth:
      headers.set('Authorization', `Bearer ${config.accessToken}`);
      break;

    case AuthMethod.Custom:
      if (config.headers) {
        Object.entries(config.headers).forEach(([key, value]) => {
          headers.set(key, value);
        });
      }
      break;

    case AuthMethod.None:
      // No authentication to apply
      break;
  }

  // Create new request with updated headers
  const authenticatedRequest = new Request(request, { headers });

  // Call custom handler if provided
  if (config.method === AuthMethod.Custom && config.handler) {
    await config.handler(authenticatedRequest);
  }

  return authenticatedRequest;
}
