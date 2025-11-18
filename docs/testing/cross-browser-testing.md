# Cross-Browser Testing Guide

This document provides procedures and guidelines for testing BrowserGit across different browsers and platforms.

## Overview

BrowserGit requires thorough testing across multiple browsers to ensure compatibility and performance. This guide outlines our testing strategy, procedures, and tools.

## Test Infrastructure

### Automated Testing

BrowserGit uses **Playwright** for automated cross-browser testing:

- **Chromium**: Latest stable + 1 previous version
- **Firefox**: Latest stable + 1 previous version
- **WebKit**: Latest stable (Safari equivalent)

### Test Suites

Located in `/tests/browser/`:

1. **storage-adapters.spec.ts**: Storage API compatibility
2. **wasm-loading.spec.ts**: WebAssembly loading and initialization
3. **git-performance.spec.ts**: Performance benchmarks
4. **example.spec.ts**: End-to-end workflow tests

## Running Tests

### Prerequisites

```bash
# Install dependencies
yarn install

# Install Playwright browsers
npx playwright install
```

### Run All Tests

```bash
# Run on all browsers
yarn test:browser

# Run with UI
yarn test:browser:ui

# Run in headed mode (see browser window)
yarn test:browser:headed
```

### Run Specific Browser

```bash
# Chrome/Chromium only
yarn test:browser:chromium

# Firefox only
yarn test:browser:firefox

# Safari/WebKit only
yarn test:browser:webkit
```

### Debug Tests

```bash
# Debug mode with pause and inspector
yarn test:browser:debug

# Show test trace on failure
yarn test:browser --trace on
```

## Manual Testing Procedures

### Initial Setup Checklist

For each browser being tested:

1. **Environment Setup**
   - [ ] Clear all browser data (cache, cookies, storage)
   - [ ] Disable browser extensions
   - [ ] Enable JavaScript console
   - [ ] Open DevTools Performance tab
   - [ ] Note browser version and OS

2. **Feature Detection**
   ```typescript
   // Run in browser console
   import { logCompatibilityReport } from '@browser-git/browser-git/utils/browser-compat';
   await logCompatibilityReport();
   ```
   - [ ] Verify all required features present
   - [ ] Note any missing optional features
   - [ ] Check storage quota

### Core Functionality Tests

#### 1. Repository Initialization

```typescript
const repo = await Repository.init('/test-repo', {
  storage: 'indexeddb', // or 'opfs' if available
});

await repo.config('user.name', 'Test User');
await repo.config('user.email', 'test@example.com');
```

**Verify:**
- [ ] Repository initializes without errors
- [ ] Storage adapter loads correctly
- [ ] Configuration is saved

**Expected Time**: < 100ms

#### 2. File Operations

```typescript
// Create file
await repo.fs.writeFile('README.md', '# Test Repository');
await repo.fs.writeFile('src/index.ts', 'console.log("hello");');

// Stage files
await repo.add(['README.md', 'src/index.ts']);

// Commit
await repo.commit('Initial commit', {
  author: {
    name: 'Test User',
    email: 'test@example.com',
  },
});

// Check status
const status = await repo.status();
```

**Verify:**
- [ ] Files are created successfully
- [ ] Staging works correctly
- [ ] Commit completes without errors
- [ ] Status shows clean working directory

**Expected Time**: Commit < 50ms

#### 3. Branching

```typescript
// Create branch
await repo.createBranch('feature-branch');

// Switch to branch
await repo.checkout('feature-branch');

// Make changes
await repo.fs.writeFile('feature.ts', 'export const feature = true;');
await repo.add(['feature.ts']);
await repo.commit('Add feature');

// Switch back
await repo.checkout('main');

// List branches
const branches = await repo.listBranches();
```

**Verify:**
- [ ] Branch creation succeeds
- [ ] Checkout switches branches
- [ ] Working directory updates correctly
- [ ] Can switch back to main

**Expected Time**: Checkout < 200ms

#### 4. Merging

```typescript
// Merge feature branch
await repo.checkout('main');
const result = await repo.merge('feature-branch');

// Check for conflicts
if (result.conflicts.length > 0) {
  console.log('Conflicts:', result.conflicts);
}
```

**Verify:**
- [ ] Merge completes successfully
- [ ] Fast-forward merge when possible
- [ ] Conflicts detected when present
- [ ] Merge commit created

