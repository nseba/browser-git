---
sidebar_position: 3
---

# WASM Bridge Design

BrowserGit uses WebAssembly to run performance-critical Git algorithms written in Go. This document explains how the JavaScript and WASM layers communicate.

## Overview

The WASM bridge provides a seamless interface between TypeScript and Go:

```
┌─────────────────────────────────────────────────────────────┐
│                     TypeScript Layer                         │
│                                                             │
│   Repository.commit()  ─────────►  wasmBridge.createCommit()│
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      WASM Bridge                             │
│                                                             │
│   ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│   │   Encoder    │    │   Memory     │    │   Decoder    │ │
│   │ (JS → WASM)  │    │  Management  │    │ (WASM → JS)  │ │
│   └──────────────┘    └──────────────┘    └──────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   WebAssembly Module                         │
│                        (Go)                                  │
│                                                             │
│   //export CreateCommit                                     │
│   func CreateCommit(treePtr, msgPtr, authorPtr int32) int32 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Loading the WASM Module

### Initialization

```typescript
import { WasmLoader } from "@browser-git/browser-git";

// Load the WASM module
const wasmModule = await WasmLoader.load("/git-core.wasm");

// The module is now ready to use
const hash = wasmModule.sha1(data);
```

### Lazy Loading

For better initial load performance, WASM is loaded lazily:

```typescript
class WasmLoader {
  private static instance: WebAssembly.Instance | null = null;
  private static loading: Promise<WebAssembly.Instance> | null = null;

  static async getInstance(): Promise<WebAssembly.Instance> {
    if (this.instance) {
      return this.instance;
    }

    if (!this.loading) {
      this.loading = this.loadModule();
    }

    return this.loading;
  }

  private static async loadModule(): Promise<WebAssembly.Instance> {
    const response = await fetch("/git-core.wasm");
    const { instance } = await WebAssembly.instantiateStreaming(response, {
      env: this.createEnvironment(),
    });

    this.instance = instance;
    return instance;
  }
}
```

## Memory Management

### WASM Linear Memory

WebAssembly uses a linear memory model. Data must be copied between JavaScript and WASM memory:

```typescript
class WasmMemory {
  private memory: WebAssembly.Memory;
  private allocator: WasmAllocator;

  // Allocate space in WASM memory
  allocate(size: number): number {
    return this.allocator.malloc(size);
  }

  // Free WASM memory
  free(ptr: number): void {
    this.allocator.free(ptr);
  }

  // Copy JavaScript data to WASM
  copyToWasm(data: Uint8Array): number {
    const ptr = this.allocate(data.length);
    const view = new Uint8Array(this.memory.buffer, ptr, data.length);
    view.set(data);
    return ptr;
  }

  // Copy WASM data to JavaScript
  copyFromWasm(ptr: number, length: number): Uint8Array {
    const view = new Uint8Array(this.memory.buffer, ptr, length);
    return new Uint8Array(view); // Create a copy
  }
}
```

### Automatic Memory Management

The bridge tracks allocations and cleans up automatically:

```typescript
class WasmBridge {
  async withMemory<T>(
    data: Uint8Array[],
    fn: (ptrs: number[]) => T,
  ): Promise<T> {
    const ptrs: number[] = [];

    try {
      // Allocate all data
      for (const d of data) {
        ptrs.push(this.memory.copyToWasm(d));
      }

      // Execute function
      return fn(ptrs);
    } finally {
      // Always free memory
      for (const ptr of ptrs) {
        this.memory.free(ptr);
      }
    }
  }
}
```

## Data Serialization

### String Encoding

Strings are encoded as UTF-8:

```typescript
const encoder = new TextEncoder();
const decoder = new TextDecoder();

function stringToWasm(str: string): Uint8Array {
  return encoder.encode(str);
}

function wasmToString(data: Uint8Array): string {
  return decoder.decode(data);
}
```

### Complex Objects

Complex objects are serialized using a binary protocol:

```typescript
interface SerializedCommit {
  treeHash: Uint8Array; // 20 bytes (SHA-1) or 32 bytes (SHA-256)
  parentHashes: Uint8Array; // Array of parent hashes
  authorName: Uint8Array; // UTF-8 encoded
  authorEmail: Uint8Array; // UTF-8 encoded
  timestamp: number; // Unix timestamp
  message: Uint8Array; // UTF-8 encoded
}

function serializeCommit(commit: Commit): Uint8Array {
  const buffer = new ArrayBuffer(calculateSize(commit));
  const view = new DataView(buffer);

  let offset = 0;

  // Tree hash
  view.setUint8(offset++, commit.treeHash.length);
  new Uint8Array(buffer, offset).set(commit.treeHash);
  offset += commit.treeHash.length;

  // ... continue with other fields

  return new Uint8Array(buffer);
}
```

## Function Exports

### Go Side

Functions are exported using TinyGo's export directive:

```go
package main

import "unsafe"

//export CreateCommit
func CreateCommit(
    treePtr, treeLen int32,
    msgPtr, msgLen int32,
    authorPtr, authorLen int32,
    resultPtr int32,
) int32 {
    // Read data from WASM memory
    tree := readBytes(treePtr, treeLen)
    msg := readString(msgPtr, msgLen)
    author := readString(authorPtr, authorLen)

    // Create commit
    commit, err := createCommitInternal(tree, msg, author)
    if err != nil {
        return writeError(resultPtr, err)
    }

    // Write result to WASM memory
    return writeResult(resultPtr, commit)
}

