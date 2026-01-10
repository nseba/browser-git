---
sidebar_position: 5
---

# Browser Compatibility

BrowserGit is designed to work across all modern browsers. This page documents browser support, feature availability, and known limitations.

## Support Matrix

### Desktop Browsers

| Feature | Chrome | Firefox | Safari | Edge |
|---------|--------|---------|--------|------|
| **Core Git Operations** | 86+ | 111+ | 15.2+ | 86+ |
| WebAssembly | ✅ | ✅ | ✅ | ✅ |
| IndexedDB | ✅ | ✅ | ✅ | ✅ |
| OPFS | ✅ | ✅ | ⚠️ | ✅ |
| Compression Streams | ✅ | ✅ | ✅ | ✅ |
| Web Crypto | ✅ | ✅ | ✅ | ✅ |
| Fetch API | ✅ | ✅ | ✅ | ✅ |
| Streams API | ✅ | ✅ | ✅ | ✅ |

### Mobile Browsers

| Feature | Chrome Mobile | Safari Mobile | Firefox Mobile | Samsung Internet |
|---------|---------------|---------------|----------------|------------------|
| **Core Git Operations** | ✅ | ✅ | ✅ | ✅ |
| WebAssembly | ✅ | ✅ | ✅ | ✅ |
| IndexedDB | ✅ | ✅ | ✅ | ✅ |
| OPFS | ✅ | ⚠️ | ✅ | ✅ |
| Compression Streams | ✅ | ✅ | ✅ | ✅ |

**Legend:**
- ✅ Full support
- ⚠️ Partial support / Known issues
- ❌ Not supported

## Feature Details

### WebAssembly

**Required for:** Core Git operations (hashing, packfile parsing, delta resolution)

```typescript
// Check WebAssembly support
const wasmSupported = typeof WebAssembly !== 'undefined' &&
                      typeof WebAssembly.instantiate === 'function';
```

**Browser versions:**
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 16+

### IndexedDB

**Required for:** Persistent storage (default adapter)

```typescript
// Check IndexedDB support
const idbSupported = typeof indexedDB !== 'undefined';
```

All modern browsers support IndexedDB. Some considerations:

- **Safari Private Browsing**: Limited quota (~50MB)
- **Firefox Private Browsing**: Full support, cleared on session end
- **Chrome Incognito**: Full support, cleared on session end

### OPFS (Origin Private File System)

**Used for:** High-performance file storage

```typescript
// Check OPFS support
const opfsSupported = typeof navigator?.storage?.getDirectory === 'function';
```

**Known Issues:**

| Browser | Issue | Workaround |
|---------|-------|------------|
| Safari | Throws `UnknownError` on some operations | Use IndexedDB instead |
| Safari < 15.2 | Not available | Use IndexedDB |
| Firefox < 111 | Not available | Use IndexedDB |

**Detection and Fallback:**

```typescript
import { detectFeatures, createAdapter } from '@browser-git/storage-adapters';

async function getBestAdapter(name: string) {
  const features = await detectFeatures();

  if (features.opfs && !features.opfsHasBugs) {
    return new OPFSAdapter(name);
  }

  return new IndexedDBAdapter(name);
}
```

### Compression Streams

**Used for:** Packfile compression, object storage

```typescript
// Check Compression Streams support
const compressionSupported = typeof CompressionStream !== 'undefined';
```

**Fallback:** BrowserGit includes a JavaScript fallback implementation using pako.

### performance.memory API

**Used for:** Memory usage tracking (optional)

```typescript
// Check memory API support (Chrome only)
const memoryApiSupported = typeof performance?.memory !== 'undefined';
```

**Note:** Only available in Chromium-based browsers. BrowserGit works without it but can't report memory metrics on other browsers.

## Storage Quotas

Browser storage quotas vary significantly:

| Browser | Default Quota | Notes |
|---------|--------------|-------|
| Chrome | 60% of disk | Eviction possible under pressure |
| Firefox | 50% of disk | Prompts at 2GB |
| Safari | 1GB | Strict quota |
| Safari Mobile | 500MB | More restrictive |
| Edge | Same as Chrome | Chromium-based |

