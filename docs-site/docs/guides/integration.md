---
sidebar_position: 1
---

# Integration Guide

This guide covers how to integrate BrowserGit into your web application, including common patterns for browser-based IDEs, documentation platforms, and offline-first applications.

## Basic Integration

### Installation

```bash
npm install @browser-git/browser-git
```

### Basic Setup

```typescript
import { Repository, FileSystem } from '@browser-git/browser-git';

// Initialize on page load
async function initGit() {
  const repo = await Repository.init('/workspace', {
    storage: 'indexeddb'
  });

  return repo;
}

// Use throughout your application
const repo = await initGit();
```

## React Integration

### Repository Context

Create a context to share the repository across components:

```tsx
import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { Repository } from '@browser-git/browser-git';

interface GitContextType {
  repo: Repository | null;
  loading: boolean;
  error: Error | null;
}

const GitContext = createContext<GitContextType>({
  repo: null,
  loading: true,
  error: null
});

export function GitProvider({ children, repoPath }: { children: ReactNode; repoPath: string }) {
  const [repo, setRepo] = useState<Repository | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    async function init() {
      try {
        const repository = await Repository.open(repoPath) ||
                          await Repository.init(repoPath, { storage: 'indexeddb' });
        setRepo(repository);
      } catch (e) {
        setError(e as Error);
      } finally {
        setLoading(false);
      }
    }

    init();
  }, [repoPath]);

  return (
    <GitContext.Provider value={{ repo, loading, error }}>
      {children}
    </GitContext.Provider>
  );
}

export function useGit() {
  return useContext(GitContext);
}
```

### Using the Hook

```tsx
function CommitButton() {
  const { repo } = useGit();
  const [committing, setCommitting] = useState(false);

  async function handleCommit() {
    if (!repo) return;

    setCommitting(true);
    try {
      await repo.add(['.']);
      await repo.commit('Update from browser', {
        author: { name: 'User', email: 'user@example.com' }
      });
    } finally {
      setCommitting(false);
    }
  }

  return (
    <button onClick={handleCommit} disabled={committing}>
      {committing ? 'Committing...' : 'Commit Changes'}
    </button>
  );
}
```

### File Tree Component

```tsx
interface FileNode {
  name: string;
  path: string;
  type: 'file' | 'directory';
  children?: FileNode[];
}

function FileTree({ path }: { path: string }) {
  const { repo } = useGit();
  const [tree, setTree] = useState<FileNode[]>([]);

  useEffect(() => {
    if (!repo) return;

    async function loadTree() {
      const entries = await repo.fs.readdir(path, { withFileTypes: true });
      const nodes = await Promise.all(
        entries.map(async (entry) => {
          const fullPath = `${path}/${entry.name}`;
          if (entry.isDirectory()) {
            return {
              name: entry.name,
              path: fullPath,
              type: 'directory' as const,
              children: [] // Load lazily
            };
          }
          return {
            name: entry.name,
            path: fullPath,
            type: 'file' as const
          };
        })
      );
      setTree(nodes);
    }

    loadTree();
  }, [repo, path]);

  return (
    <ul>
      {tree.map(node => (
        <li key={node.path}>
          {node.type === 'directory' ? '📁' : '📄'} {node.name}
        </li>
      ))}
    </ul>
  );
}
```

## Vue Integration

### Composable

```typescript
import { ref, onMounted } from 'vue';
import { Repository } from '@browser-git/browser-git';

export function useGitRepository(path: string) {
  const repo = ref<Repository | null>(null);
  const loading = ref(true);
  const error = ref<Error | null>(null);

  onMounted(async () => {
    try {
      repo.value = await Repository.open(path) ||
                   await Repository.init(path, { storage: 'indexeddb' });
    } catch (e) {
      error.value = e as Error;
    } finally {
      loading.value = false;
    }
  });

  async function commit(message: string, author: { name: string; email: string }) {
    if (!repo.value) throw new Error('Repository not initialized');

    await repo.value.add(['.']);
    return repo.value.commit(message, { author });
  }

  async function getStatus() {
    if (!repo.value) throw new Error('Repository not initialized');
    return repo.value.status();
  }

  return {
    repo,
    loading,
    error,
    commit,
    getStatus
  };
}
```

