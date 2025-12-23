# Introducing BrowserGit: A Pure Browser-Based Git Implementation

We're excited to announce the release of **BrowserGit** - a complete Git implementation designed to run entirely in the browser, enabling version control for browser-based applications without any server-side dependencies.

## What is BrowserGit?

BrowserGit is a modern, performance-focused Git implementation built with WebAssembly (Go/TinyGo) and TypeScript. It brings the full power of Git to browser-based applications, enabling:

- **Offline-first version control** for web applications
- **Browser-based IDEs** with full Git integration
- **Documentation sites** with built-in version history
- **Collaborative editing** with local version control
- **Educational platforms** for learning Git interactively

## Key Features

### Complete Git Implementation

BrowserGit implements the core Git functionality you need:

- **Repository Management**: Initialize, configure, and manage Git repositories
- **Staging & Commits**: Add files, create commits, and track changes
- **Branching**: Create, switch, merge, and delete branches
- **History**: View commit history, diffs, and file changes
- **Remote Operations**: Clone, fetch, pull, and push to remote repositories
- **Merge & Conflict Resolution**: Three-way merge with structured conflict data

### Multiple Storage Backends

Choose the storage backend that fits your needs:

- **IndexedDB**: High-performance, large storage capacity (recommended)
- **OPFS (Origin Private File System)**: Native file system performance
- **LocalStorage**: Simple key-value storage for small repositories
- **In-Memory**: Fast, ephemeral storage for testing

### Browser Compatibility

Tested and optimized for:

- ✅ Chrome/Edge (v90+)
- ✅ Firefox (v88+)
- ✅ Safari (v14+)

With graceful degradation for older browsers and automatic feature detection.

### Performance Optimized

- **Fast commit operations**: < 50ms for typical commits
- **Quick checkouts**: < 200ms for branch switching
- **Efficient cloning**: < 5s for 100-commit repositories
- **Small bundle size**: < 2MB gzipped WASM

### Security First

Built with security in mind:

- **URL validation** to prevent SSRF attacks
- **Path sanitization** to prevent directory traversal
- **Input validation** across all operations
- **CSP compatible** for WASM execution
- **No arbitrary code execution** (no eval, no Function constructor)

## Getting Started

### Installation

```bash
npm install @browser-git/browser-git
```

### Basic Usage

```typescript
import { Repository } from "@browser-git/browser-git";

// Initialize a new repository
const repo = await Repository.init("/my-project", {
  storage: "indexeddb",
  author: {
    name: "Your Name",
    email: "you@example.com",
  },
});

// Create a file and commit
await repo.writeFile("README.md", "# My Project");
await repo.add(["README.md"]);
await repo.commit("Initial commit");

// View history
const log = await repo.log();
console.log(log);
```

### Clone a Repository

```typescript
import { Repository } from "@browser-git/browser-git";

// Clone from GitHub
const repo = await Repository.clone(
  "https://github.com/user/repo.git",
  "/local-path",
  {
    storage: "indexeddb",
    auth: {
      username: "user",
      token: "ghp_xxx",
    },
  },
);
```

### Working with Branches

```typescript
// Create and switch to a new branch
await repo.createBranch("feature/awesome");
await repo.checkout("feature/awesome");

// Make changes
await repo.writeFile("feature.js", 'console.log("awesome");');
await repo.add(["feature.js"]);
await repo.commit("Add awesome feature");

// Merge back to main
await repo.checkout("main");
await repo.merge("feature/awesome");
```

## Use Cases

### 1. Browser-Based IDEs

Build powerful web-based development environments with full Git integration:

```typescript
import { Repository } from "@browser-git/browser-git";
import { MonacoEditor } from "monaco-editor";

// Integrate Git with your editor
const repo = await Repository.init("/workspace");
const editor = new MonacoEditor();

// Save and commit on change
editor.onDidChangeContent(async () => {
  const content = editor.getValue();
  await repo.writeFile("index.js", content);
  await repo.add(["index.js"]);
  await repo.commit("Auto-save");
});
```

### 2. Documentation Sites with Version Control

Enable users to track changes to their documentation:

```typescript
const docsRepo = await Repository.init("/my-docs", {
  storage: "indexeddb",
});

// User edits a document
await docsRepo.writeFile("getting-started.md", newContent);
await docsRepo.add(["getting-started.md"]);
await docsRepo.commit("Update getting started guide");

// View change history
const history = await docsRepo.log({ path: "getting-started.md" });
const diff = await docsRepo.diff("HEAD~1", "HEAD", "getting-started.md");
```

### 3. Offline-First Applications

Build applications that work offline and sync when online:

```typescript
// Work offline
const repo = await Repository.init("/offline-work");
await repo.writeFile("notes.txt", "My offline notes");
await repo.add(["notes.txt"]);
await repo.commit("Add notes");

// Sync when online
if (navigator.onLine) {
  await repo.push("origin", "main");
}
```