**Expected Time**: < 100ms (no conflicts)

#### 5. History

```typescript
// Get commit history
const log = await repo.log({ maxCount: 10 });
log.forEach(commit => {
  console.log(commit.hash, commit.message);
});

// Show commit details
const commit = await repo.showCommit(log[0].hash);
console.log(commit);

// Generate diff
const diff = await repo.diff('HEAD~1', 'HEAD');
console.log(diff);
```

**Verify:**
- [ ] Log shows all commits
- [ ] Commit details are accurate
- [ ] Diff shows correct changes

**Expected Time**: < 50ms per operation

### Performance Tests

#### 1. Commit Performance

```typescript
const times = [];
for (let i = 0; i < 10; i++) {
  await repo.fs.writeFile(`file${i}.txt`, `Content ${i}`);
  await repo.add([`file${i}.txt`]);

  const start = performance.now();
  await repo.commit(`Commit ${i}`);
  const duration = performance.now() - start;

  times.push(duration);
  console.log(`Commit ${i}: ${duration.toFixed(2)}ms`);
}

const avg = times.reduce((a, b) => a + b) / times.length;
console.log(`Average: ${avg.toFixed(2)}ms`);
```

**Targets:**
- Chrome: < 30ms avg
- Firefox: < 40ms avg
- Safari: < 50ms avg

#### 2. Checkout Performance

```typescript
// Create branch with 50 files
await repo.createBranch('perf-test');
await repo.checkout('perf-test');

for (let i = 0; i < 50; i++) {
  await repo.fs.writeFile(`perf${i}.txt`, 'x'.repeat(100));
}
await repo.add(Array.from({ length: 50 }, (_, i) => `perf${i}.txt`));
await repo.commit('Perf test commit');

// Measure checkout back to main
await repo.checkout('main');

const start = performance.now();
await repo.checkout('perf-test');
const duration = performance.now() - start;

console.log(`Checkout (50 files): ${duration.toFixed(2)}ms`);
```

**Targets:**
- Chrome: < 120ms
- Firefox: < 180ms
- Safari: < 200ms

#### 3. Memory Usage

```typescript
// Chrome only - check performance.memory
if (performance.memory) {
  const before = performance.memory.usedJSHeapSize;

  // Perform operations
  for (let i = 0; i < 100; i++) {
    await repo.fs.writeFile(`mem${i}.txt`, 'x'.repeat(1000));
    await repo.add([`mem${i}.txt`]);
    await repo.commit(`Commit ${i}`);
  }

  const after = performance.memory.usedJSHeapSize;
  const used = (after - before) / 1024 / 1024;

  console.log(`Memory used: ${used.toFixed(2)}MB`);
}
```

**Targets:**
- < 50MB for 100 commits with small files

### Storage Tests

#### 1. Storage Quota

```typescript
import { getStorageQuota } from '@browser-git/browser-git/utils/browser-compat';

// Check initial quota
const initial = await getStorageQuota();
console.log('Initial:', initial);

// Create large repository
for (let i = 0; i < 1000; i++) {
  await repo.fs.writeFile(`large${i}.txt`, 'x'.repeat(10000));
}
await repo.add(Array.from({ length: 1000 }, (_, i) => `large${i}.txt`));
await repo.commit('Large commit');

// Check after
const after = await getStorageQuota();
console.log('After:', after);
console.log('Used:', (after.usage - initial.usage) / 1024 / 1024, 'MB');
```

**Verify:**
- [ ] Storage quota is reported correctly
- [ ] Usage increases after operations
- [ ] No quota errors (unless intentional)

#### 2. Storage Persistence

```typescript
import { hasPersistentStorage, requestPersistentStorage } from '@browser-git/browser-git/utils/browser-compat';

// Check persistence
const isPersistent = await hasPersistentStorage();
console.log('Is persistent:', isPersistent);

// Request persistence
if (!isPersistent) {
  const granted = await requestPersistentStorage();
  console.log('Persistence granted:', granted);
}
```

**Verify:**
- [ ] Persistence status is accurate
- [ ] Request shows browser prompt (if supported)
- [ ] Storage persists after browser restart (manual check)

### Browser-Specific Tests

#### Chrome/Edge Tests