## Browser-Based IDE Integration

### Monaco Editor Integration

```typescript
import * as monaco from 'monaco-editor';
import { Repository } from '@browser-git/browser-git';

class GitIntegratedEditor {
  private repo: Repository;
  private editor: monaco.editor.IStandaloneCodeEditor;
  private currentFile: string | null = null;

  constructor(repo: Repository, container: HTMLElement) {
    this.repo = repo;
    this.editor = monaco.editor.create(container, {
      language: 'typescript',
      theme: 'vs-dark'
    });

    // Auto-save on change
    this.editor.onDidChangeModelContent(
      debounce(() => this.autoSave(), 1000)
    );
  }

  async openFile(path: string) {
    const content = await this.repo.fs.readFile(path, 'utf-8');
    const model = monaco.editor.createModel(
      content,
      undefined,
      monaco.Uri.file(path)
    );
    this.editor.setModel(model);
    this.currentFile = path;
  }

  private async autoSave() {
    if (!this.currentFile) return;

    const content = this.editor.getValue();
    await this.repo.fs.writeFile(this.currentFile, content);
  }

  async getFileDiff() {
    if (!this.currentFile) return null;

    const status = await this.repo.status();
    if (status.modified.includes(this.currentFile)) {
      return this.repo.diff({ paths: [this.currentFile] });
    }
    return null;
  }
}
```

### Diff Viewer

```typescript
import { DiffResult } from '@browser-git/browser-git';

function renderDiff(diff: DiffResult): string {
  let html = '';

  for (const file of diff.files) {
    html += `<div class="diff-file">
      <div class="diff-header">${file.path}</div>`;

    for (const hunk of file.hunks) {
      html += `<div class="diff-hunk">
        <div class="hunk-header">@@ -${hunk.oldStart},${hunk.oldLines} +${hunk.newStart},${hunk.newLines} @@</div>`;

      for (const change of hunk.changes) {
        const lineClass = change.type === 'add' ? 'line-add' :
                         change.type === 'delete' ? 'line-delete' : 'line-context';
        const prefix = change.type === 'add' ? '+' :
                      change.type === 'delete' ? '-' : ' ';

        html += `<div class="diff-line ${lineClass}">${prefix}${escapeHtml(change.content)}</div>`;
      }

      html += '</div>';
    }

    html += '</div>';
  }

  return html;
}
```

## Offline-First Applications

### Service Worker Integration

```typescript
// sw.js
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open('git-wasm-v1').then((cache) => {
      return cache.addAll([
        '/git-core.wasm',
        '/index.html',
        '/app.js'
      ]);
    })
  );
});

// Serve WASM from cache
self.addEventListener('fetch', (event) => {
  if (event.request.url.endsWith('.wasm')) {
    event.respondWith(
      caches.match(event.request).then((response) => {
        return response || fetch(event.request);
      })
    );
  }
});
```

### Sync Queue

```typescript
class GitSyncQueue {
  private queue: Array<{ type: string; data: unknown }> = [];
  private syncing = false;

  async enqueue(operation: { type: string; data: unknown }) {
    this.queue.push(operation);
    await this.persistQueue();

    if (navigator.onLine && !this.syncing) {
      this.processQueue();
    }
  }

  private async persistQueue() {
    localStorage.setItem('git-sync-queue', JSON.stringify(this.queue));
  }

  private async processQueue() {
    this.syncing = true;

    while (this.queue.length > 0 && navigator.onLine) {
      const operation = this.queue[0];

      try {
        await this.executeOperation(operation);
        this.queue.shift();
        await this.persistQueue();
      } catch (e) {
        console.error('Sync failed:', e);
        break;
      }
    }

    this.syncing = false;
  }

  private async executeOperation(op: { type: string; data: unknown }) {
    switch (op.type) {
      case 'push':
        await repo.push(op.data.remote, op.data.branch);
        break;
      case 'fetch':
        await repo.fetch(op.data.remote);
        break;
    }
  }
}
```