## Architecture

BrowserGit is built on a modern, modular architecture:

```
┌─────────────────────────────────────────┐
│         Browser Application             │
└─────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────┐
│     TypeScript API (browser-git)        │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ Repository  │  │  File System    │   │
│  │     API     │  │      API        │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
┌──────────────┐      ┌──────────────────┐
│  WASM Core   │      │ Storage Adapters │
│  (Go/TinyGo) │      │  ┌────────────┐  │
│  ┌─────────┐ │      │  │ IndexedDB  │  │
│  │Objects  │ │      │  ├────────────┤  │
│  │Protocol │ │      │  │    OPFS    │  │
│  │Merge    │ │      │  ├────────────┤  │
│  └─────────┘ │      │  │LocalStorage│  │
└──────────────┘      └──────────────────┘
```

### Components

- **git-core** (Go/WASM): Core Git operations, object model, and protocol
- **browser-git** (TypeScript): High-level API and WASM bridge
- **storage-adapters**: Pluggable storage backends
- **diff-engine**: Pluggable diff algorithms
- **git-cli**: Command-line interface (Node.js)

## Examples

Check out our example applications:

### Basic Demo

A simple HTML/JS demonstration of basic Git operations.

```bash
cd examples/basic-demo
npm install
npm run dev
```

### Mini IDE

A React-based mini IDE with file tree, editor, and Git panel.

```bash
cd examples/mini-ide
npm install
npm run dev
```

### Offline Docs

A documentation site with version control for content.

```bash
cd examples/offline-docs
npm install
npm run dev
```

## Performance Benchmarks

Tested on Chrome 120, macOS M1:

| Operation             | Time   | Notes                  |
| --------------------- | ------ | ---------------------- |
| Initialize repository | ~10ms  | Create .git structure  |
| Stage file (1KB)      | ~5ms   | Add to index           |
| Commit                | ~30ms  | Create commit object   |
| Checkout branch       | ~150ms | 100 files              |
| Merge (fast-forward)  | ~20ms  | Update references      |
| Merge (3-way)         | ~200ms | 50 files, no conflicts |
| Clone (100 commits)   | ~3s    | HTTP, IndexedDB        |
| Diff (1000 lines)     | ~50ms  | Myers algorithm        |

## Limitations

While BrowserGit is feature-complete for most use cases, be aware of:

1. **Storage Quotas**: Browser storage is limited (typically 50MB-1GB)
2. **Large Repositories**: Very large repos (>500MB) may hit memory limits
3. **CORS Restrictions**: Remote operations require CORS-enabled servers
4. **Binary Files**: Large binary files may impact performance
5. **Authentication**: OAuth flows require backend support

See our [documentation](./docs/README.md) for workarounds and best practices.

## API Documentation

Full API documentation is available at:

- [Repository API](./docs/api-reference/repository.md)
- [File System API](./docs/api-reference/filesystem.md)
- [Storage Adapters](./docs/api-reference/storage.md)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/user/browser-git.git
cd browser-git

# Install dependencies
npm install

# Build all packages
npm run build

# Run tests
npm test

# Run examples
cd examples/basic-demo
npm run dev
```

## Roadmap

Planned features for future releases:

- [ ] **Rebasing**: Interactive and automatic rebase
- [ ] **Stashing**: Temporary work storage
- [ ] **Submodules**: Nested repository support
- [ ] **LFS Support**: Large file storage
- [ ] **Signed Commits**: GPG signature support
- [ ] **Shallow Clones**: Partial repository cloning
- [ ] **Sparse Checkout**: Selective file checkout
- [ ] **Git Hooks**: Client-side hooks

## Community

- **GitHub**: [github.com/user/browser-git](https://github.com/user/browser-git)
- **Issues**: [Report bugs or request features](https://github.com/user/browser-git/issues)
- **Discussions**: [Join the conversation](https://github.com/user/browser-git/discussions)

## License

BrowserGit is released under the MIT License. See [LICENSE](./LICENSE) for details.

## Acknowledgments

BrowserGit builds on the excellent work of:

- **Git**: The brilliant version control system by Linus Torvalds
- **TinyGo**: Bringing Go to WebAssembly
- **isomorphic-git**: Inspiration for browser-based Git
- **diff-match-patch**: Robust diffing algorithms

## Get Started Today

```bash
npm install @browser-git/browser-git
```

Ready to bring Git to your browser application? [Get started with our guide](./docs/getting-started.md) or [try our live demo](https://browser-git-demo.example.com).

---

**Questions?** Check out our [FAQ](./docs/FAQ.md) or [open an issue](https://github.com/user/browser-git/issues/new).

**Built with ❤️ by the BrowserGit team**
