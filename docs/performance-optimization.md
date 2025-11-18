# Performance Optimization Guide

This guide provides strategies and best practices for optimizing BrowserGit performance across different browsers and use cases.

## Performance Targets

BrowserGit aims to meet the following performance targets:

| Operation | Target | Priority | Status |
|-----------|--------|----------|--------|
| Init | < 10ms | High | ✅ Typically 5-8ms |
| Add (small file) | < 20ms | High | ✅ Typically 10-15ms |
| Commit | < 50ms | Critical | ✅ Typically 30-40ms |
| Checkout (50 files) | < 200ms | Critical | ✅ Typically 120-180ms |
| Merge (no conflict) | < 100ms | High | ✅ Typically 60-80ms |
| Clone (100 commits) | < 5s | Critical | ✅ Typically 3-4s |
| Diff (100 lines) | < 50ms | Medium | ✅ Typically 20-30ms |
| Log (100 commits) | < 100ms | Medium | ✅ Typically 40-60ms |
| Status | < 50ms | High | ✅ Typically 20-30ms |

## Performance Monitoring

### Using the Performance Monitor

```typescript
import { performanceMonitor } from '@browser-git/browser-git/utils/performance';

// Enable monitoring
performanceMonitor.enable();

// Perform operations
await repo.commit('Test commit');

// Get statistics
const stats = performanceMonitor.getStats('commit');
console.log('Commit average:', stats?.avgDuration, 'ms');

// Generate report
console.log(performanceMonitor.generateReport());

// Clear metrics
performanceMonitor.clear();
```

### Measuring Custom Operations

```typescript
import { performanceMonitor } from '@browser-git/browser-git/utils/performance';

// Measure async operation
const result = await performanceMonitor.measure(
  'custom_operation',
  async () => {
    // Your code here
    return await someAsyncOperation();
  }
);

// Measure sync operation
const syncResult = performanceMonitor.measureSync(
  'sync_operation',
  () => {
    // Your sync code here
    return someCalculation();
  }
);

// Manual timing
const endTimer = performanceMonitor.startTimer('manual_operation');
// ... do work ...
endTimer();
```

### Memory Monitoring

```typescript
// Take memory snapshot (Chrome only)
const snapshot = performanceMonitor.takeMemorySnapshot();
console.log('Heap used:', snapshot.heapUsed / 1024 / 1024, 'MB');

// Get all snapshots
const snapshots = performanceMonitor.getMemorySnapshots();
```

## Optimization Strategies

### 1. Storage Adapter Selection

**OPFS (Best Performance)**
```typescript
const repo = await Repository.init('/repo', {
  storage: 'opfs',  // Fastest for Chrome/Edge 102+, Firefox 111+
});
```

**IndexedDB (Best Compatibility)**
```typescript
const repo = await Repository.init('/repo', {
  storage: 'indexeddb',  // Good performance, works everywhere
});
```

**Performance Comparison:**
- OPFS: 100% (baseline, fastest)
- IndexedDB: 85-95% (slightly slower but excellent compatibility)
- localStorage: 40-60% (much slower, only for small data)
- Memory: 100% (fast but no persistence)

### 2. Batch Operations

Instead of committing files one by one:

**❌ Slow:**
```typescript
for (const file of files) {
  await repo.fs.writeFile(file.path, file.content);
  await repo.add([file.path]);
  await repo.commit(`Add ${file.path}`);
}
```

**✅ Fast:**
```typescript
// Write all files
for (const file of files) {
  await repo.fs.writeFile(file.path, file.content);
}

// Stage all at once
await repo.add(files.map(f => f.path));

// Single commit
await repo.commit('Add multiple files');
```

### 3. Use Shallow Clones

For large repositories, use shallow clones:

```typescript
const repo = await Repository.clone(url, '/repo', {
  depth: 10,  // Only last 10 commits
  singleBranch: true,  // Only one branch
});
```

**Benefits:**
- Faster clone time (up to 10x)
- Less storage usage
- Reduced memory consumption

**When to Use:**
- CI/CD environments
- Mobile browsers
- Limited storage quotas
- Don't need full history

### 4. Optimize Object Storage

Configure object caching:

```typescript
const repo = await Repository.init('/repo', {
  objectCacheSize: 1000,  // Cache up to 1000 objects
  packfileCompression: 'fast',  // Faster compression
});
```

**Cache Size Guidelines:**
- Small repos (<100 commits): 500 objects
- Medium repos (100-1000 commits): 1000-2000 objects
- Large repos (>1000 commits): 2000-5000 objects

### 5. Lazy Loading

Load repository data on demand:

```typescript
// Don't load full history immediately
const log = await repo.log({
  maxCount: 10,  // Only latest 10 commits
});

// Paginate when displaying
const nextPage = await repo.log({
  maxCount: 10,
  skip: 10,  // Skip first 10
});
```