## Performance Optimization

### Lazy Loading

```typescript
// Only load WASM when needed
let repoPromise: Promise<Repository> | null = null;

export function getRepository(): Promise<Repository> {
  if (!repoPromise) {
    repoPromise = Repository.init('/workspace', { storage: 'indexeddb' });
  }
  return repoPromise;
}
```

### Virtual Scrolling for Large Histories

```tsx
import { useVirtualizer } from '@tanstack/react-virtual';

function CommitLog() {
  const { repo } = useGit();
  const [commits, setCommits] = useState<Commit[]>([]);
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: commits.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60, // Estimated row height
    overscan: 5
  });

  return (
    <div ref={parentRef} style={{ height: '400px', overflow: 'auto' }}>
      <div style={{ height: `${virtualizer.getTotalSize()}px`, position: 'relative' }}>
        {virtualizer.getVirtualItems().map((virtualRow) => (
          <div
            key={virtualRow.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualRow.size}px`,
              transform: `translateY(${virtualRow.start}px)`
            }}
          >
            <CommitRow commit={commits[virtualRow.index]} />
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Web Worker for Heavy Operations

```typescript
// git-worker.ts
import { Repository } from '@browser-git/browser-git';

let repo: Repository | null = null;

self.onmessage = async (event) => {
  const { type, payload, id } = event.data;

  try {
    let result;

    switch (type) {
      case 'init':
        repo = await Repository.init(payload.path, payload.options);
        result = { success: true };
        break;

      case 'clone':
        repo = await Repository.clone(payload.url, payload.path, payload.options);
        result = { success: true };
        break;

      case 'status':
        result = await repo?.status();
        break;

      // ... other operations
    }

    self.postMessage({ id, result });
  } catch (error) {
    self.postMessage({ id, error: error.message });
  }
};

// main.ts
class GitWorkerClient {
  private worker = new Worker('/git-worker.js', { type: 'module' });
  private pending = new Map<string, { resolve: Function; reject: Function }>();

  constructor() {
    this.worker.onmessage = (event) => {
      const { id, result, error } = event.data;
      const handler = this.pending.get(id);

      if (handler) {
        if (error) {
          handler.reject(new Error(error));
        } else {
          handler.resolve(result);
        }
        this.pending.delete(id);
      }
    };
  }

  private send(type: string, payload: unknown): Promise<unknown> {
    return new Promise((resolve, reject) => {
      const id = crypto.randomUUID();
      this.pending.set(id, { resolve, reject });
      this.worker.postMessage({ type, payload, id });
    });
  }

  init(path: string, options: unknown) {
    return this.send('init', { path, options });
  }

  clone(url: string, path: string, options: unknown) {
    return this.send('clone', { url, path, options });
  }
}
```

## Error Handling

### User-Friendly Error Messages

```typescript
import { GitError, StorageError, NetworkError } from '@browser-git/browser-git';

function handleGitError(error: Error): string {
  if (error instanceof NetworkError) {
    if (!navigator.onLine) {
      return 'You are offline. Changes will sync when you reconnect.';
    }
    return 'Network error. Please check your connection and try again.';
  }

  if (error instanceof StorageError) {
    if (error.code === 'QUOTA_EXCEEDED') {
      return 'Storage is full. Please free some space or delete unused repositories.';
    }
    return 'Storage error. Please try refreshing the page.';
  }

  if (error instanceof GitError) {
    switch (error.code) {
      case 'MERGE_CONFLICT':
        return 'Merge conflict detected. Please resolve conflicts before continuing.';
      case 'NOT_A_REPO':
        return 'This directory is not a Git repository.';
      case 'DIRTY_WORKING_TREE':
        return 'You have uncommitted changes. Please commit or stash them first.';
      default:
        return `Git error: ${error.message}`;
    }
  }

  return 'An unexpected error occurred. Please try again.';
}
```

## Next Steps

- [CORS Workarounds](./cors-workarounds) - Handle cross-origin requests
- [Authentication](./authentication) - Set up remote authentication
- [Browser Compatibility](../browser-compatibility) - Ensure cross-browser support
