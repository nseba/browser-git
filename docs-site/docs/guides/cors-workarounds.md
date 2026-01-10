---
sidebar_position: 2
---

# CORS Workarounds

Cross-Origin Resource Sharing (CORS) is a common challenge when working with Git in the browser. This guide covers strategies for handling CORS issues when communicating with remote Git servers.

## Understanding the Problem

Browsers enforce the Same-Origin Policy, which prevents JavaScript from making requests to different domains. When BrowserGit tries to communicate with a remote Git server (like GitHub), the browser may block the request unless the server includes proper CORS headers.

```
Browser                           Git Server (github.com)
   │                                       │
   │──── fetch(git-upload-pack) ──────────►│
   │                                       │
   │◄─────── CORS Error ──────────────────│
   │     (No Access-Control-Allow-Origin)  │
```

## Solutions

### 1. CORS Proxy

The most common solution is to route requests through a CORS proxy that adds the appropriate headers.

#### Using a Public Proxy

```typescript
import { Repository } from '@browser-git/browser-git';

const repo = await Repository.clone(
  'https://github.com/user/repo.git',
  '/local',
  {
    corsProxy: 'https://cors-anywhere.herokuapp.com'
  }
);
```

**Note**: Public proxies have rate limits and may not be suitable for production use.

#### Self-Hosted Proxy

Deploy your own CORS proxy for production:

```javascript
// cors-proxy.js (Node.js/Express)
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();

// CORS headers
app.use((req, res, next) => {
  res.header('Access-Control-Allow-Origin', 'https://your-app.com');
  res.header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');

  if (req.method === 'OPTIONS') {
    return res.sendStatus(200);
  }
  next();
});

// Proxy to GitHub
app.use('/github', createProxyMiddleware({
  target: 'https://github.com',
  changeOrigin: true,
  pathRewrite: { '^/github': '' }
}));

// Proxy to GitLab
app.use('/gitlab', createProxyMiddleware({
  target: 'https://gitlab.com',
  changeOrigin: true,
  pathRewrite: { '^/gitlab': '' }
}));

app.listen(3001);
```

Usage:

```typescript
const repo = await Repository.clone(
  'https://github.com/user/repo.git',
  '/local',
  {
    corsProxy: 'https://your-proxy.com/github'
  }
);
```

### 2. Server-Side Git Operations

For better security and reliability, perform Git operations on your server:

```typescript
// Client-side
async function pushChanges() {
  // Get the packfile from BrowserGit
  const packfile = await repo.createPackfile('main');

  // Send to your server
  const response = await fetch('/api/git/push', {
    method: 'POST',
    body: packfile,
    headers: {
      'Content-Type': 'application/x-git-upload-pack-result'
    }
  });
}
```

```typescript
// Server-side (Node.js)
app.post('/api/git/push', async (req, res) => {
  // Receive packfile from browser
  const packfile = req.body;

  // Push to GitHub using server-side Git
  const git = simpleGit('/repo');
  await git.push('origin', 'main');

  res.json({ success: true });
});
```

### 3. GitHub/GitLab API

Use the provider's REST API instead of Git protocol:

```typescript
// Use GitHub's Contents API for simple operations
async function saveFile(path: string, content: string, message: string) {
  const response = await fetch(
    `https://api.github.com/repos/${owner}/${repo}/contents/${path}`,
    {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        message,
        content: btoa(content), // Base64 encode
        sha: existingSha // Required for updates
      })
    }
  );

  return response.json();
}
```

### 4. Cloudflare Workers Proxy

Cloudflare Workers provide a scalable, edge-deployed proxy:

```javascript
// worker.js
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

async function handleRequest(request) {
  const url = new URL(request.url);

  // Extract target from path
  const targetUrl = url.pathname.slice(1); // Remove leading /

  // Validate target
  if (!targetUrl.startsWith('https://github.com/') &&
      !targetUrl.startsWith('https://gitlab.com/')) {
    return new Response('Invalid target', { status: 400 });
  }

  // Forward request
  const response = await fetch(targetUrl, {
    method: request.method,
    headers: request.headers,
    body: request.body
  });

  // Add CORS headers
  const newHeaders = new Headers(response.headers);
  newHeaders.set('Access-Control-Allow-Origin', 'https://your-app.com');
  newHeaders.set('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');

  return new Response(response.body, {
    status: response.status,
    headers: newHeaders
  });
}
```

## Provider-Specific Solutions

### GitHub

GitHub's API supports CORS for many operations:

```typescript
// These work without a proxy
const apiResponse = await fetch('https://api.github.com/repos/user/repo', {
  headers: { 'Authorization': `Bearer ${token}` }
});