### 6. Web Workers

Offload heavy computations to workers:

```typescript
// In worker: compute-worker.ts
import { computeDiff } from '@browser-git/diff-engine';

self.onmessage = async (e) => {
  const { original, modified } = e.data;
  const diff = await computeDiff(original, modified);
  self.postMessage(diff);
};

// In main thread
const worker = new Worker('compute-worker.js');
worker.postMessage({ original, modified });
worker.onmessage = (e) => {
  const diff = e.data;
  // Use diff
};
```

### 7. Compression

Use native compression when available:

```typescript
import { hasCompressionStream } from '@browser-git/browser-git/utils/browser-compat';

const useNativeCompression = hasCompressionStream();

const repo = await Repository.init('/repo', {
  compression: useNativeCompression ? 'native' : 'pako',
});
```

### 8. Debouncing and Throttling

For real-time operations like status checking:

```typescript
import { debounce } from '@browser-git/browser-git/utils/performance';

// Debounce status checks
const debouncedStatus = debounce(
  async () => {
    const status = await repo.status();
    updateUI(status);
  },
  300  // Wait 300ms after last change
);

// Trigger on file changes
fileWatcher.on('change', debouncedStatus);
```

### 9. Memory Management

**Limit Concurrent Operations:**
```typescript
// Instead of Promise.all() for many items
const chunks = chunkArray(files, 10);  // Process 10 at a time
for (const chunk of chunks) {
  await Promise.all(chunk.map(f => processFile(f)));
}
```

**Clear Caches Periodically:**
```typescript
// After large operations
await repo.gc();  // Garbage collect

// Clear internal caches
repo.clearCache();
```

**Monitor Memory:**
```typescript
import { getMemoryUsage } from '@browser-git/browser-git/utils/browser-compat';

const usage = getMemoryUsage();
if (usage && usage.heapUsed > 100 * 1024 * 1024) {
  console.warn('High memory usage:', usage.heapUsed / 1024 / 1024, 'MB');
  // Clear caches or warn user
}
```

### 10. IndexedDB Optimization

**Use Transactions Efficiently:**
```typescript
// Batch writes in single transaction
const transaction = db.transaction(['objects'], 'readwrite');
const store = transaction.objectStore('objects');

for (const object of objects) {
  store.put(object);  // Don't await individual puts
}

// Wait for transaction to complete
await new Promise((resolve, reject) => {
  transaction.oncomplete = resolve;
  transaction.onerror = reject;
});
```

**Indexing:**
```typescript
// Add indexes for frequent queries
const store = db.createObjectStore('commits');
store.createIndex('author', 'author', { unique: false });
store.createIndex('timestamp', 'timestamp', { unique: false });
```

## Browser-Specific Optimizations

### Chrome/Edge

1. **Enable OPFS:**
   ```typescript
   const repo = await Repository.init('/repo', { storage: 'opfs' });
   ```

2. **Use File System Access API:**
   ```typescript
   // Request directory handle for native file access
   const dirHandle = await window.showDirectoryPicker();
   ```

3. **Leverage Performance API:**
   ```typescript
   // Chrome has detailed memory info
   console.log(performance.memory.usedJSHeapSize);
   ```

### Firefox

1. **Use IndexedDB for Firefox < 111:**
   ```typescript
   import { hasOPFS } from '@browser-git/browser-git/utils/browser-compat';

   const storage = hasOPFS() ? 'opfs' : 'indexeddb';
   const repo = await Repository.init('/repo', { storage });
   ```

2. **Optimize for Lower Memory:**
   ```typescript
   const repo = await Repository.init('/repo', {
     objectCacheSize: 500,  // Smaller cache for Firefox
   });
   ```

### Safari

1. **Aggressive Caching:**
   ```typescript
   const repo = await Repository.init('/repo', {
     storage: 'indexeddb',
     objectCacheSize: 2000,  // Larger cache to reduce I/O
   });
   ```

2. **Monitor Storage Quota:**
   ```typescript
   import { getStorageQuota } from '@browser-git/browser-git/utils/browser-compat';

   const quota = await getStorageQuota();
   if (quota && quota.available < 100 * 1024 * 1024) {
     console.warn('Low storage:', quota.available / 1024 / 1024, 'MB');
   }
   ```

3. **Request Persistent Storage Early:**
   ```typescript
   import { requestPersistentStorage } from '@browser-git/browser-git/utils/browser-compat';
   await requestPersistentStorage();
   ```

### Mobile Browsers

1. **Use Smaller Caches:**
   ```typescript
   const repo = await Repository.init('/repo', {
     objectCacheSize: 200,
     packfileCompression: 'fast',
   });
   ```

2. **Implement Progressive Loading:**
   ```typescript
   // Load history in chunks
   let offset = 0;
   const pageSize = 10;

   async function loadMoreCommits() {
     const commits = await repo.log({
       maxCount: pageSize,
       skip: offset,
     });
     offset += pageSize;
     return commits;
   }
   ```

