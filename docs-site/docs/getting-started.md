---
sidebar_position: 2
---

# Getting Started

This guide will help you install BrowserGit and create your first repository.

## Installation

Install the main package using npm or yarn:

```bash
npm install @browser-git/browser-git
# or
yarn add @browser-git/browser-git
```

If you need additional packages:

```bash
# Storage adapters (included in browser-git, but available separately)
npm install @browser-git/storage-adapters

# Diff engine (included in browser-git, but available separately)
npm install @browser-git/diff-engine
```

## Basic Setup

### Import the Library

```typescript
import { Repository, FileSystem } from '@browser-git/browser-git';
```

### Initialize a New Repository

```typescript
// Create a new repository with IndexedDB storage
const repo = await Repository.init('/my-project', {
  storage: 'indexeddb',
  hashAlgorithm: 'sha1'  // or 'sha256'
});

console.log('Repository initialized:', repo.path);
```

### Storage Options

BrowserGit supports multiple storage backends:

```typescript
// IndexedDB - recommended for most use cases
const repo1 = await Repository.init('/project1', { storage: 'indexeddb' });

// OPFS - best performance on supported browsers
const repo2 = await Repository.init('/project2', { storage: 'opfs' });

// LocalStorage - limited capacity, good for small repos
const repo3 = await Repository.init('/project3', { storage: 'localstorage' });

// Memory - ephemeral, good for testing
const repo4 = await Repository.init('/project4', { storage: 'memory' });
```

## Working with Files

### Using the FileSystem API

```typescript
// Access the filesystem
const fs = repo.fs;

// Write a file
await fs.writeFile('/my-project/src/index.ts', `
export function hello() {
  return 'Hello, World!';
}
`);

// Read a file
const content = await fs.readFile('/my-project/src/index.ts', 'utf-8');

// List directory contents
const files = await fs.readdir('/my-project/src');

// Check if file exists
const exists = await fs.exists('/my-project/src/index.ts');

// Create directories
await fs.mkdir('/my-project/src/utils', { recursive: true });
```

## Basic Git Operations

### Adding and Committing

```typescript
// Stage files
await repo.add(['src/index.ts']);

// Or stage all changes
await repo.add(['.']);

// Commit with author information
const commit = await repo.commit('Add initial source files', {
  author: {
    name: 'Your Name',
    email: 'you@example.com'
  }
});

console.log('Created commit:', commit.hash);
```

### Checking Status

```typescript
const status = await repo.status();

console.log('Staged files:', status.staged);
console.log('Modified files:', status.modified);
console.log('Untracked files:', status.untracked);
```

### Viewing History

```typescript
// Get commit log
const commits = await repo.log({ maxCount: 10 });

for (const commit of commits) {
  console.log(`${commit.hash.slice(0, 7)} - ${commit.message}`);
  console.log(`  Author: ${commit.author.name} <${commit.author.email}>`);
  console.log(`  Date: ${commit.date}`);
}
```

### Working with Branches

```typescript
// Create a new branch
await repo.createBranch('feature/new-feature');

// Switch to a branch
await repo.checkout('feature/new-feature');

// List all branches
const branches = await repo.listBranches();
console.log('Branches:', branches);

// Get current branch
const current = await repo.getCurrentBranch();
console.log('Current branch:', current);
```

### Viewing Diffs

```typescript
// Diff working directory against HEAD
const diff = await repo.diff();

for (const file of diff.files) {
  console.log(`File: ${file.path}`);
  console.log(`Status: ${file.status}`);
  console.log('Changes:', file.hunks);
}

// Diff between commits
const diffBetween = await repo.diff({
  from: 'abc1234',
  to: 'def5678'
});
```

## Cloning Remote Repositories

```typescript
// Clone a public repository
const cloned = await Repository.clone(
  'https://github.com/user/repo.git',
  '/local-path',
  {
    storage: 'indexeddb'
  }
);

// Clone with authentication
const privateRepo = await Repository.clone(
  'https://github.com/user/private-repo.git',
  '/private-local',
  {
    storage: 'indexeddb',
    auth: {
      type: 'token',
      token: 'ghp_your_personal_access_token'
    }
  }
);
```

## Pushing and Pulling

```typescript
// Add a remote
await repo.addRemote('origin', 'https://github.com/user/repo.git');

// Push to remote
await repo.push('origin', 'main', {
  auth: {
    type: 'token',
    token: 'ghp_your_token'
  }
});

// Fetch from remote
await repo.fetch('origin');

// Pull changes
await repo.pull('origin', 'main');
```

## Merging

```typescript
// Merge a branch into current branch
const result = await repo.merge('feature/new-feature');

if (result.conflicts.length > 0) {
  console.log('Merge conflicts in:', result.conflicts);
  // Handle conflicts...
} else {
  console.log('Merge successful!');
}
```

## Error Handling

```typescript
import { GitError, StorageError } from '@browser-git/browser-git';

try {
  await repo.checkout('non-existent-branch');
} catch (error) {
  if (error instanceof GitError) {
    console.error('Git error:', error.message);
    console.error('Code:', error.code);
  } else if (error instanceof StorageError) {
    console.error('Storage error:', error.message);
  } else {
    throw error;
  }
}
```

## Complete Example

Here's a complete example putting it all together:

```typescript
import { Repository } from '@browser-git/browser-git';

async function main() {
  // Initialize repository
  const repo = await Repository.init('/my-app', {
    storage: 'indexeddb'
  });

  // Create some files
  await repo.fs.writeFile('/my-app/README.md', '# My App\n\nA sample application.');
  await repo.fs.mkdir('/my-app/src');
  await repo.fs.writeFile('/my-app/src/index.ts', 'console.log("Hello!");');

  // Stage and commit
  await repo.add(['.']);
  await repo.commit('Initial commit', {
    author: { name: 'Developer', email: 'dev@example.com' }
  });

  // Create a feature branch
  await repo.createBranch('feature/add-utils');
  await repo.checkout('feature/add-utils');

  // Add more files
  await repo.fs.mkdir('/my-app/src/utils');
  await repo.fs.writeFile('/my-app/src/utils/helpers.ts', 'export const add = (a, b) => a + b;');

  await repo.add(['src/utils/helpers.ts']);
  await repo.commit('Add helper utilities', {
    author: { name: 'Developer', email: 'dev@example.com' }
  });

  // Switch back and merge
  await repo.checkout('main');
  await repo.merge('feature/add-utils');

  // View the log
  const log = await repo.log();
  console.log('Commit history:');
  for (const commit of log) {
    console.log(`  ${commit.hash.slice(0, 7)} ${commit.message}`);
  }
}

main().catch(console.error);
```

## Next Steps

- [Repository API Reference](./api/repository) - Complete API documentation
- [Storage Adapters](./api/storage-adapters) - Learn about storage options
- [CORS Workarounds](./guides/cors-workarounds) - Handle cross-origin requests
- [Authentication](./guides/authentication) - Set up authentication for remotes
