# Authentication Guide

BrowserGit supports multiple authentication methods for accessing private repositories and performing authenticated Git operations.

## Table of Contents

- [Supported Authentication Methods](#supported-authentication-methods)
- [GitHub Authentication](#github-authentication)
- [GitLab Authentication](#gitlab-authentication)
- [Bitbucket Authentication](#bitbucket-authentication)
- [Basic Authentication](#basic-authentication)
- [Token Authentication](#token-authentication)
- [OAuth Authentication](#oauth-authentication)
- [Custom Authentication](#custom-authentication)
- [Credential Storage](#credential-storage)
- [Security Considerations](#security-considerations)

## Supported Authentication Methods

BrowserGit supports the following authentication methods:

| Method   | Description            | Use Case                     |
| -------- | ---------------------- | ---------------------------- |
| `none`   | No authentication      | Public repositories          |
| `basic`  | HTTP Basic Auth        | Username and password/token  |
| `token`  | Bearer token           | Personal Access Tokens (PAT) |
| `oauth`  | OAuth 2.0              | OAuth applications           |
| `custom` | Custom headers/handler | Advanced use cases           |

**Note:** SSH authentication is not supported in browsers as the Git HTTP smart protocol does not use SSH.

## GitHub Authentication

### Using Personal Access Token (Recommended)

GitHub Personal Access Tokens (PAT) are the recommended way to authenticate with GitHub repositories.

**Step 1: Generate a Personal Access Token**

1. Go to GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)
2. Click "Generate new token"
3. Select scopes: `repo` (for private repositories) or `public_repo` (for public repositories)
4. Generate and copy the token

**Step 2: Configure BrowserGit**

```typescript
import { Repository, AuthMethod } from "browser-git";

const repo = await Repository.clone("https://github.com/user/repo.git", {
  auth: {
    method: AuthMethod.Token,
    token: "ghp_your_token_here",
  },
});
```

Or using the auth manager:

```typescript
import { createAuthManager, AuthMethod } from "browser-git";

const authManager = createAuthManager("https://github.com/user/repo.git");
await authManager.setAuth({
  method: AuthMethod.Token,
  token: "ghp_your_token_here",
});
```

### Using OAuth

For web applications that need to authenticate users via GitHub OAuth:

```typescript
// After completing OAuth flow and obtaining access token
const repo = await Repository.clone("https://github.com/user/repo.git", {
  auth: {
    method: AuthMethod.OAuth,
    accessToken: oauthAccessToken,
    refreshToken: oauthRefreshToken, // optional
  },
});
```

## GitLab Authentication

### Using Personal Access Token

**Step 1: Generate a Personal Access Token**

1. Go to GitLab User Settings > Access Tokens
2. Create a new token with `read_repository` or `write_repository` scope
3. Copy the token

**Step 2: Configure BrowserGit**

```typescript
const repo = await Repository.clone("https://gitlab.com/user/repo.git", {
  auth: {
    method: AuthMethod.Token,
    token: "glpat-your_token_here",
  },
});
```

### Using OAuth

```typescript
// After completing GitLab OAuth flow
const repo = await Repository.clone("https://gitlab.com/user/repo.git", {
  auth: {
    method: AuthMethod.OAuth,
    accessToken: oauthAccessToken,
  },
});
```

## Bitbucket Authentication

### Using App Password

**Step 1: Create an App Password**

1. Go to Bitbucket Personal settings > App passwords
2. Create a new app password with appropriate permissions
3. Copy the password

**Step 2: Configure BrowserGit**

```typescript
const repo = await Repository.clone("https://bitbucket.org/user/repo.git", {
  auth: {
    method: AuthMethod.Basic,
    username: "your_username",
    password: "app_password",
  },
});
```

## Basic Authentication

Basic authentication uses a username and password (or app password/token).

```typescript
import { AuthMethod } from "browser-git";

const authConfig = {
  method: AuthMethod.Basic,
  username: "myusername",
  password: "mypassword", // or token
};

// With repository
const repo = await Repository.clone(url, { auth: authConfig });

// With auth manager
const authManager = createAuthManager(url);
await authManager.setAuth(authConfig);
```

**Security Warning:** Storing passwords in browser storage is not secure. Use tokens or OAuth when possible.

## Token Authentication

Token authentication is the recommended method for most use cases.

```typescript
import { AuthMethod } from "browser-git";

// Simple token auth
const repo = await Repository.clone(url, {
  auth: {
    method: AuthMethod.Token,
    token: "your_token_here",
  },
});

// With credential storage
const authManager = createAuthManager(url, {
  storageOptions: {
    store: true,
    storage: "session", // 'memory', 'session', 'local', or 'credential-manager'
  },
});

await authManager.setAuth({
  method: AuthMethod.Token,
  token: "your_token_here",
});
```

## OAuth Authentication

OAuth 2.0 authentication for web applications:

```typescript
import { AuthMethod } from "browser-git";

// After completing OAuth flow
const authConfig = {
  method: AuthMethod.OAuth,
  accessToken: "oauth_access_token",
  refreshToken: "oauth_refresh_token", // optional
};

const repo = await Repository.clone(url, { auth: authConfig });

// Check if token needs refresh (if refresh token is available)
const authManager = repo.getAuthManager();
if (authManager.getAuth()?.method === AuthMethod.OAuth) {
  // Implement token refresh logic
}
```

## Custom Authentication

For advanced use cases or custom authentication schemes:

```typescript
import { AuthMethod } from "browser-git";

const authConfig = {
  method: AuthMethod.Custom,
  headers: {
    "X-API-Key": "your_api_key",
    "X-Custom-Auth": "custom_value",
  },
  handler: async (request: Request) => {
    // Custom authentication logic
    const signature = await generateSignature(request);
    request.headers.set("X-Signature", signature);
  },
};

const repo = await Repository.clone(url, { auth: authConfig });
```

## Credential Storage

BrowserGit supports multiple storage mechanisms for credentials:

### Storage Types

```typescript
import { createCredentialStorage } from "browser-git";

// In-memory storage (lost on page reload)
const memoryStorage = createCredentialStorage("memory");

// Session storage (lost when tab closes)
const sessionStorage = createCredentialStorage("session");

// Local storage (persists across sessions)
const localStorage = createCredentialStorage("local");

// Credential Management API (browser's native credential manager)
const credentialManager = createCredentialStorage("credential-manager");
```

### Using Credential Storage

```typescript
import { createAuthManager } from "browser-git";

// Create auth manager with storage
const authManager = createAuthManager("https://github.com/user/repo.git", {
  storageOptions: {
    store: true,
    storage: "session",
    key: "github-repo-auth", // optional custom key
  },
});

// Set auth (will be stored)
await authManager.setAuth({
  method: AuthMethod.Token,
  token: "your_token",
});

// Later, load stored credentials
const loaded = await authManager.loadStoredCredentials();
if (loaded) {
  console.log("Credentials loaded from storage");
}

// Clear stored credentials
await authManager.clearAuth();
```

### Storage Security

⚠️ **Important Security Considerations:**

- **Memory storage**: Most secure but lost on page reload
- **Session storage**: Cleared when tab closes, moderately secure
- **Local storage**: Persists across sessions, **NOT ENCRYPTED**
- **Credential Manager API**: Uses browser's credential manager, limited support

**Recommendations:**

1. Use memory or session storage for tokens
2. Never store sensitive data in local storage without encryption
3. Use HTTPS to prevent token interception
4. Implement token expiration and rotation
5. Consider using short-lived tokens with refresh tokens

## Security Considerations

### Best Practices

1. **Use Tokens, Not Passwords**
   - Prefer Personal Access Tokens over passwords
   - Tokens can be revoked without changing passwords
   - Tokens can have limited scopes

2. **Token Scope**
   - Use the minimum necessary scopes
   - For read-only operations, use read-only tokens

3. **Token Storage**
   - Avoid storing tokens in local storage
   - Use memory or session storage for temporary storage
   - Consider implementing token encryption if persisting

4. **HTTPS Only**
   - Always use HTTPS for repository URLs
   - HTTP connections expose credentials

5. **Token Expiration**
   - Implement token refresh mechanisms
   - Handle authentication errors gracefully

6. **Error Handling**
   - Never log tokens or credentials
   - Provide user-friendly error messages
   - Handle 401 (Unauthorized) and 403 (Forbidden) responses

### Example: Secure Authentication Flow

```typescript
import {
  createAuthManager,
  AuthMethod,
  AuthenticationError,
} from "browser-git";

class SecureAuthFlow {
  private authManager: ReturnType<typeof createAuthManager>;

  constructor(repoUrl: string) {
    this.authManager = createAuthManager(repoUrl, {
      storageOptions: {
        store: true,
        storage: "session", // Use session storage
      },
    });
  }

  async authenticate(token: string): Promise<void> {
    try {
      await this.authManager.setAuth({
        method: AuthMethod.Token,
        token,
      });
    } catch (error) {
      if (error instanceof AuthenticationError) {
        // Handle auth error
        console.error("Authentication failed:", error.message);
        throw error;
      }
      throw error;
    }
  }

  async loadSavedAuth(): Promise<boolean> {
    return await this.authManager.loadStoredCredentials();
  }

  async clearAuth(): Promise<void> {
    await this.authManager.clearAuth();
  }

  getAuthManager() {
    return this.authManager;
  }
}

// Usage
const authFlow = new SecureAuthFlow("https://github.com/user/repo.git");

// Try loading saved credentials
const loaded = await authFlow.loadSavedAuth();

if (!loaded) {
  // Prompt user for token
  const token = await promptUserForToken();
  await authFlow.authenticate(token);
}

// Use auth manager with repository operations
const repo = await Repository.clone(url, {
  authManager: authFlow.getAuthManager(),
});
```

## Handling Authentication Errors

BrowserGit provides detailed error messages for authentication failures:

```typescript
import { Repository, AuthenticationError } from "browser-git";

try {
  const repo = await Repository.clone(url, { auth });
} catch (error) {
  if (error instanceof AuthenticationError) {
    // Authentication failed
    console.error("Auth error:", error.message);
    console.error("Status code:", error.statusCode);
    console.error("Hint:", error.hint);

    if (error.statusCode === 401) {
      // Invalid credentials
      // Prompt user to re-enter credentials
    } else if (error.statusCode === 403) {
      // Access denied
      // User may not have permission to access this repository
    }
  }
}
```

## CORS Considerations

When using authentication with cross-origin requests, you may encounter CORS errors. See the [CORS Guide](./cors-workarounds.md) for solutions.

## Examples

### Complete GitHub Clone Example

```typescript
import { Repository, AuthMethod } from "browser-git";

async function cloneGitHubRepo(owner: string, repo: string, token: string) {
  const url = `https://github.com/${owner}/${repo}.git`;

  try {
    const repository = await Repository.clone(url, {
      path: `/repos/${owner}/${repo}`,
      auth: {
        method: AuthMethod.Token,
        token,
      },
    });

    console.log("Repository cloned successfully!");
    return repository;
  } catch (error) {
    console.error("Clone failed:", error);
    throw error;
  }
}

// Usage
const repo = await cloneGitHubRepo("facebook", "react", "ghp_your_token");
```

### Repository with Persistent Auth

```typescript
import { Repository, createAuthManager, AuthMethod } from "browser-git";

async function setupRepositoryWithAuth(url: string) {
  // Create auth manager with session storage
  const authManager = createAuthManager(url, {
    storageOptions: {
      store: true,
      storage: "session",
    },
  });

  // Try loading stored credentials
  const hasStored = await authManager.loadStoredCredentials();

  if (!hasStored) {
    // Prompt user for token
    const token = prompt("Enter GitHub Personal Access Token:");
    if (!token) throw new Error("Token required");

    await authManager.setAuth({
      method: AuthMethod.Token,
      token,
    });
  }

  // Open or clone repository with auth
  const repo = await Repository.init(url, {
    authManager,
  });

  return { repo, authManager };
}
```

## Resources

- [GitHub Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [GitLab Personal Access Tokens](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)
- [Bitbucket App Passwords](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/)
- [OAuth 2.0 Specification](https://oauth.net/2/)
- [Web Crypto API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Crypto_API)
- [Credential Management API](https://developer.mozilla.org/en-US/docs/Web/API/Credential_Management_API)
