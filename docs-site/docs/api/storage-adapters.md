---
sidebar_position: 3
---

# Storage Adapters API

Storage adapters provide the persistence layer for BrowserGit, abstracting browser storage APIs behind a common interface.

## StorageAdapter Interface

All adapters implement this interface:

```typescript
interface StorageAdapter {
  // Lifecycle
  initialize(): Promise<void>;
  close(): Promise<void>;

  // Basic operations
  get(key: string): Promise<Uint8Array | null>;
  set(key: string, value: Uint8Array): Promise<void>;
  delete(key: string): Promise<void>;
  exists(key: string): Promise<boolean>;

  // Listing
  list(prefix?: string): Promise<string[]>;

  // Batch operations
  getBatch(keys: string[]): Promise<Map<string, Uint8Array>>;
  setBatch(entries: Map<string, Uint8Array>): Promise<void>;
  deleteBatch(keys: string[]): Promise<void>;

  // Management
  clear(): Promise<void>;
  getQuota(): Promise<StorageQuota>;
}

interface StorageQuota {
  usage: number; // Bytes used
  quota: number; // Total available
  available: number; // Remaining
}
```

## IndexedDBAdapter

Primary storage adapter for most use cases.

### Constructor

```typescript
constructor(name: string, options?: IndexedDBOptions)
```

**Options:**

```typescript
interface IndexedDBOptions {
  version?: number;
  storeName?: string;
}
```

### Usage

```typescript
import { IndexedDBAdapter } from "@browser-git/storage-adapters";

const adapter = new IndexedDBAdapter("my-repo");
await adapter.initialize();

// Store data
await adapter.set("objects/abc123", new Uint8Array([1, 2, 3]));

// Retrieve data
const data = await adapter.get("objects/abc123");

// Check existence
const exists = await adapter.exists("objects/abc123");

// List keys
const keys = await adapter.list("objects/");

// Delete
await adapter.delete("objects/abc123");

// Batch operations
await adapter.setBatch(
  new Map([
    ["key1", new Uint8Array([1])],
    ["key2", new Uint8Array([2])],
  ]),
);

const batch = await adapter.getBatch(["key1", "key2"]);

// Get quota
const quota = await adapter.getQuota();
console.log(`Used: ${quota.usage} / ${quota.quota}`);

// Cleanup
await adapter.clear();
await adapter.close();
```

### Browser Support

| Browser | Version | Notes        |
| ------- | ------- | ------------ |
| Chrome  | 23+     | Full support |
| Firefox | 10+     | Full support |
| Safari  | 10+     | Full support |
| Edge    | 12+     | Full support |

## OPFSAdapter

Origin Private File System adapter for high-performance storage.

### Constructor

```typescript
constructor(name: string, options?: OPFSOptions)
```

**Options:**

```typescript
interface OPFSOptions {
  rootPath?: string;
}
```

### Usage

```typescript
import { OPFSAdapter } from "@browser-git/storage-adapters";

const adapter = new OPFSAdapter("my-repo");
await adapter.initialize();

// Same interface as IndexedDBAdapter
await adapter.set("objects/abc123", data);
const retrieved = await adapter.get("objects/abc123");
```

### Browser Support

| Browser | Version | Notes                     |
| ------- | ------- | ------------------------- |
| Chrome  | 86+     | Full support              |
| Firefox | 111+    | Full support              |
| Safari  | 15.2+   | Limited, may throw errors |
| Edge    | 86+     | Full support              |

### Feature Detection

```typescript
import { isOPFSSupported } from "@browser-git/storage-adapters";

if (await isOPFSSupported()) {
  const adapter = new OPFSAdapter("my-repo");
} else {
  const adapter = new IndexedDBAdapter("my-repo");
}
```

## LocalStorageAdapter

Fallback adapter for limited environments.

### Constructor

```typescript
constructor(prefix: string)
```

### Usage

```typescript
import { LocalStorageAdapter } from "@browser-git/storage-adapters";

const adapter = new LocalStorageAdapter("my-repo");
await adapter.initialize();

// Data is base64 encoded internally
await adapter.set("key", new Uint8Array([1, 2, 3]));
```

### Limitations

- **Size limit**: ~5-10MB per origin
- **Synchronous**: Blocks main thread
- **Base64 overhead**: ~33% size increase

### Browser Support

All modern browsers support LocalStorage.

## MemoryAdapter