1. **OPFS Performance**
   ```typescript
   const repo = await Repository.init('/opfs-test', {
     storage: 'opfs',
   });

   // Run performance tests
   // Compare with IndexedDB
   ```

2. **Large Repository**
   - Test with 1000+ commits
   - Monitor DevTools Performance
   - Check memory usage in Task Manager

#### Firefox Tests

1. **OPFS Availability**
   ```typescript
   import { hasOPFS } from '@browser-git/browser-git/utils/browser-compat';
   const opfs = hasOPFS();
   console.log('OPFS available:', opfs);
   ```

2. **IndexedDB Performance**
   - Run same tests as Chrome
   - Compare performance metrics

#### Safari Tests

1. **Storage Limits**
   - Create repository approaching 1GB
   - Monitor quota warnings
   - Test behavior when quota exceeded

2. **Private Browsing**
   - Test in private browsing mode
   - Verify fallback to memory storage
   - Check for IndexedDB errors

### Mobile Testing

#### Mobile Chrome

1. **Reduced Storage**
   - Test with smaller repositories
   - Monitor storage usage closely
   - Test quota exceeded scenarios

2. **Background Throttling**
   - Start long operation
   - Switch to another app
   - Verify operation completes

#### Mobile Safari

1. **Strict Quotas**
   - Test with very small repositories
   - Implement aggressive cleanup
   - Monitor storage constantly

2. **iOS Restrictions**
   - Test in various iOS versions
   - Check for API limitations
   - Verify WASM performance

## Test Results Template

### Browser Test Report

```markdown
## Test Report

**Date**: YYYY-MM-DD
**Browser**: [Name] [Version]
**OS**: [Operating System]
**Tester**: [Name]

### Environment
- Browser Version:
- OS Version:
- Device: (if mobile)
- Storage Adapter Used:

### Feature Detection
- WebAssembly: ✅/❌
- IndexedDB: ✅/❌
- OPFS: ✅/❌
- Web Crypto: ✅/❌
- CompressionStream: ✅/❌

### Storage Quota
- Available: [X] GB
- Used: [X] MB
- Persistent: ✅/❌

### Core Functionality
- [ ] Repository Init
- [ ] File Operations
- [ ] Branching
- [ ] Merging
- [ ] History/Log

### Performance Results
- Commit (avg): [X]ms (target: <50ms)
- Checkout (50 files): [X]ms (target: <200ms)
- Memory Usage: [X]MB (target: <50MB)

### Issues Found
1. [Issue description]
   - Severity: High/Medium/Low
   - Reproduction: [Steps]
   - Error: [Error message]

### Notes
[Additional observations]

### Overall Status
✅ Pass / ⚠️ Pass with warnings / ❌ Fail
```

## Continuous Integration

### GitHub Actions

```yaml
# .github/workflows/browser-tests.yml
name: Browser Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: yarn install
      - run: npx playwright install
      - run: yarn test:browser
      - uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: test-results
          path: test-results/
```

## Regression Testing

### Before Each Release

1. **Full Browser Suite**
   - Run all automated tests on all browsers
   - Perform manual testing checklist
   - Document any new issues

2. **Performance Benchmarks**
   - Run performance tests on all browsers
   - Compare with previous release
   - Flag any regressions

3. **Storage Tests**
   - Test all storage adapters
   - Verify quota handling
   - Check persistence

## Troubleshooting

### Tests Won't Run

**Playwright not installed**:
```bash
npx playwright install
```

**Port conflicts**:
```bash
# Kill process on port
lsof -ti:3000 | xargs kill -9
```

### Tests Failing

**Timing issues**:
- Increase timeouts in test config
- Add explicit waits
- Check for race conditions

**Storage issues**:
- Clear browser data between runs
- Check storage quotas
- Verify IndexedDB/OPFS access

### Performance Issues

**Slower than expected**:
- Check CPU throttling in DevTools
- Verify no background processes
- Use production build, not dev

## Resources

- [Playwright Documentation](https://playwright.dev/)
- [WebDriver BiDi](https://w3c.github.io/webdriver-bidi/)
- [Browser Compatibility Data](https://github.com/mdn/browser-compat-data)
- [Web Platform Tests](https://web-platform-tests.org/)

---

**Last Updated**: 2025-11-18
