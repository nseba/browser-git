---
sidebar_position: 3
---

# Authentication Guide

This guide covers how to authenticate with remote Git repositories when using BrowserGit. We support several authentication methods for different providers and use cases.

## Authentication Methods

### Personal Access Tokens

The most common authentication method for browser environments.

#### GitHub

1. Generate a token at [GitHub Settings > Developer settings > Personal access tokens](https://github.com/settings/tokens)
2. Select scopes: `repo` for full access, or `public_repo` for public repositories only

```typescript
const repo = await Repository.clone(
  'https://github.com/user/private-repo.git',
  '/local',
  {
    auth: {
      type: 'token',
      token: 'ghp_xxxxxxxxxxxxxxxxxxxx'
    }
  }
);

// Push with authentication
await repo.push('origin', 'main', {
  auth: {
    type: 'token',
    token: 'ghp_xxxxxxxxxxxxxxxxxxxx'
  }
});
```

#### GitLab

1. Generate a token at [GitLab Settings > Access Tokens](https://gitlab.com/-/profile/personal_access_tokens)
2. Select scopes: `read_repository`, `write_repository`

```typescript
const repo = await Repository.clone(
  'https://gitlab.com/user/repo.git',
  '/local',
  {
    auth: {
      type: 'token',
      token: 'glpat-xxxxxxxxxxxxxxxxxxxx'
    }
  }
);
```

#### Bitbucket

1. Generate an App Password at [Bitbucket Settings > App passwords](https://bitbucket.org/account/settings/app-passwords/)
2. Select permissions: `repository:read`, `repository:write`

```typescript
const repo = await Repository.clone(
  'https://bitbucket.org/user/repo.git',
  '/local',
  {
    auth: {
      type: 'basic',
      username: 'your-username',
      password: 'your-app-password'
    }
  }
);
```

### OAuth Authentication

For user-facing applications, OAuth provides a better user experience.

#### OAuth Flow

```typescript
class GitHubOAuth {
  private clientId: string;
  private redirectUri: string;

  constructor(clientId: string, redirectUri: string) {
    this.clientId = clientId;
    this.redirectUri = redirectUri;
  }

  // Step 1: Redirect user to GitHub
  authorize() {
    const params = new URLSearchParams({
      client_id: this.clientId,
      redirect_uri: this.redirectUri,
      scope: 'repo',
      state: crypto.randomUUID()
    });

    // Store state for CSRF protection
    sessionStorage.setItem('oauth_state', params.get('state')!);

    window.location.href = `https://github.com/login/oauth/authorize?${params}`;
  }

  // Step 2: Handle callback (requires server-side token exchange)
  async handleCallback(code: string, state: string): Promise<string> {
    // Verify state
    if (state !== sessionStorage.getItem('oauth_state')) {
      throw new Error('Invalid state - possible CSRF attack');
    }

    // Exchange code for token (must be done server-side)
    const response = await fetch('/api/github/token', {
      method: 'POST',
      body: JSON.stringify({ code })
    });

    const { access_token } = await response.json();
    return access_token;
  }
}

// Usage
const oauth = new GitHubOAuth('your-client-id', 'https://your-app.com/callback');

// On "Connect GitHub" button click
oauth.authorize();

// On callback page
const token = await oauth.handleCallback(code, state);
```

#### Server-Side Token Exchange

```typescript
// server.ts (Node.js/Express)
app.post('/api/github/token', async (req, res) => {
  const { code } = req.body;

  const response = await fetch('https://github.com/login/oauth/access_token', {
    method: 'POST',
    headers: {
      'Accept': 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      client_id: process.env.GITHUB_CLIENT_ID,
      client_secret: process.env.GITHUB_CLIENT_SECRET,
      code
    })
  });

  const data = await response.json();

  if (data.error) {
    return res.status(400).json({ error: data.error_description });
  }

  res.json({ access_token: data.access_token });
});
```

### HTTP Basic Authentication

For self-hosted Git servers or services that support it:

```typescript
const repo = await Repository.clone(
  'https://git.example.com/repo.git',
  '/local',
  {
    auth: {
      type: 'basic',
      username: 'username',
      password: 'password'
    }
  }
);
```

## Credential Storage

### Secure Storage Options

#### Session Storage (Recommended for Web Apps)

```typescript
class CredentialManager {
  private static readonly KEY = 'git_credentials';

