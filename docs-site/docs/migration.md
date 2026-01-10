---
sidebar_position: 7
---

# Migration Guide

This guide helps you migrate to BrowserGit from other Git libraries or upgrade between BrowserGit versions.

## Migrating from isomorphic-git

BrowserGit provides a similar API to isomorphic-git but with some key differences.

### Installation

```bash
# Remove isomorphic-git
npm uninstall isomorphic-git

# Install BrowserGit
npm install @browser-git/browser-git
```

### Import Changes

```typescript
// Before (isomorphic-git)
import git from "isomorphic-git";
import http from "isomorphic-git/http/web";
import FS from "@isomorphic-git/lightning-fs";

// After (BrowserGit)
import { Repository } from "@browser-git/browser-git";
```

### Repository Initialization

```typescript
// Before (isomorphic-git)
const fs = new FS("my-app");
await git.init({ fs, dir: "/project" });

// After (BrowserGit)
const repo = await Repository.init("/project", {
  storage: "indexeddb",
});
```

### Cloning

```typescript
// Before (isomorphic-git)
await git.clone({
  fs,
  http,
  dir: "/project",
  url: "https://github.com/user/repo",
  corsProxy: "https://cors.isomorphic-git.org",
  onAuth: () => ({ username: token, password: "x-oauth-basic" }),
});

// After (BrowserGit)
const repo = await Repository.clone(
  "https://github.com/user/repo",
  "/project",
  {
    corsProxy: "https://cors.isomorphic-git.org",
    auth: { type: "token", token },
  },
);
```

### Basic Operations

```typescript
// Before (isomorphic-git)
await git.add({ fs, dir: "/project", filepath: "file.txt" });
const sha = await git.commit({
  fs,
  dir: "/project",
  message: "Add file",
  author: { name: "John", email: "john@example.com" },
});

// After (BrowserGit)
await repo.add(["file.txt"]);
const commit = await repo.commit("Add file", {
  author: { name: "John", email: "john@example.com" },
});
```

### Status

```typescript
// Before (isomorphic-git)
const status = await git.statusMatrix({ fs, dir: "/project" });

// After (BrowserGit)
const status = await repo.status();
// Returns structured object with staged, modified, untracked arrays
```

### Reading Files

```typescript
// Before (isomorphic-git)
const content = await fs.promises.readFile("/project/file.txt", "utf8");

// After (BrowserGit)
const content = await repo.fs.readFile("/project/file.txt", "utf8");
```

### Key Differences

| Feature     | isomorphic-git           | BrowserGit                |
| ----------- | ------------------------ | ------------------------- |
| API Style   | Function-based           | Object-oriented           |
| Filesystem  | External (lightning-fs)  | Integrated                |
| Storage     | lightning-fs (IndexedDB) | Multiple backends         |
| WASM        | No                       | Yes (Go/TinyGo)           |
| TypeScript  | Types available          | Native TypeScript         |
| Diff Engine | Basic                    | Pluggable Myers algorithm |

## Migrating from js-git

js-git is an older library with a callback-based API.

### Installation

```bash
npm uninstall js-git
npm install @browser-git/browser-git
```

### API Changes

```typescript
// Before (js-git) - callback style
repo.loadAs("commit", hash, (err, commit) => {
  if (err) return handleError(err);
  processCommit(commit);
});

// After (BrowserGit) - async/await
const commit = await repo.getCommit(hash);
```

### Repository Creation

```typescript
// Before (js-git)
const repo = {};
require("js-git/mixins/mem-db")(repo);
require("js-git/mixins/create-tree")(repo);
require("js-git/mixins/pack-ops")(repo);
require("js-git/mixins/walkers")(repo);
require("js-git/mixins/read-combiner")(repo);
require("js-git/mixins/formats")(repo);

// After (BrowserGit)
const repo = await Repository.init("/project");
```

## Migrating from nodegit (Node.js)

If you're moving a Node.js application to the browser:

### Key Differences

1. **No native bindings**: BrowserGit uses WASM instead of native libgit2
2. **Async-first**: All operations are async in the browser
3. **Storage abstraction**: Use browser storage instead of filesystem

### Example Migration

