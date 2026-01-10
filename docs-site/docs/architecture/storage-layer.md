---
sidebar_position: 2
---

# Storage Layer Architecture

BrowserGit uses a pluggable storage layer that abstracts browser storage APIs behind a common interface. This allows repositories to be stored using the most appropriate backend for each use case.

## Storage Interface

All storage adapters implement the `StorageAdapter` interface:

```typescript
interface StorageAdapter {
  // Basic operations
  get(key: string): Promise<Uint8Array | null>;
  set(key: string, value: Uint8Array): Promise<void>;
  delete(key: string): Promise<void>;
  exists(key: string): Promise<boolean>;

  // Listing and iteration
  list(prefix?: string): Promise<string[]>;

  // Batch operations
  getBatch(keys: string[]): Promise<Map<string, Uint8Array>>;
  setBatch(entries: Map<string, Uint8Array>): Promise<void>;
  deleteBatch(keys: string[]): Promise<void>;

  // Management
  clear(): Promise<void>;
  close(): Promise<void>;

  // Metadata
  getQuota(): Promise<StorageQuota>;
}

interface StorageQuota {
  usage: number;
  quota: number;
  available: number;
}
```

## Available Adapters

### IndexedDB Adapter

The recommended storage backend for most applications.

**Advantages:**

- Wide browser support (all modern browsers)
- Large storage quota (typically 50% of disk space)
- Transactional consistency
- Good performance for most operations

**Limitations:**

- Slower for many small files compared to OPFS
- Requires serialization of all data

**Usage:**

```typescript
import { IndexedDBAdapter } from "@browser-git/storage-adapters";

const adapter = new IndexedDBAdapter("my-repo");
await adapter.initialize();

// Store data
await adapter.set("objects/abc123", objectData);

// Retrieve data
const data = await adapter.get("objects/abc123");
```

**Internal Structure:**

```
IndexedDB Database: "browser-git-{repo-name}"
├── Object Store: "objects"
│   └── Key-Value pairs (path → Uint8Array)
├── Object Store: "refs"
│   └── Key-Value pairs (ref-name → commit-hash)
└── Object Store: "metadata"
    └── Repository configuration
```

### OPFS Adapter

Origin Private File System - best performance for file-heavy operations.

**Advantages:**

- True file system semantics
- Best performance for streaming operations
- Efficient for large files
- Synchronous access handle API (in workers)

**Limitations:**

- Limited browser support (Chrome 86+, Firefox 111+, Safari 15.2+ with issues)
- WebKit has known bugs with OPFS
- Requires more complex error handling

**Usage:**

```typescript
import { OPFSAdapter } from "@browser-git/storage-adapters";

const adapter = new OPFSAdapter("my-repo");
await adapter.initialize();

// Works like a regular filesystem
await adapter.set("objects/abc123", objectData);
```

**Internal Structure:**

```
OPFS Root
└── browser-git/
    └── {repo-name}/
        ├── objects/
        │   ├── pack/
        │   └── loose/
        ├── refs/
        │   ├── heads/
        │   └── tags/
        └── config
```

### LocalStorage Adapter

Fallback for environments with limited API support.

**Advantages:**

- Universal browser support
- Synchronous API (simpler code paths)
- Persistent across sessions

**Limitations:**

- Very limited storage (5-10MB typical)
- Only stores strings (requires base64 encoding for binary)
- Blocks main thread

**Usage:**

```typescript
import { LocalStorageAdapter } from "@browser-git/storage-adapters";

const adapter = new LocalStorageAdapter("my-repo");
await adapter.initialize();

// Note: Data is base64 encoded internally
await adapter.set("small-file", smallData);
```

### Memory Adapter

Ephemeral storage for testing and temporary operations.

**Advantages:**

- Fastest possible performance
- No persistence overhead
- Predictable behavior for testing

**Limitations:**

- Data lost on page refresh
- Limited by available RAM

**Usage:**

```typescript
import { MemoryAdapter } from "@browser-git/storage-adapters";

const adapter = new MemoryAdapter();
await adapter.initialize();

// Perfect for testing
await adapter.set("test-key", testData);
```

## Storage Selection

BrowserGit automatically selects the best available storage:

```typescript
import {
  createBestAdapter,
  detectFeatures,
} from "@browser-git/storage-adapters";

// Automatic selection
const adapter = await createBestAdapter("my-repo");

// Or check features manually
const features = await detectFeatures();
console.log("OPFS available:", features.opfs);
console.log("IndexedDB available:", features.indexeddb);
console.log("Available quota:", features.quota);
```

**Selection Priority:**