  static save(credentials: GitCredentials): void {
    // Encrypt before storing
    const encrypted = this.encrypt(JSON.stringify(credentials));
    sessionStorage.setItem(this.KEY, encrypted);
  }

  static load(): GitCredentials | null {
    const encrypted = sessionStorage.getItem(this.KEY);
    if (!encrypted) return null;

    const decrypted = this.decrypt(encrypted);
    return JSON.parse(decrypted);
  }

  static clear(): void {
    sessionStorage.removeItem(this.KEY);
  }

  private static encrypt(data: string): string {
    // Use Web Crypto API for encryption
    // Implementation depends on your security requirements
    return btoa(data); // Simple base64 for demo
  }

  private static decrypt(data: string): string {
    return atob(data);
  }
}
```

#### IndexedDB with Encryption

```typescript
class SecureCredentialStore {
  private db: IDBDatabase | null = null;
  private encryptionKey: CryptoKey | null = null;

  async initialize(userPassword: string): Promise<void> {
    // Derive encryption key from password
    const encoder = new TextEncoder();
    const passwordBuffer = encoder.encode(userPassword);

    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      passwordBuffer,
      'PBKDF2',
      false,
      ['deriveKey']
    );

    this.encryptionKey = await crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: encoder.encode('browser-git-credentials'),
        iterations: 100000,
        hash: 'SHA-256'
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );

    // Open database
    this.db = await this.openDatabase();
  }

  async saveCredentials(remote: string, credentials: GitCredentials): Promise<void> {
    if (!this.db || !this.encryptionKey) {
      throw new Error('Store not initialized');
    }

    const encrypted = await this.encrypt(JSON.stringify(credentials));

    return new Promise((resolve, reject) => {
      const tx = this.db!.transaction(['credentials'], 'readwrite');
      const store = tx.objectStore('credentials');
      store.put({ remote, encrypted });
      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error);
    });
  }

  async getCredentials(remote: string): Promise<GitCredentials | null> {
    if (!this.db || !this.encryptionKey) {
      throw new Error('Store not initialized');
    }

    return new Promise((resolve, reject) => {
      const tx = this.db!.transaction(['credentials'], 'readonly');
      const store = tx.objectStore('credentials');
      const request = store.get(remote);

      request.onsuccess = async () => {
        if (!request.result) {
          resolve(null);
          return;
        }

        const decrypted = await this.decrypt(request.result.encrypted);
        resolve(JSON.parse(decrypted));
      };

      request.onerror = () => reject(request.error);
    });
  }

  private async encrypt(data: string): Promise<ArrayBuffer> {
    const encoder = new TextEncoder();
    const iv = crypto.getRandomValues(new Uint8Array(12));

    const encrypted = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv },
      this.encryptionKey!,
      encoder.encode(data)
    );

    // Prepend IV to encrypted data
    const result = new Uint8Array(iv.length + encrypted.byteLength);
    result.set(iv);
    result.set(new Uint8Array(encrypted), iv.length);

    return result.buffer;
  }

  private async decrypt(data: ArrayBuffer): Promise<string> {
    const dataArray = new Uint8Array(data);
    const iv = dataArray.slice(0, 12);
    const encrypted = dataArray.slice(12);

    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      this.encryptionKey!,
      encrypted
    );

    return new TextDecoder().decode(decrypted);
  }

  private openDatabase(): Promise<IDBDatabase> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('git-credentials', 1);

      request.onupgradeneeded = () => {
        const db = request.result;
        db.createObjectStore('credentials', { keyPath: 'remote' });
      };

      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }
}
```

## Authentication Hooks

BrowserGit supports authentication hooks for dynamic credential handling:

```typescript
const repo = await Repository.clone(
  'https://github.com/user/repo.git',
  '/local',
  {
    auth: {
      type: 'callback',
      callback: async (url: string, operation: string) => {
        // Prompt user for credentials if needed
        const credentials = await promptForCredentials(url);

        return {
          type: 'token',
          token: credentials.token
        };
      }
    }
  }
);
```

### Implementing a Credential Prompt

```typescript
async function promptForCredentials(url: string): Promise<GitCredentials> {
  return new Promise((resolve, reject) => {
    const modal = document.createElement('div');
    modal.innerHTML = `
      <div class="modal">
        <h2>Authentication Required</h2>
        <p>Enter credentials for ${new URL(url).hostname}</p>
        <form id="credential-form">
          <input type="text" id="username" placeholder="Username or Token" />
          <input type="password" id="password" placeholder="Password (if required)" />
          <button type="submit">Connect</button>
          <button type="button" id="cancel">Cancel</button>
        </form>
      </div>
    `;

    document.body.appendChild(modal);

    const form = modal.querySelector('#credential-form')!;
    const cancelBtn = modal.querySelector('#cancel')!;

    form.addEventListener('submit', (e) => {
      e.preventDefault();
      const username = (modal.querySelector('#username') as HTMLInputElement).value;
      const password = (modal.querySelector('#password') as HTMLInputElement).value;

      document.body.removeChild(modal);

      if (password) {
        resolve({ type: 'basic', username, password });
      } else {
        resolve({ type: 'token', token: username });
      }
    });

    cancelBtn.addEventListener('click', () => {
      document.body.removeChild(modal);
      reject(new Error('User cancelled'));
    });
  });
}
```

## Multi-Account Support

Handle multiple accounts for the same provider:

```typescript
interface Account {
  id: string;
  provider: 'github' | 'gitlab' | 'bitbucket';
  username: string;
  token: string;
}