```typescript
// Before (nodegit - Node.js)
const NodeGit = require("nodegit");
const repo = await NodeGit.Repository.open("/path/to/repo");
const commit = await repo.getHeadCommit();

// After (BrowserGit - Browser)
const repo = await Repository.open("/project");
const head = await repo.resolveRef("HEAD");
const commit = await repo.getCommit(head);
```

## Upgrading BrowserGit

### Version 0.x to 1.0

When upgrading to version 1.0 (future), the following changes may apply:

#### Breaking Changes

1. **Storage adapter interface changes**

```typescript
// Before (0.x)
adapter.put(key, value);

// After (1.0)
adapter.set(key, value);
```

2. **Repository options restructured**

```typescript
// Before (0.x)
const repo = await Repository.init("/project", "indexeddb");

// After (1.0)
const repo = await Repository.init("/project", {
  storage: "indexeddb",
});
```

#### Deprecation Warnings

Enable deprecation warnings to prepare for upgrades:

```typescript
import { setDeprecationHandler } from "@browser-git/browser-git";

setDeprecationHandler((message) => {
  console.warn("Deprecation:", message);
});
```

## Storage Migration

### Migrating Storage Backends

To migrate data from one storage backend to another:

```typescript
import {
  IndexedDBAdapter,
  OPFSAdapter,
  migrateStorage,
} from "@browser-git/storage-adapters";

// Migrate from IndexedDB to OPFS
const source = new IndexedDBAdapter("my-repo");
const target = new OPFSAdapter("my-repo");

await source.initialize();
await target.initialize();

await migrateStorage(source, target, {
  onProgress: (copied, total) => {
    console.log(`Migrating: ${copied}/${total}`);
  },
});

// Verify migration
const verified = await verifyMigration(source, target);
if (verified) {
  await source.clear(); // Clean up old storage
}
```

### Manual Migration

For custom migration logic:

```typescript
async function migrateRepository(oldAdapter, newAdapter) {
  await oldAdapter.initialize();
  await newAdapter.initialize();

  const keys = await oldAdapter.list();

  for (const key of keys) {
    const data = await oldAdapter.get(key);
    if (data) {
      await newAdapter.set(key, data);
    }
  }

  // Verify
  for (const key of keys) {
    const oldData = await oldAdapter.get(key);
    const newData = await newAdapter.get(key);
    if (!arraysEqual(oldData, newData)) {
      throw new Error(`Migration failed for key: ${key}`);
    }
  }
}
```

## Data Format Migration

### Repository Format Upgrades

BrowserGit automatically handles repository format upgrades:

```typescript
const repo = await Repository.open("/project");
// Automatic upgrade if needed

// Check format version
console.log("Format version:", repo.formatVersion);
```

### Manual Format Upgrade

```typescript
import { upgradeRepository } from "@browser-git/browser-git";

const result = await upgradeRepository("/project", {
  targetVersion: 2,
  backup: true,
});

if (result.upgraded) {
  console.log("Upgraded from", result.fromVersion, "to", result.toVersion);
}
```

## Troubleshooting Migration

### Common Issues

#### "Storage not found" after migration

The storage key may have changed. Check the adapter prefix:

```typescript
// Old format
const oldAdapter = new IndexedDBAdapter("repo-name");

// New format (if changed)
const newAdapter = new IndexedDBAdapter("browser-git-repo-name");
```

#### "Object not found" errors

Pack files may need to be unpacked:

```typescript
await repo.unpackPackfiles();
```

#### Type errors after upgrade

Update TypeScript types:

```bash
npm update @browser-git/browser-git
```

### Migration Checklist

- [ ] Backup existing data
- [ ] Update package dependencies
- [ ] Update import statements
- [ ] Update repository initialization code
- [ ] Update authentication handling
- [ ] Update progress callback signatures
- [ ] Test all Git operations
- [ ] Verify storage migration (if applicable)
- [ ] Update error handling for new error types

## Getting Help

If you encounter issues during migration:

1. Check the [GitHub Issues](https://github.com/user/browser-git/issues) for known issues
2. Review the [API Reference](./api/repository) for updated method signatures
3. Join the community Discord for real-time help

## See Also

- [Getting Started](./getting-started) - Fresh installation guide
- [API Reference](./api/repository) - Complete API documentation
- [Limitations](./limitations) - Known limitations to consider
