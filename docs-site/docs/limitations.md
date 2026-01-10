---
sidebar_position: 6
---

# Limitations

BrowserGit aims to provide a comprehensive Git implementation for browsers, but there are inherent limitations due to the browser environment. This page documents known limitations and their workarounds.

## Network Limitations

### CORS Restrictions

**Limitation:** Browsers enforce Cross-Origin Resource Sharing (CORS) policies, preventing direct access to most Git servers.

**Impact:**

- Cannot directly clone from GitHub, GitLab, Bitbucket without a proxy
- Cannot push/pull to remote repositories without CORS headers

**Workarounds:**

1. Use a CORS proxy (see [CORS Workarounds Guide](./guides/cors-workarounds))
2. Use server-side relay endpoints
3. Self-host Git servers with proper CORS headers

```typescript
// Using a CORS proxy
const repo = await Repository.clone(url, "/local", {
  corsProxy: "https://cors.isomorphic-git.org",
});
```

### SSH Protocol Not Supported

**Limitation:** Browsers cannot open raw TCP sockets, so SSH protocol is not available.

**Impact:**

- Only HTTPS URLs supported for remotes
- SSH key authentication not possible

**Workaround:** Use HTTPS with personal access tokens:

```typescript
await repo.push("origin", "main", {
  auth: {
    type: "token",
    token: "ghp_xxxxxxxxxxxx",
  },
});
```

### No Background Sync

**Limitation:** Browsers don't allow persistent background processes.

**Impact:**

- No automatic fetch/sync when tab is inactive
- Long operations interrupted if tab is closed

**Workaround:** Use Service Workers for limited background capability:

```typescript
// In service worker
self.addEventListener("sync", (event) => {
  if (event.tag === "git-sync") {
    event.waitUntil(syncRepository());
  }
});
```

## Storage Limitations

### Storage Quotas

**Limitation:** Browsers impose storage limits that vary by browser and device.

| Browser | Typical Quota | Notes                         |
| ------- | ------------- | ----------------------------- |
| Chrome  | 60% of disk   | Can be evicted under pressure |
| Firefox | 50% of disk   | Prompts at 2GB                |
| Safari  | 1GB           | Strict, no expansion          |
| Mobile  | 50-500MB      | Device dependent              |

**Impact:**

- Cannot store very large repositories
- Data may be evicted if storage pressure is high

**Workarounds:**

```typescript
// Request persistent storage
if (navigator.storage?.persist) {
  await navigator.storage.persist();
}

// Check available space before clone
const estimate = await navigator.storage.estimate();
const available = estimate.quota - estimate.usage;

if (available < requiredSize) {
  throw new Error("Insufficient storage");
}
```

### No True Filesystem

**Limitation:** Browser storage is not a real filesystem.

**Impact:**

- No hard links support
- No file permissions (mode bits are simulated)
- No symbolic link guarantees across browsers
- Path length limits may differ

**Notes:**

- File modes are stored but not enforced
- Symlinks work within BrowserGit but not with external tools

### Private Browsing Restrictions

**Limitation:** Private/Incognito mode has severe storage restrictions.

| Browser | Private Mode Behavior         |
| ------- | ----------------------------- |
| Safari  | ~50MB quota, cleared on close |
| Chrome  | Full quota, cleared on close  |
| Firefox | Full quota, cleared on close  |

**Workaround:** Detect private browsing and warn users:

```typescript
async function detectPrivateBrowsing(): Promise<boolean> {
  try {
    const db = await indexedDB.open("test");
    db.close();
    await indexedDB.deleteDatabase("test");
    return false;
  } catch {
    return true;
  }
}
```

## Performance Limitations

### Main Thread Blocking

**Limitation:** Heavy Git operations can block the UI thread.

**Impact:**

- UI freezes during large commits
- Long clone operations may trigger browser warnings

**Workarounds:**

```typescript
// Use Web Workers for heavy operations
const worker = new Worker("git-worker.js");
worker.postMessage({ action: "clone", url, path });

// Or use chunked processing
for await (const chunk of repo.streamingClone(url)) {
  await new Promise((r) => setTimeout(r, 0)); // Yield to UI
  processChunk(chunk);
}
```

### Memory Constraints

**Limitation:** Browsers have limited memory, especially on mobile.

**Impact:**

- Very large files may fail to process
- Many files in a single commit can cause issues
- Large packfiles may not fit in memory

**Recommendations:**

- Use shallow clones for large repositories
- Avoid committing files larger than 100MB
- Use streaming APIs where available

