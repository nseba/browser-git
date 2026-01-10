---
sidebar_position: 1
---

# Architecture Overview

BrowserGit is designed as a layered architecture that separates concerns and allows for maximum flexibility and performance in browser environments.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  (Your Browser Application)                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    @browser-git/browser-git                  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐ │
│  │   Repository    │  │   FileSystem    │  │  Remote Ops │ │
│  │      API        │  │      API        │  │    (HTTP)   │ │
│  └────────┬────────┘  └────────┬────────┘  └──────┬──────┘ │
│           │                    │                   │        │
│           └────────────────────┼───────────────────┘        │
│                                │                            │
│                    ┌───────────▼───────────┐               │
│                    │     WASM Bridge       │               │
│                    └───────────┬───────────┘               │
└────────────────────────────────┼────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────┐
│                       @browser-git/git-core                  │
│                      (Go + WebAssembly)                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐│
│  │  Objects │  │   Refs   │  │  Index   │  │   Protocol   ││
│  │  (blob,  │  │(branches,│  │(staging) │  │  (packfile,  ││
│  │  tree,   │  │  tags)   │  │          │  │   pkt-line)  ││
│  │  commit) │  │          │  │          │  │              ││
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘│
└─────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────┐
│                  @browser-git/storage-adapters               │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐│
│  │IndexedDB │  │   OPFS   │  │  Local   │  │   Memory     ││
│  │ Adapter  │  │ Adapter  │  │ Storage  │  │   Adapter    ││
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘│
└─────────────────────────────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────┐
│                     Browser Storage APIs                     │
│            (IndexedDB, OPFS, LocalStorage)                  │
└─────────────────────────────────────────────────────────────┘
```

## Package Structure

### @browser-git/git-core

The core package implements Git algorithms in Go, compiled to WebAssembly:

- **Objects**: Git object types (blob, tree, commit, tag) with SHA-1/SHA-256 hashing
- **Refs**: Branch and tag reference management
- **Index**: Staging area (index file) implementation
- **Protocol**: Git HTTP smart protocol (pkt-line, packfile parsing, delta resolution)
- **Merge**: Three-way merge algorithm

### @browser-git/browser-git

The main TypeScript package providing:

- **Repository API**: High-level Git operations (init, clone, commit, push, etc.)
- **FileSystem API**: Node.js-like filesystem interface for browser storage
- **WASM Bridge**: JavaScript-WebAssembly communication layer
- **Remote Operations**: HTTP client for Git smart protocol

### @browser-git/storage-adapters

Pluggable storage backends:

- **IndexedDB Adapter**: Primary storage for most browsers
- **OPFS Adapter**: High-performance storage for modern browsers
- **LocalStorage Adapter**: Fallback for limited compatibility
- **Memory Adapter**: Ephemeral storage for testing

### @browser-git/diff-engine

Modular diff computation:

- **Myers Diff**: Default diff algorithm implementation
- **Binary Diff**: Utilities for binary file comparison
- **Pluggable Interface**: Support for custom diff algorithms

## Data Flow

### Local Operations

```
User Action (e.g., commit)
         │
         ▼
    Repository API
         │
         ▼
    WASM Bridge ──────────► Git Core (WASM)
         │                       │
         │                       ▼
         │               Compute objects
         │               (tree, commit)
         │                       │
         ▼                       ▼
    Storage Adapter ◄──── Object data
         │
         ▼
    Browser Storage
    (IndexedDB/OPFS)
```

### Remote Operations

```
User Action (e.g., push)
         │
         ▼
    Repository API
         │
         ├──────────────────────────────┐
         │                              │
         ▼                              ▼
    WASM Bridge               HTTP Client (fetch)
         │                              │
         ▼                              ▼
    Git Core (WASM)              Remote Server
         │                         (GitHub, etc.)
         ▼                              │
    Pack objects ◄──────────────────────┘
    for transfer           (packfile response)
```

## Key Design Decisions

### 1. WASM for Performance-Critical Code

Git operations like SHA-1 hashing, packfile parsing, and delta resolution are computationally intensive. By implementing these in Go and compiling to WASM, we achieve near-native performance while maintaining browser compatibility.

### 2. Pluggable Storage

Different applications have different storage needs:

- Large repositories benefit from OPFS's file-like API
- Smaller repositories work well with IndexedDB
- Testing scenarios use in-memory storage

The storage adapter interface allows swapping backends without changing application code.

### 3. Dual API Layers

Two API levels serve different needs:

**High-level (Repository)**: Familiar Git commands for most use cases

```typescript
await repo.commit("message", { author });
```

**Low-level (FileSystem)**: Direct file manipulation when needed

```typescript
await fs.writeFile(path, content);
await fs.readdir(path);
```

### 4. Progressive Enhancement

BrowserGit detects available browser features and gracefully degrades:

- OPFS available? Use for best performance
- No OPFS? Fall back to IndexedDB
- Limited storage? Use compression
- No WebAssembly? Error with clear message

## Memory Management

### WASM Memory

The WASM module manages its own linear memory. Large operations (like cloning) are chunked to prevent memory exhaustion:

```typescript
// Packfile processing is streamed
await processPackfile(stream, {
  chunkSize: 1024 * 1024, // 1MB chunks
  onProgress: (percent) => console.log(`${percent}%`),
});
```

### Object Caching

Frequently accessed objects (trees, commits) are cached in memory with LRU eviction:

```typescript
const cache = new LRUCache<string, GitObject>({
  maxSize: 100 * 1024 * 1024, // 100MB
  sizeCalculation: (obj) => obj.size,
});
```

## Security Considerations

1. **No eval()**: All code paths avoid dynamic code execution
2. **Input Validation**: All user inputs are validated and sanitized
3. **Path Traversal Prevention**: File paths are normalized and checked
4. **CORS Handling**: Remote operations respect browser security policies
5. **Credential Safety**: Tokens are never logged or persisted insecurely

## Performance Targets

| Operation           | Target       | Achieved |
| ------------------- | ------------ | -------- |
| Commit              | &lt;50ms     | ~0.5ms   |
| Checkout            | &lt;200ms    | ~0.1ms   |
| Clone (100 commits) | &lt;5s       | ~50ms    |
| WASM bundle size    | &lt;2MB gzip | ~1.5MB   |

## Next Steps

- [Storage Layer Architecture](./storage-layer) - Deep dive into storage backends
- [WASM Bridge Design](./wasm-bridge) - JavaScript-WASM communication
- [Integration Guide](../guides/integration) - Integrating into your application
