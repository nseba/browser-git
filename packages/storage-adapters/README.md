# @browser-git/storage-adapters

Storage adapter implementations for BrowserGit supporting multiple browser storage backends.

## Features

- **IndexedDB**: Primary storage with large capacity and async operations
- **OPFS**: Origin Private File System for modern browsers with best performance
- **LocalStorage**: Fallback for limited storage scenarios
- **In-Memory**: Fast, ephemeral storage for testing

## Installation

```bash
npm install @browser-git/storage-adapters
```

## Usage

```typescript
import {
  IndexedDBAdapter,
  OPFSAdapter,
  LocalStorageAdapter,
  MemoryAdapter,
} from "@browser-git/storage-adapters";

// Use IndexedDB (recommended)
const storage = new IndexedDBAdapter("my-repo");
await storage.set("key", new Uint8Array([1, 2, 3]));
const data = await storage.get("key");

// Use OPFS (best performance on modern browsers)
const opfsStorage = new OPFSAdapter("my-repo");
await opfsStorage.set("file.txt", new TextEncoder().encode("content"));

// Use LocalStorage (limited capacity)
const localStorage = new LocalStorageAdapter("my-repo");
await localStorage.set("small-key", new Uint8Array([1, 2, 3]));

// Use in-memory (testing)
const memStorage = new MemoryAdapter();
await memStorage.set("test-key", new Uint8Array([1, 2, 3]));
```

## Storage Interface

All adapters implement the common `StorageAdapter` interface:

```typescript
interface StorageAdapter {
  get(key: string): Promise<Uint8Array | null>;
  set(key: string, value: Uint8Array): Promise<void>;
  delete(key: string): Promise<void>;
  list(prefix?: string): Promise<string[]>;
  exists(key: string): Promise<boolean>;
  clear(): Promise<void>;
}
```

## Browser Compatibility

| Adapter      | Chrome | Firefox | Safari   |
| ------------ | ------ | ------- | -------- |
| IndexedDB    | ✅     | ✅      | ✅       |
| OPFS         | ✅ 86+ | ✅ 111+ | ✅ 15.2+ |
| LocalStorage | ✅     | ✅      | ✅       |
| Memory       | ✅     | ✅      | ✅       |

## License

MIT