1. **OPFS** - if available and quota sufficient
2. **IndexedDB** - if available (almost always)
3. **LocalStorage** - last resort fallback

## Data Organization

Git data is organized within storage adapters:

```
{storage-root}/
├── objects/
│   ├── pack/
│   │   ├── pack-{hash}.pack    # Packfile data
│   │   └── pack-{hash}.idx     # Packfile index
│   └── {xx}/
│       └── {hash}              # Loose objects
├── refs/
│   ├── heads/
│   │   └── {branch-name}       # Branch references
│   ├── tags/
│   │   └── {tag-name}          # Tag references
│   └── remotes/
│       └── {remote}/
│           └── {branch}        # Remote tracking refs
├── HEAD                         # Current reference
├── index                        # Staging area
└── config                       # Repository config
```

## Transaction Handling

### IndexedDB Transactions

```typescript
class IndexedDBAdapter {
  async setBatch(entries: Map<string, Uint8Array>): Promise<void> {
    return new Promise((resolve, reject) => {
      const tx = this.db.transaction(["objects"], "readwrite");
      const store = tx.objectStore("objects");

      for (const [key, value] of entries) {
        store.put(value, key);
      }

      tx.oncomplete = () => resolve();
      tx.onerror = () => reject(tx.error);
    });
  }
}
```

### OPFS Atomic Writes

```typescript
class OPFSAdapter {
  async set(key: string, value: Uint8Array): Promise<void> {
    const tempPath = `${key}.tmp`;

    // Write to temporary file
    const tempHandle = await this.getFileHandle(tempPath, { create: true });
    const writable = await tempHandle.createWritable();
    await writable.write(value);
    await writable.close();

    // Atomic rename
    await this.rename(tempPath, key);
  }
}
```

## Quota Management

Monitor and manage storage usage:

```typescript
import { StorageQuotaManager } from "@browser-git/storage-adapters";

const manager = new StorageQuotaManager(adapter);

// Check available space
const quota = await manager.getQuota();
console.log(`Used: ${quota.usage} / ${quota.quota} bytes`);

// Request more space if needed
if (quota.available < requiredSpace) {
  const granted = await manager.requestQuota(requiredSpace);
  if (!granted) {
    throw new Error("Insufficient storage quota");
  }
}

// Cleanup old objects
await manager.pruneUnreachableObjects();
```

## Performance Optimization

### Batch Operations

Always prefer batch operations for multiple items:

```typescript
// Slow: Multiple round-trips
for (const key of keys) {
  await adapter.get(key);
}

// Fast: Single batch operation
const results = await adapter.getBatch(keys);
```

### Caching Layer

BrowserGit includes an optional caching layer:

```typescript
import { CachedAdapter } from "@browser-git/storage-adapters";

const cached = new CachedAdapter(baseAdapter, {
  maxSize: 50 * 1024 * 1024, // 50MB cache
  ttl: 5 * 60 * 1000, // 5 minute TTL
});
```

### Compression

Large objects are automatically compressed:

```typescript
import { CompressionAdapter } from "@browser-git/storage-adapters";

const compressed = new CompressionAdapter(baseAdapter, {
  threshold: 1024, // Compress objects > 1KB
  algorithm: "gzip",
});
```

## Error Handling

Storage operations can fail for various reasons:

```typescript
import {
  StorageError,
  QuotaExceededError,
} from "@browser-git/storage-adapters";

try {
  await adapter.set(key, largeData);
} catch (error) {
  if (error instanceof QuotaExceededError) {
    console.error(
      "Storage quota exceeded:",
      error.available,
      "bytes available",
    );
    // Prompt user to free space or use different storage
  } else if (error instanceof StorageError) {
    console.error("Storage error:", error.code, error.message);
  }
}
```

## Testing Storage

The Memory adapter is ideal for testing:

```typescript
import { MemoryAdapter } from "@browser-git/storage-adapters";
import { Repository } from "@browser-git/browser-git";

describe("My Git tests", () => {
  let repo: Repository;

  beforeEach(async () => {
    repo = await Repository.init("/test", {
      storage: "memory",
    });
  });

  it("should commit files", async () => {
    await repo.fs.writeFile("/test/file.txt", "content");
    await repo.add(["file.txt"]);
    await repo.commit("test commit", { author });

    const log = await repo.log();
    expect(log).toHaveLength(1);
  });
});
```

## Next Steps

- [WASM Bridge Design](./wasm-bridge) - How JavaScript communicates with Go
- [Browser Compatibility](../browser-compatibility) - Detailed browser support matrix
- [Performance Optimization](../guides/integration#performance-optimization) - Tuning for your use case
