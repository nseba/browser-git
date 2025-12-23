# Browser Compatibility

BrowserGit is designed to work across all modern browsers. This document outlines browser compatibility, feature support, and recommended configurations.

## Quick Reference

| Browser       | Version | Status             | Storage                         | WASM | Notes                   |
| ------------- | ------- | ------------------ | ------------------------------- | ---- | ----------------------- |
| Chrome        | 90+     | ✅ Full Support    | IndexedDB, OPFS, localStorage   | Full | Best performance        |
| Edge          | 90+     | ✅ Full Support    | IndexedDB, OPFS, localStorage   | Full | Chromium-based          |
| Firefox       | 88+     | ✅ Full Support    | IndexedDB, OPFS\*, localStorage | Full | OPFS in recent versions |
| Safari        | 15+     | ⚠️ Partial Support | IndexedDB, localStorage         | Full | No OPFS support         |
| Mobile Safari | 15+     | ⚠️ Partial Support | IndexedDB, localStorage         | Full | Limited storage quota   |
| Mobile Chrome | 90+     | ✅ Full Support    | IndexedDB, OPFS, localStorage   | Full | -                       |

\*OPFS support in Firefox requires version 111+

## Minimum Requirements

BrowserGit requires the following browser features:

### Essential Features

- ✅ WebAssembly support
- ✅ SubtleCrypto API (for SHA-1/SHA-256 hashing)
- ✅ At least one persistent storage option (IndexedDB, OPFS, or localStorage)

### Recommended Features

- 🔷 IndexedDB (for optimal performance)
- 🔷 OPFS (Origin Private File System) for best file system performance
- 🔷 CompressionStream API (for efficient data compression)
- 🔷 Web Workers (for background operations)

## Feature Support Matrix

### Storage APIs

| Feature            | Chrome | Firefox | Safari | Mobile Chrome | Mobile Safari |
| ------------------ | ------ | ------- | ------ | ------------- | ------------- |
| IndexedDB          | ✅     | ✅      | ✅     | ✅            | ✅            |
| OPFS               | ✅ 86+ | ✅ 111+ | ❌     | ✅ 86+        | ❌            |
| localStorage       | ✅     | ✅      | ✅     | ✅            | ⚠️ Limited    |
| Storage Manager    | ✅     | ✅      | ✅     | ✅            | ⚠️ Partial    |
| Persistent Storage | ✅     | ✅      | ❌     | ✅            | ❌            |

**Recommended Storage Adapter by Browser:**

- Chrome/Edge: OPFS (best performance)
- Firefox 111+: OPFS
- Firefox <111: IndexedDB
- Safari: IndexedDB
- Mobile browsers: IndexedDB (for quota management)

### WebAssembly Features

| Feature    | Chrome | Firefox | Safari | Notes                      |
| ---------- | ------ | ------- | ------ | -------------------------- |
| Basic WASM | ✅     | ✅      | ✅     | Required                   |
| Streaming  | ✅     | ✅      | ✅     | Preferred for loading      |
| Threads    | ✅     | ✅      | ⚠️     | Requires SharedArrayBuffer |
| SIMD       | ✅     | ✅      | ✅     | Performance optimization   |

### Compression APIs

| Feature             | Chrome | Firefox | Safari   | Notes                            |
| ------------------- | ------ | ------- | -------- | -------------------------------- |
| CompressionStream   | ✅ 80+ | ✅ 113+ | ✅ 16.4+ | For packfile compression         |
| DecompressionStream | ✅ 80+ | ✅ 113+ | ✅ 16.4+ | For packfile decompression       |
| Fallback (pako)     | ✅     | ✅      | ✅       | Used when native API unavailable |

### Crypto APIs

| Feature      | Chrome | Firefox | Safari | Notes             |
| ------------ | ------ | ------- | ------ | ----------------- |
| Web Crypto   | ✅     | ✅      | ✅     | Required          |
| SubtleCrypto | ✅     | ✅      | ✅     | For SHA-1/SHA-256 |
| SHA-1        | ✅     | ✅      | ✅     | Git default hash  |
| SHA-256      | ✅     | ✅      | ✅     | For newer repos   |

## Performance Characteristics

### Chrome/Chromium-based Browsers

**Best Overall Performance**

- **OPFS**: Excellent file system performance
- **IndexedDB**: Fast, reliable
- **WASM**: Full optimization support
- **Memory**: Large heap size available

**Recommended Configuration:**

```typescript
const repo = await Repository.init("/repo", {
  storage: "opfs",
  hash: "sha1",
  compression: "native",
});
```

### Firefox

**Good Performance with Modern Versions**

- **OPFS** (v111+): Good performance
- **IndexedDB**: Excellent performance
- **WASM**: Full support with good optimization
- **Memory**: Moderate heap size

**Recommended Configuration:**

```typescript
const repo = await Repository.init("/repo", {
  storage: navigator.userAgent.includes("Firefox/1") ? "opfs" : "indexeddb",
  hash: "sha1",
  compression: "native",
});
```

### Safari (Desktop)