```typescript
// Shallow clone to reduce memory usage
const repo = await Repository.clone(url, "/project", {
  depth: 1,
});
```

### WASM Startup Time

**Limitation:** WebAssembly modules take time to initialize.

**Impact:**

- First Git operation has additional latency (100-500ms)
- Cold starts are slower than subsequent operations

**Workaround:** Pre-initialize during app startup:

```typescript
// Warm up WASM during app initialization
import { initializeWasm } from "@browser-git/browser-git";

// Call early in app lifecycle
await initializeWasm();
```

## Git Feature Limitations

### Partial Clone Not Supported

**Limitation:** Git partial clone (sparse checkout) is not fully supported.

**Impact:**

- Must clone entire repository history
- Cannot fetch only specific directories

**Workaround:** Use shallow clones:

```typescript
const repo = await Repository.clone(url, "/project", {
  depth: 1,
  branch: "main",
});
```

### Submodules Limited Support

**Limitation:** Git submodules have limited support.

**Current Status:**

- ✅ Clone repositories with submodules
- ⚠️ Submodule initialization requires manual steps
- ❌ Recursive submodule operations not supported

### LFS Not Supported

**Limitation:** Git Large File Storage (LFS) is not supported.

**Impact:**

- LFS pointer files are checked out as-is
- Large binary files tracked by LFS won't be available

**Workaround:** Store large files externally and reference them.

### Hooks Not Supported

**Limitation:** Git hooks (pre-commit, post-commit, etc.) are not executed.

**Workaround:** Implement hook-like behavior in application code:

```typescript
async function commitWithHooks(repo, message, options) {
  // Pre-commit hook equivalent
  await runLinting();
  await runTests();

  const commit = await repo.commit(message, options);

  // Post-commit hook equivalent
  await notifyCI(commit);

  return commit;
}
```

### GPG Signing Not Supported

**Limitation:** Commit signing with GPG keys is not supported.

**Impact:**

- Commits cannot be cryptographically signed
- Verified badges won't appear on GitHub

## Browser-Specific Limitations

### Safari

- OPFS may throw unexpected errors - use IndexedDB instead
- Stricter storage quotas
- ITP may affect cross-origin storage

```typescript
// Safari-safe configuration
const repo = await Repository.init("/project", {
  storage: "indexeddb", // Avoid OPFS on Safari
});
```

### Firefox

- `performance.memory` API not available
- Some older versions have ArrayBuffer size limits

### Mobile Browsers

- Aggressive tab killing interrupts long operations
- Significantly lower storage quotas
- Background operations very limited

## Concurrency Limitations

### No Multi-Tab Sync

**Limitation:** Operations in one tab don't automatically reflect in others.

**Impact:**

- Changes made in one tab may not appear in another
- Concurrent writes from multiple tabs can cause conflicts

**Workaround:** Use BroadcastChannel for coordination:

```typescript
const channel = new BroadcastChannel("git-sync");

channel.onmessage = (event) => {
  if (event.data.type === "commit") {
    // Reload repository state
    await repo.reload();
  }
};

// After commit
channel.postMessage({ type: "commit", hash: commit.hash });
```

### Single Writer Constraint

**Limitation:** Only one tab should write to a repository at a time.

**Workaround:** Use Web Locks API:

```typescript
await navigator.locks.request("repo-/project", async () => {
  await repo.commit("message", options);
});
```

## Comparison with Native Git

| Feature              | Native Git   | BrowserGit           |
| -------------------- | ------------ | -------------------- |
| SSH protocol         | ✅           | ❌                   |
| HTTPS protocol       | ✅           | ✅ (with CORS proxy) |
| Full repository size | Unlimited    | Limited by quota     |
| Background sync      | ✅           | ❌                   |
| Submodules           | ✅           | ⚠️ Limited           |
| LFS                  | ✅           | ❌                   |
| Hooks                | ✅           | ❌                   |
| GPG signing          | ✅           | ❌                   |
| Performance          | Native speed | WASM overhead        |
| Concurrent access    | ✅           | ⚠️ Limited           |

## Reporting Issues

If you encounter a limitation not documented here, please:

1. Check the [GitHub Issues](https://github.com/user/browser-git/issues)
2. Search for existing reports
3. Create a new issue with:
   - Browser and version
   - Steps to reproduce
   - Expected vs actual behavior

## See Also

- [Browser Compatibility](./browser-compatibility) - Browser support details
- [Architecture Overview](./architecture/overview) - Design decisions
- [CORS Workarounds](./guides/cors-workarounds) - Network solutions