### Checking Available Storage

```typescript
async function checkStorage() {
  if (navigator.storage && navigator.storage.estimate) {
    const estimate = await navigator.storage.estimate();
    console.log(`Used: ${estimate.usage} of ${estimate.quota} bytes`);
    console.log(`Available: ${estimate.quota - estimate.usage} bytes`);
  }
}
```

### Requesting Persistent Storage

```typescript
async function requestPersistence() {
  if (navigator.storage && navigator.storage.persist) {
    const granted = await navigator.storage.persist();
    if (granted) {
      console.log('Storage will not be cleared automatically');
    }
  }
}
```

## Known Browser-Specific Issues

### Safari

1. **OPFS Unreliable**: Safari's OPFS implementation may throw `UnknownError` on valid operations. Always use IndexedDB as fallback.

2. **Private Browsing Quota**: Very limited storage in private mode.

3. **ITP (Intelligent Tracking Prevention)**: May affect cross-origin storage.

**Recommended Configuration:**

```typescript
const repo = await Repository.init('/project', {
  storage: 'indexeddb' // Avoid OPFS on Safari
});
```

### Firefox

1. **Memory API**: `performance.memory` not available. Memory tracking won't work.

2. **Large ArrayBuffers**: Some older versions have issues with very large ArrayBuffers.

### Chrome

1. **Storage Eviction**: Under storage pressure, Chrome may evict IndexedDB data. Request persistent storage for important repositories.

```typescript
await navigator.storage.persist();
```

### Mobile Browsers

1. **Background Tabs**: Browsers may kill background tabs, interrupting long operations.

2. **Memory Limits**: Lower memory limits may affect large repositories.

3. **Storage Limits**: Generally more restrictive than desktop.

## Feature Detection

BrowserGit includes comprehensive feature detection:

```typescript
import { detectBrowserFeatures } from '@browser-git/browser-git';

const features = await detectBrowserFeatures();

console.log({
  wasm: features.webAssembly,
  wasmStreaming: features.webAssemblyStreaming,
  indexedDB: features.indexedDB,
  opfs: features.opfs,
  opfsWorking: features.opfsWorking, // Actually works, not just API present
  compression: features.compressionStreams,
  crypto: features.webCrypto,
  memory: features.memoryAPI,
  persistentStorage: features.persistentStorage,
  storageQuota: features.storageQuota
});

// Get recommendations
const recommendations = features.getRecommendations();
console.log('Recommended storage:', recommendations.storage);
console.log('Warnings:', recommendations.warnings);
```

## Polyfills

BrowserGit includes polyfills for some features:

| Feature | Polyfill | Impact |
|---------|----------|--------|
| Compression Streams | pako | Slightly slower compression |
| TextEncoder/Decoder | Built-in | Negligible |
| crypto.randomUUID | Built-in | Negligible |

**Enabling Polyfills:**

```typescript
import { enablePolyfills } from '@browser-git/browser-git';

// Automatically detect and apply needed polyfills
await enablePolyfills();
```

## Testing Your Environment

Run the compatibility check:

```typescript
import { runCompatibilityCheck } from '@browser-git/browser-git';

const result = await runCompatibilityCheck();

if (result.compatible) {
  console.log('Full compatibility');
} else if (result.degraded) {
  console.log('Degraded mode:', result.limitations);
} else {
  console.log('Not compatible:', result.reason);
}
```

## Minimum Requirements

For BrowserGit to function at all:

- **WebAssembly**: Required
- **IndexedDB or LocalStorage**: At least one required
- **Fetch API**: Required for remote operations
- **Modern JavaScript**: ES2020+ features used

## See Also

- [Storage Adapters](./api/storage-adapters) - Storage backend details
- [Limitations](./limitations) - Known limitations
- [Architecture Overview](./architecture/overview) - How components interact