**Good Compatibility, Some Limitations**

- **No OPFS support**: Use IndexedDB
- **IndexedDB**: Good performance
- **WASM**: Full support
- **Storage quota**: More restrictive than Chrome

**Recommended Configuration:**

```typescript
const repo = await Repository.init("/repo", {
  storage: "indexeddb",
  hash: "sha1",
  compression: "native",
});
```

### Mobile Safari (iOS)

**Limited but Functional**

- **Storage quota**: Very restrictive (typically 50-100MB)
- **No OPFS**: Use IndexedDB
- **WASM**: Supported but may be slower
- **Memory**: Limited heap size

**Recommended Configuration:**

```typescript
const repo = await Repository.init("/repo", {
  storage: "indexeddb",
  hash: "sha1",
  compression: "native",
  maxCacheSize: 20 * 1024 * 1024, // 20MB cache limit
});
```

## Storage Quotas

### Typical Storage Limits

| Browser         | Typical Quota | Notes                             |
| --------------- | ------------- | --------------------------------- |
| Chrome Desktop  | ~60% of disk  | Very generous                     |
| Firefox Desktop | ~50% of disk  | Can request more                  |
| Safari Desktop  | ~1GB          | Requires user permission for more |
| Mobile Chrome   | ~6GB          | Device dependent                  |
| Mobile Safari   | 50-100MB      | Very restrictive                  |

### Checking Available Storage

```typescript
import { getStorageQuota } from "browser-git";

const quota = await getStorageQuota();
if (quota) {
  console.log(`Available: ${quota.available / 1024 / 1024}MB`);
  console.log(`Used: ${quota.percentage}%`);
}
```

### Requesting Persistent Storage

```typescript
import { requestPersistentStorage } from "browser-git";

const granted = await requestPersistentStorage();
if (granted) {
  console.log("Storage will persist across sessions");
}
```

## Performance Targets by Browser

### Chrome/Edge (Desktop)

- **Commit**: <20ms ✅
- **Checkout** (50 files): <100ms ✅
- **Clone** (100 commits): <3s ✅
- **Diff**: <30ms ✅

### Firefox (Desktop)

- **Commit**: <30ms ✅
- **Checkout** (50 files): <150ms ✅
- **Clone** (100 commits): <4s ✅
- **Diff**: <40ms ✅

### Safari (Desktop)

- **Commit**: <50ms ✅
- **Checkout** (50 files): <200ms ✅
- **Clone** (100 commits): <5s ✅
- **Diff**: <50ms ✅

### Mobile Browsers

- **Commit**: <100ms ⚠️
- **Checkout** (50 files): <300ms ⚠️
- **Clone** (100 commits): <8s ⚠️
- **Diff**: <100ms ⚠️

## Feature Detection

BrowserGit includes comprehensive automatic feature detection to use the best available APIs and provide detailed compatibility reports:

### Quick Compatibility Check

```typescript
import {
  checkCompatibility,
  logCompatibilityReport,
} from "@browser-git/browser-git/utils/browser-compat";

// Check compatibility and get detailed results
const result = await checkCompatibility();

if (!result.compatible) {
  console.error("Browser not compatible");
  console.error("Missing features:", result.missingFeatures);
  throw new Error(`Unsupported browser: ${result.missingFeatures.join(", ")}`);
}

// Show warnings to user
if (result.warnings.length > 0) {
  result.warnings.forEach((warning) => console.warn(warning));
}

// Show recommendations
if (result.recommendations.length > 0) {
  console.log("Recommendations:", result.recommendations);
}

// Or get a formatted compatibility report
await logCompatibilityReport();
```

### Detailed Capability Detection

```typescript
import {
  detectCapabilities,
  detectBrowser,
} from "@browser-git/browser-git/utils/browser-compat";

// Detect browser
const browser = detectBrowser();
console.log(`${browser.name} ${browser.version} (${browser.engine})`);

// Detect all capabilities
const capabilities = await detectCapabilities();
console.log("WebAssembly:", capabilities.webAssembly);
console.log("IndexedDB:", capabilities.indexedDB);
console.log("OPFS:", capabilities.opfs);
console.log("Web Crypto:", capabilities.webCrypto);
console.log("SHA-1:", capabilities.sha1Support);
console.log("SHA-256:", capabilities.sha256Support);
console.log("CompressionStream:", capabilities.compressionStream);

// Check storage quota
if (capabilities.storageQuota) {
  const { usage, quota, available, percentage } = capabilities.storageQuota;
  console.log(
    `Storage: ${(usage / 1024 / 1024).toFixed(2)}MB / ${(quota / 1024 / 1024 / 1024).toFixed(2)}GB (${percentage.toFixed(1)}% used)`,
  );
  console.log(`Available: ${(available / 1024 / 1024).toFixed(2)}MB`);
}
```

### Individual Feature Checks