In-memory adapter for testing and temporary storage.

### Constructor

```typescript
constructor();
```

### Usage

```typescript
import { MemoryAdapter } from "@browser-git/storage-adapters";

const adapter = new MemoryAdapter();
await adapter.initialize();

// Perfect for testing
await adapter.set("test-key", testData);

// Data is lost when adapter is garbage collected
```

## MockAdapter

Programmable mock for testing.

### Constructor

```typescript
constructor(responses?: MockResponses)
```

### Usage

```typescript
import { MockAdapter } from "@browser-git/storage-adapters";

const mock = new MockAdapter({
  "objects/abc": new Uint8Array([1, 2, 3]),
});

await mock.initialize();

// Returns predefined response
const data = await mock.get("objects/abc");

// Track calls
console.log(mock.calls);
// [{ method: 'get', args: ['objects/abc'] }]

// Simulate errors
mock.setError("get", new Error("Network error"));
```

## Adapter Factory

Create the best available adapter:

```typescript
import { createAdapter, AdapterType } from "@browser-git/storage-adapters";

// Automatic selection
const adapter = await createAdapter("my-repo");

// Explicit selection
const indexed = await createAdapter("my-repo", AdapterType.IndexedDB);
const opfs = await createAdapter("my-repo", AdapterType.OPFS);
const local = await createAdapter("my-repo", AdapterType.LocalStorage);
const memory = await createAdapter("my-repo", AdapterType.Memory);
```

## Feature Detection

```typescript
import { detectFeatures } from "@browser-git/storage-adapters";

const features = await detectFeatures();

console.log("IndexedDB:", features.indexeddb);
console.log("OPFS:", features.opfs);
console.log("LocalStorage:", features.localstorage);
console.log("Storage Quota:", features.quota);
```

## Wrapper Adapters

### CachedAdapter

Adds LRU caching layer:

```typescript
import { CachedAdapter } from "@browser-git/storage-adapters";

const cached = new CachedAdapter(baseAdapter, {
  maxSize: 50 * 1024 * 1024, // 50MB
  maxAge: 5 * 60 * 1000, // 5 minutes
});

// Reads are cached
const data = await cached.get("key"); // From storage
const data2 = await cached.get("key"); // From cache

// Clear cache
cached.clearCache();
```

### CompressionAdapter

Adds compression:

```typescript
import { CompressionAdapter } from "@browser-git/storage-adapters";

const compressed = new CompressionAdapter(baseAdapter, {
  threshold: 1024, // Compress data > 1KB
  algorithm: "gzip",
});

// Data is automatically compressed/decompressed
await compressed.set("large-key", largeData);
```

### EncryptedAdapter

Adds encryption:

```typescript
import { EncryptedAdapter } from "@browser-git/storage-adapters";

const encrypted = new EncryptedAdapter(baseAdapter, {
  key: cryptoKey, // Web Crypto API key
});

// Data is encrypted at rest
await encrypted.set("secret", sensitiveData);
```

## Error Handling

```typescript
import {
  StorageError,
  QuotaExceededError,
} from "@browser-git/storage-adapters";

try {
  await adapter.set("large-key", veryLargeData);
} catch (error) {
  if (error instanceof QuotaExceededError) {
    console.log("Storage full:", error.quota);
    // Prompt user to free space
  } else if (error instanceof StorageError) {
    console.log("Storage error:", error.code, error.message);
  }
}
```

## Testing with Adapters

```typescript
import { MemoryAdapter } from "@browser-git/storage-adapters";
import { Repository } from "@browser-git/browser-git";

describe("My Git Tests", () => {
  let adapter: MemoryAdapter;
  let repo: Repository;

  beforeEach(async () => {
    adapter = new MemoryAdapter();
    await adapter.initialize();

    repo = await Repository.init("/test", {
      storage: adapter, // Pass adapter directly
    });
  });

  afterEach(async () => {
    await adapter.close();
  });

  it("should create commits", async () => {
    await repo.fs.writeFile("/test/file.txt", "content");
    await repo.add(["file.txt"]);
    await repo.commit("test", { author });

    const log = await repo.log();
    expect(log).toHaveLength(1);
  });
});
```

## See Also

- [Storage Layer Architecture](../architecture/storage-layer) - Design details
- [Repository API](./repository) - Using adapters with repositories
- [Browser Compatibility](../browser-compatibility) - Browser support matrix
