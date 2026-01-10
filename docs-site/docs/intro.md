---
sidebar_position: 1
slug: /
---

# Introduction

**BrowserGit** is a full-featured Git implementation designed to run entirely in web browsers, built with Go + WebAssembly and TypeScript.

## Why BrowserGit?

Modern web applications increasingly need version control capabilities directly in the browser. Whether you're building:

- **Browser-based IDEs** like CodeSandbox, StackBlitz, or Gitpod
- **Collaborative documentation platforms** with offline support
- **Educational tools** teaching Git without server dependencies
- **Offline-first applications** that sync when connected

BrowserGit provides a complete Git implementation that runs entirely client-side, with no server required for local operations.

## Features

- **Complete Git Operations**: init, add, commit, branch, merge, checkout, clone, fetch, push, pull, log, diff, blame, status
- **Multiple Storage Backends**: IndexedDB (primary), OPFS, LocalStorage (fallback), in-memory
- **Dual API Layers**: Low-level Node.js-like filesystem API + high-level Git operations API
- **Flexible Authentication**: HTTP Basic Auth, OAuth hooks, token-based authentication
- **Future-Proof**: SHA-1 and SHA-256 hash algorithm support
- **Pluggable Architecture**: Modular diff engine, storage adapters, and authentication handlers
- **Cross-Browser**: Chrome, Firefox, Safari (latest 2 versions)
- **High Performance**: Optimized for browser environments with WASM
- **Well Tested**: >80% code coverage with unit, integration, and cross-browser tests

## Architecture Overview

BrowserGit is built as a monorepo with several packages:

```
browser-git/
├── packages/
│   ├── git-core/          # Go + WASM Git implementation
│   ├── browser-git/       # TypeScript wrapper & API layer
│   ├── storage-adapters/  # Storage backend implementations
│   ├── diff-engine/       # Pluggable diff algorithm
│   └── git-cli/           # CLI tool for testing
```

### How It Works

1. **Git Core (Go + WASM)**: The core Git algorithms (object hashing, pack file parsing, merge algorithms) are implemented in Go and compiled to WebAssembly for performance.

2. **TypeScript API**: A high-level TypeScript API wraps the WASM module, providing an ergonomic interface similar to popular Git libraries.

3. **Storage Adapters**: Pluggable storage backends allow repositories to be stored in IndexedDB, OPFS, LocalStorage, or memory.

4. **Diff Engine**: A pluggable diff system using the Myers algorithm for computing file differences.

## Quick Example

```typescript
import { Repository } from "@browser-git/browser-git";

// Initialize a new repository
const repo = await Repository.init("/my-project", {
  storage: "indexeddb",
});

// Create and add a file
await repo.fs.writeFile("/my-project/README.md", "# My Project");
await repo.add(["README.md"]);

// Commit the changes
await repo.commit("Initial commit", {
  author: { name: "Your Name", email: "you@example.com" },
});

// Check the log
const commits = await repo.log();
console.log(commits);
```

## Browser Support

| Browser       | Version | Status                      |
| ------------- | ------- | --------------------------- |
| Chrome        | 86+     | Full support                |
| Firefox       | 111+    | Full support                |
| Safari        | 15.2+   | Full support (OPFS limited) |
| Edge          | 86+     | Full support                |
| Mobile Chrome | Latest  | Full support                |
| Mobile Safari | 15.2+   | Full support (OPFS limited) |

## Next Steps

- [Getting Started](./getting-started) - Install and create your first repository
- [API Reference](./api/repository) - Detailed API documentation
- [Architecture](./architecture/overview) - Deep dive into how BrowserGit works
- [Integration Guide](./guides/integration) - Integrate BrowserGit into your application