```typescript
import {
  hasWebAssembly,
  hasIndexedDB,
  hasOPFS,
  hasLocalStorage,
  hasWebCrypto,
  hasHashAlgorithm,
  hasCompressionStream,
  getStorageQuota,
  hasPersistentStorage,
  requestPersistentStorage,
} from "@browser-git/browser-git/utils/browser-compat";

// Check individual features
const wasm = hasWebAssembly();
const idb = await hasIndexedDB();
const opfs = hasOPFS();
const crypto = hasWebCrypto();
const sha1 = await hasHashAlgorithm("SHA-1");
const compression = hasCompressionStream();

// Get storage information
const quota = await getStorageQuota();
const isPersistent = await hasPersistentStorage();

// Request persistent storage
const granted = await requestPersistentStorage();
if (granted) {
  console.log("Storage will not be automatically cleared");
}
```

### Compatibility Report Example

The `checkCompatibility()` function generates a comprehensive report:

```
=== Browser Compatibility Report ===

Browser: Chrome 120
Engine: Blink

✅ Compatible - All required features are supported

Core Features:
  ✅ WebAssembly
  ✅ Web Crypto API
  ✅ SHA-1
  ✅ SHA-256

Storage Features:
  ✅ IndexedDB
  ✅ OPFS
  ✅ localStorage
  ✅ sessionStorage

Storage Quota:
  Total: 234.56 GB
  Used: 15.32 MB (0.0%)
  Available: 234.54 GB

Recommendations:
  💡 Use OPFS storage adapter for best performance
```

## Automatic Adapter Selection

BrowserGit automatically selects the best storage adapter:

```typescript
import { getRecommendedStorageAdapter } from "browser-git";

const adapter = await getRecommendedStorageAdapter();
// Returns: 'opfs' | 'indexeddb' | 'localstorage' | 'memory'

console.log(`Using ${adapter} storage adapter`);
```

## Browser-Specific Optimizations

### Chrome Optimization Tips

1. Enable OPFS for best performance
2. Use `chrome://flags/#enable-experimental-web-platform-features` for latest features
3. Leverage Chrome DevTools Performance profiler

### Firefox Optimization Tips

1. Use IndexedDB for Firefox <111
2. Enable OPFS in Firefox 111+
3. Monitor memory usage in Developer Tools

### Safari Optimization Tips

1. Always use IndexedDB (no OPFS)
2. Implement aggressive cache cleanup
3. Monitor storage quota carefully
4. Request persistent storage early

### Mobile Optimization Tips

1. Implement progressive loading
2. Use smaller cache sizes
3. Clean up old objects regularly
4. Show storage usage to users

## Known Issues and Workarounds

### Safari Private Browsing

**Issue**: IndexedDB disabled in private browsing mode

**Workaround**: Use memory adapter

```typescript
const repo = await Repository.init("/repo", {
  storage: "memory",
});
```

### Mobile Safari Storage Limits

**Issue**: Very restrictive storage quotas (50-100MB)

**Workaround**: Implement shallow clone

```typescript
const repo = await Repository.clone(url, "/repo", {
  depth: 10, // Only last 10 commits
  storage: "indexeddb",
});
```

### Firefox OPFS Support

**Issue**: OPFS only available in Firefox 111+

**Workaround**: Feature detection

```typescript
import { checkOPFS } from "browser-git";

const hasOPFS = await checkOPFS();
const storage = hasOPFS ? "opfs" : "indexeddb";
```

## Testing Across Browsers

### Running Cross-Browser Tests

```bash
# Run tests on Chrome
npm run test:browser:chromium

# Run tests on Firefox
npm run test:browser:firefox

# Run tests on Safari/WebKit
npm run test:browser:webkit

# Run all browser tests
npm run test:browser
```

### Manual Testing Checklist

- [ ] Basic operations (init, add, commit)
- [ ] Storage adapter initialization
- [ ] Clone from remote repository
- [ ] Large file handling (>1MB)
- [ ] Storage quota monitoring
- [ ] WASM loading and execution
- [ ] Performance benchmarks

## Browser Support Policy

- **Evergreen browsers**: Always supported (Chrome, Firefox, Edge, Safari)
- **Version support**: Last 2 major versions
- **Mobile browsers**: iOS 15+ Safari, Android Chrome 90+
- **Deprecated**: IE11, Legacy Edge (<79)

## Getting Help

If you encounter browser-specific issues:

1. Check this compatibility guide
2. Review [GitHub Issues](https://github.com/nseba/browser-git/issues)
3. Run feature detection to identify missing capabilities
4. Check browser console for specific errors
5. File an issue with browser version and error details

## Future Enhancements

Planned improvements for browser compatibility:

- [ ] Better mobile browser optimizations
- [ ] Improved Safari storage quota handling
- [ ] Web Worker support for background operations
- [ ] ServiceWorker integration for offline support
- [ ] Progressive Web App (PWA) features

## Resources

- [Web Platform Status](https://chromestatus.com/)
- [Can I Use](https://caniuse.com/)
- [MDN Browser Compatibility](https://developer.mozilla.org/en-US/docs/Web/API#browser_compatibility)
- [WebAssembly Browser Support](https://webassembly.org/roadmap/)