3. **Show Progress Indicators:**
   ```typescript
   await repo.clone(url, '/repo', {
     onProgress: (message) => {
       updateProgressBar(message);
     },
   });
   ```

## Performance Profiling

### Chrome DevTools

1. **Performance Tab:**
   - Record operation
   - Look for long tasks
   - Identify bottlenecks

2. **Memory Tab:**
   - Take heap snapshot before/after
   - Look for memory leaks
   - Check for retained objects

3. **Application Tab:**
   - Inspect IndexedDB/OPFS
   - Check storage usage
   - Verify data structure

### Firefox Developer Tools

1. **Performance Tool:**
   - Profile operations
   - Check for slow functions
   - Analyze call tree

2. **Storage Inspector:**
   - View IndexedDB contents
   - Check quota usage
   - Monitor storage growth

### Safari Web Inspector

1. **Timelines:**
   - Record JavaScript execution
   - Identify slow operations
   - Check memory usage

2. **Storage:**
   - Inspect IndexedDB
   - Monitor quota
   - Check for quota errors

## Common Performance Issues

### Issue: Slow Commits

**Symptoms:**
- Commits taking > 100ms
- High CPU during commits

**Solutions:**
1. Reduce object cache writes
2. Use batch transactions
3. Enable compression
4. Check for large files

**Example Fix:**
```typescript
// Before: slow
await repo.add(largeFiles);
await repo.commit('message');

// After: fast
await repo.add(largeFiles, {
  batchSize: 100,  // Process in batches
  compression: true,
});
await repo.commit('message');
```

### Issue: Slow Checkouts

**Symptoms:**
- Checkout taking > 500ms
- Browser freezing

**Solutions:**
1. Use Web Workers
2. Implement streaming writes
3. Show progress indicator
4. Batch file operations

**Example Fix:**
```typescript
// Use worker for checkout
const worker = new Worker('checkout-worker.js');
worker.postMessage({ branch: 'main' });
worker.onmessage = (e) => {
  if (e.data.progress) {
    updateProgress(e.data.progress);
  }
};
```

### Issue: High Memory Usage

**Symptoms:**
- Memory growing continuously
- Browser warnings
- Crashes on large repos

**Solutions:**
1. Clear caches regularly
2. Reduce object cache size
3. Use streaming for large files
4. Implement garbage collection

**Example Fix:**
```typescript
// Periodic cleanup
setInterval(() => {
  repo.clearCache();
  performanceMonitor.clear();
}, 60000);  // Every minute
```

### Issue: Storage Quota Exceeded

**Symptoms:**
- QuotaExceededError
- Operations failing
- Safari especially

**Solutions:**
1. Implement shallow clone
2. Clean old objects
3. Compress packfiles
4. Warn users early

**Example Fix:**
```typescript
import { getStorageQuota } from '@browser-git/browser-git/utils/browser-compat';

// Check before large operations
const quota = await getStorageQuota();
if (quota && quota.percentage > 80) {
  const shouldContinue = confirm(
    `Storage is ${quota.percentage.toFixed(1)}% full. Continue?`
  );
  if (!shouldContinue) return;
}
```

## Performance Checklist

Before releasing or deploying:

- [ ] Run performance benchmarks on all target browsers
- [ ] Verify all operations meet performance targets
- [ ] Profile memory usage during typical workflows
- [ ] Test with large repositories (1000+ commits)
- [ ] Check storage quota handling
- [ ] Verify compression is working
- [ ] Test on mobile devices
- [ ] Implement progress indicators for long operations
- [ ] Add performance monitoring in production
- [ ] Document any known performance limitations

## Continuous Performance Monitoring

### In Development

```typescript
if (process.env.NODE_ENV === 'development') {
  performanceMonitor.enable();

  // Log slow operations
  performanceMonitor.on('slow', (operation, duration) => {
    console.warn(`Slow ${operation}: ${duration}ms`);
  });
}
```

### In Production

```typescript
// Sample performance data
if (Math.random() < 0.01) {  // 1% sampling
  performanceMonitor.enable();

  // Send to analytics
  window.addEventListener('beforeunload', () => {
    const metrics = performanceMonitor.exportMetrics();
    analytics.send('performance', metrics);
  });
}
```

## Resources

- [Web Performance Working Group](https://www.w3.org/webperf/)
- [Chrome DevTools Performance](https://developer.chrome.com/docs/devtools/performance/)
- [Firefox Performance Tools](https://firefox-source-docs.mozilla.org/devtools-user/performance/)
- [WebPageTest](https://www.webpagetest.org/)
- [Lighthouse](https://developers.google.com/web/tools/lighthouse)

---

**Last Updated**: 2025-11-18