class AccountManager {
  private accounts: Map<string, Account> = new Map();

  addAccount(account: Account): void {
    this.accounts.set(account.id, account);
  }

  getAccountForRemote(remoteUrl: string): Account | undefined {
    const hostname = new URL(remoteUrl).hostname;

    // Find matching account
    for (const account of this.accounts.values()) {
      if (hostname.includes(account.provider)) {
        return account;
      }
    }

    return undefined;
  }

  async createAuthForRemote(remoteUrl: string): Promise<GitAuth | undefined> {
    const account = this.getAccountForRemote(remoteUrl);

    if (!account) {
      return undefined;
    }

    return {
      type: 'token',
      token: account.token
    };
  }
}
```

## Security Best Practices

### 1. Never Store Tokens in Code

```typescript
// BAD
const token = 'ghp_xxxxxxxxxxxx';

// GOOD
const token = await getTokenFromSecureStorage();
```

### 2. Use Short-Lived Tokens

```typescript
// Request tokens with limited lifetime
const response = await fetch('/api/github/token', {
  body: JSON.stringify({
    scope: 'repo',
    expires_in: 3600 // 1 hour
  })
});
```

### 3. Implement Token Refresh

```typescript
class TokenManager {
  private token: string | null = null;
  private expiresAt: number = 0;

  async getToken(): Promise<string> {
    if (this.token && Date.now() < this.expiresAt - 60000) {
      return this.token;
    }

    // Refresh token
    const { token, expires_in } = await this.refreshToken();
    this.token = token;
    this.expiresAt = Date.now() + expires_in * 1000;

    return this.token;
  }

  private async refreshToken(): Promise<{ token: string; expires_in: number }> {
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include'
    });

    return response.json();
  }
}
```

### 4. Clear Credentials on Logout

```typescript
async function logout(): Promise<void> {
  // Clear stored credentials
  CredentialManager.clear();

  // Revoke OAuth tokens if possible
  await revokeToken();

  // Clear repository data if needed
  await clearRepositories();
}
```

## Error Handling

### Authentication Errors

```typescript
import { AuthenticationError } from '@browser-git/browser-git';

try {
  await repo.push('origin', 'main');
} catch (error) {
  if (error instanceof AuthenticationError) {
    switch (error.code) {
      case 'INVALID_CREDENTIALS':
        showError('Invalid credentials. Please check your token.');
        break;
      case 'TOKEN_EXPIRED':
        await refreshAndRetry();
        break;
      case 'INSUFFICIENT_SCOPE':
        showError('Token lacks required permissions. Please grant repo access.');
        break;
      case 'RATE_LIMITED':
        showError(`Rate limited. Try again after ${error.retryAfter} seconds.`);
        break;
    }
  }
}
```

## Next Steps

- [CORS Workarounds](./cors-workarounds) - Handle cross-origin requests
- [Integration Guide](./integration) - Full integration patterns
- [Browser Compatibility](../browser-compatibility) - Cross-browser support