// Git protocol requires a proxy
const repo = await Repository.clone('https://github.com/user/repo.git', '/local', {
  corsProxy: 'https://your-proxy.com'
});
```

### GitLab

GitLab has similar restrictions:

```typescript
// GraphQL API works
const graphqlResponse = await fetch('https://gitlab.com/api/graphql', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({ query: '...' })
});

// Git protocol needs proxy
const repo = await Repository.clone('https://gitlab.com/user/repo.git', '/local', {
  corsProxy: 'https://your-proxy.com'
});
```

### Self-Hosted Git Servers

Configure your server to allow CORS:

```nginx
# nginx.conf
location /git/ {
    add_header 'Access-Control-Allow-Origin' 'https://your-app.com';
    add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
    add_header 'Access-Control-Allow-Headers' 'Content-Type, Authorization';

    if ($request_method = 'OPTIONS') {
        return 204;
    }

    proxy_pass http://git-server:9418;
}
```

## Handling CORS Errors

### Detection

```typescript
import { NetworkError } from '@browser-git/browser-git';

try {
  await repo.fetch('origin');
} catch (error) {
  if (error instanceof NetworkError && error.isCorsError) {
    console.log('CORS error detected');
    // Suggest solutions to user
  }
}
```

### User Feedback

```typescript
function handleCorsError(error: NetworkError) {
  if (error.isCorsError) {
    showModal({
      title: 'Connection Blocked',
      message: `
        Your browser blocked the connection to the Git server.

        Options:
        1. Use a different Git hosting provider
        2. Configure a CORS proxy
        3. Use offline mode and sync later
      `,
      actions: [
        { label: 'Configure Proxy', onClick: openProxySettings },
        { label: 'Work Offline', onClick: enableOfflineMode }
      ]
    });
  }
}
```

## Security Considerations

### Proxy Security

1. **Validate Target URLs**: Only allow requests to known Git hosts
2. **Rate Limiting**: Prevent abuse of your proxy
3. **Authentication**: Consider requiring authentication for proxy access
4. **Logging**: Monitor proxy usage for suspicious activity

```javascript
// Secure proxy example
const ALLOWED_HOSTS = [
  'github.com',
  'gitlab.com',
  'bitbucket.org'
];

app.use('/proxy', (req, res, next) => {
  const targetUrl = new URL(req.query.url);

  if (!ALLOWED_HOSTS.includes(targetUrl.hostname)) {
    return res.status(403).json({ error: 'Host not allowed' });
  }

  // Rate limiting
  if (rateLimiter.isLimited(req.ip)) {
    return res.status(429).json({ error: 'Too many requests' });
  }

  next();
});
```

### Token Handling

Never send tokens through untrusted proxies:

```typescript
// Bad: Token exposed to proxy
await fetch(`https://untrusted-proxy.com/https://github.com/...`, {
  headers: { Authorization: `Bearer ${token}` }
});

// Good: Use your own trusted proxy
await fetch(`https://your-proxy.com/github/...`, {
  headers: { Authorization: `Bearer ${token}` }
});

// Better: Server handles authentication
await fetch('/api/git/fetch', {
  headers: { 'X-Session': sessionId }
});
```

## Offline-First Approach

Consider an offline-first architecture to minimize CORS issues:

```typescript
class OfflineFirstRepository {
  private repo: Repository;
  private syncQueue: SyncQueue;

  async commit(message: string, author: Author) {
    // Always works locally
    await this.repo.commit(message, { author });

    // Queue sync for later
    this.syncQueue.enqueue({ type: 'push' });
  }

  async sync() {
    // Only sync when online and possible
    if (!navigator.onLine) {
      return { status: 'offline' };
    }

    try {
      await this.repo.push('origin', 'main');
      return { status: 'synced' };
    } catch (error) {
      if (error.isCorsError) {
        return { status: 'blocked', reason: 'cors' };
      }
      throw error;
    }
  }
}
```

## Next Steps

- [Authentication](./authentication) - Set up authentication for remotes
- [Integration Guide](./integration) - Full integration patterns
- [Browser Compatibility](../browser-compatibility) - Cross-browser support