func readBytes(ptr, len int32) []byte {
    return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), len)
}

func readString(ptr, len int32) string {
    return string(readBytes(ptr, len))
}
```

### JavaScript Side

```typescript
interface WasmExports {
  CreateCommit(
    treePtr: number,
    treeLen: number,
    msgPtr: number,
    msgLen: number,
    authorPtr: number,
    authorLen: number,
    resultPtr: number,
  ): number;

  SHA1(dataPtr: number, dataLen: number, resultPtr: number): void;
  SHA256(dataPtr: number, dataLen: number, resultPtr: number): void;

  ParsePackfile(dataPtr: number, dataLen: number, resultPtr: number): number;
  ApplyDelta(
    basePtr: number,
    baseLen: number,
    deltaPtr: number,
    deltaLen: number,
    resultPtr: number,
  ): number;
}
```

## Error Handling

### Error Protocol

Errors are returned as a special structure:

```typescript
interface WasmResult {
  success: boolean;
  errorCode: number;
  errorMessage: string;
  data: Uint8Array;
}

function parseResult(resultPtr: number, resultLen: number): WasmResult {
  const data = memory.copyFromWasm(resultPtr, resultLen);
  const view = new DataView(data.buffer);

  const success = view.getUint8(0) === 1;

  if (!success) {
    const errorCode = view.getUint32(1, true);
    const messageLen = view.getUint32(5, true);
    const message = decoder.decode(data.slice(9, 9 + messageLen));

    return {
      success: false,
      errorCode,
      errorMessage: message,
      data: new Uint8Array(),
    };
  }

  return {
    success: true,
    errorCode: 0,
    errorMessage: "",
    data: data.slice(1),
  };
}
```

### JavaScript Error Mapping

```typescript
class WasmError extends Error {
  constructor(
    public code: number,
    message: string,
  ) {
    super(message);
    this.name = "WasmError";
  }
}

const ERROR_CODES = {
  0: "Success",
  1: "InvalidObject",
  2: "CorruptedData",
  3: "OutOfMemory",
  4: "InvalidHash",
  5: "DeltaApplicationFailed",
} as const;
```

## Performance Optimization

### Batch Operations

Minimize crossing the JS-WASM boundary:

```typescript
// Slow: Multiple boundary crossings
for (const obj of objects) {
  const hash = wasm.sha1(obj);
  hashes.push(hash);
}

// Fast: Single boundary crossing
const hashes = wasm.sha1Batch(objects);
```

### Memory Pooling

Reuse allocated memory for repeated operations:

```typescript
class MemoryPool {
  private pools: Map<number, number[]> = new Map();

  acquire(size: number): number {
    // Round up to power of 2 for pooling
    const poolSize = nextPowerOf2(size);
    const pool = this.pools.get(poolSize) || [];

    if (pool.length > 0) {
      return pool.pop()!;
    }

    return memory.allocate(poolSize);
  }

  release(ptr: number, size: number): void {
    const poolSize = nextPowerOf2(size);
    const pool = this.pools.get(poolSize) || [];
    pool.push(ptr);
    this.pools.set(poolSize, pool);
  }
}
```

### Streaming Large Data

For large files, stream data through WASM:

```typescript
async function* streamThroughWasm(
  input: ReadableStream<Uint8Array>,
  processChunk: (ptr: number, len: number) => void,
): AsyncGenerator<Uint8Array> {
  const reader = input.getReader();
  const chunkSize = 64 * 1024; // 64KB chunks
  const chunkPtr = memory.allocate(chunkSize);

  try {
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      // Process in chunks
      for (let i = 0; i < value.length; i += chunkSize) {
        const chunk = value.slice(i, i + chunkSize);
        memory.copyToWasm(chunk, chunkPtr);
        processChunk(chunkPtr, chunk.length);
      }
    }
  } finally {
    memory.free(chunkPtr);
    reader.releaseLock();
  }
}
```

## Testing the Bridge

```typescript
describe("WASM Bridge", () => {
  let bridge: WasmBridge;

  beforeAll(async () => {
    bridge = await WasmBridge.create();
  });

  it("should compute SHA-1 correctly", async () => {
    const data = new TextEncoder().encode("hello world");
    const hash = await bridge.sha1(data);

    expect(hash).toBe("2aae6c35c94fcfb415dbe95f408b9ce91ee846ed");
  });

  it("should handle large data", async () => {
    const data = new Uint8Array(10 * 1024 * 1024); // 10MB
    crypto.getRandomValues(data);

    const hash = await bridge.sha1(data);
    expect(hash).toHaveLength(40);
  });

  it("should handle errors gracefully", async () => {
    const corruptedPackfile = new Uint8Array([0, 1, 2, 3]);

    await expect(bridge.parsePackfile(corruptedPackfile)).rejects.toThrow(
      "Invalid packfile header",
    );
  });
});
```

## Next Steps

- [Storage Layer](./storage-layer) - How data is persisted
- [Performance Optimization](../guides/integration#performance-optimization) - Tuning for your use case
- [API Reference](../api/repository) - Complete API documentation
