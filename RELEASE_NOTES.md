# Release Notes

## Version 0.1.0 (Initial Pre-Release) - 2025-11-18

### Overview

BrowserGit v0.1.0 is the initial pre-release of a pure browser-based Git implementation. This version provides core Git functionality that runs entirely in the browser using WebAssembly and modern browser APIs.

**⚠️ Important**: This is a pre-release version intended for evaluation and testing. It is **not recommended for production use** with sensitive data.

### Features

#### Core Git Operations
- ✅ **Repository Management**
  - Initialize new repositories (`git init`)
  - Open and manage existing repositories
  - Configure repository settings (user, email, hash algorithm)

- ✅ **File Operations**
  - Add files to staging area (`git add`)
  - Commit changes (`git commit`)
  - Check repository status (`git status`)
  - View commit history (`git log`)
  - Show commit details (`git show`)
  - Generate diffs (`git diff`)

- ✅ **Branching**
  - Create branches (`git branch`)
  - Delete branches
  - Rename branches
  - Switch branches (`git checkout`)
  - Track current branch

- ✅ **Merging**
  - Three-way merge
  - Fast-forward merge
  - Conflict detection
  - Structured conflict resolution API

- ✅ **Remote Operations**
  - Clone repositories (`git clone`)
  - Fetch from remotes (`git fetch`)
  - Pull changes (`git pull`)
  - Push commits (`git push`)
  - HTTP smart protocol support
  - Authentication (Basic, Token, OAuth)

#### Storage Options
- ✅ **Multiple Storage Backends**
  - IndexedDB (recommended for persistent storage)
  - OPFS (Origin Private File System, where available)
  - LocalStorage (limited capacity)
  - In-memory (for testing)

#### Diff Engine
- ✅ **Pluggable Diff System**
  - Myers diff algorithm
  - Line-by-line text diffing
  - Binary file detection
  - Unified diff format
  - Word-level diffs

#### Developer Tools
- ✅ **Command-Line Interface**
  - Full-featured CLI tool (`bgit`)
  - Supports all major git commands
  - Colored output
  - Progress reporting

- ✅ **Example Applications**
  - Basic demo application
  - Mini IDE with Git integration
  - Offline documentation system

### Architecture

#### Technology Stack
- **Go + TinyGo WASM**: Core Git implementation compiled to WebAssembly
- **TypeScript**: High-level API and browser integrations
- **Yarn Workspaces**: Monorepo structure with multiple packages

#### Packages
1. **git-core**: Core Git operations in Go/WASM
2. **browser-git**: TypeScript API wrapper
3. **storage-adapters**: Multiple storage backend implementations
4. **diff-engine**: Pluggable diff algorithm system
5. **git-cli**: Command-line interface

### Security Features

- ✅ Path normalization to prevent directory traversal
- ✅ URL validation with SSRF protection
- ✅ Authentication validation
- ✅ CORS error detection and handling
- ✅ No eval() or unsafe code execution
- ✅ Secure random number generation using Web Crypto API

### Known Limitations

#### Security Limitations
- ⚠️ **Credential Storage**: Tokens/passwords stored in browser storage are accessible to JavaScript (XSS risk)
- ⚠️ **No Encryption at Rest**: Git objects and credentials are not encrypted in storage
- ⚠️ **No GPG Signing**: Commit signing with GPG keys is not supported
- ⚠️ **No SSH**: SSH key-based authentication is not available in browsers

#### Functional Limitations
- ⚠️ **No Rebase**: Interactive rebase is not yet implemented
- ⚠️ **No Stash**: Git stash functionality is not available
- ⚠️ **No Submodules**: Git submodules are not supported
- ⚠️ **Limited Merge Strategies**: Only basic merge strategies are available
- ⚠️ **No LFS**: Large File Storage is not supported
- ⚠️ **No Hooks**: Git hooks cannot be executed in the browser

#### Performance Limitations
- ⚠️ **Large Repositories**: Performance degrades with very large repositories (>1000 commits)
- ⚠️ **Binary Files**: Large binary files may cause memory issues
- ⚠️ **Slow Cloning**: Clone operations are slower than native Git

#### Browser Limitations
- ⚠️ **Storage Quotas**: Browser storage is limited (typically 50-100GB, but can be less)
- ⚠️ **CORS Required**: Remote Git servers must send appropriate CORS headers
- ⚠️ **CSP Restrictions**: Requires `wasm-unsafe-eval` in Content Security Policy

### Browser Support

| Browser | Version | Status | Notes |
|---------|---------|--------|-------|
| Chrome | 90+ | ✅ Fully Supported | Best performance with OPFS |
| Edge | 90+ | ✅ Fully Supported | Best performance with OPFS |
| Firefox | 88+ | ✅ Supported | IndexedDB recommended |
| Safari | 15+ | ⚠️ Limited | Storage quotas more restrictive |
| Mobile Chrome | 90+ | ⚠️ Limited | Reduced storage capacity |
| Mobile Safari | 15+ | ⚠️ Limited | Very restrictive storage |

### Performance Benchmarks

Based on testing with a 100-commit repository:

| Operation | Target | Typical | Status |
|-----------|--------|---------|--------|
| Init | < 10ms | 5-8ms | ✅ Excellent |
| Add (small file) | < 20ms | 10-15ms | ✅ Excellent |
| Commit | < 50ms | 30-40ms | ✅ Good |
| Checkout | < 200ms | 120-180ms | ✅ Good |
| Clone (100 commits) | < 5s | 3-4s | ✅ Good |
| Merge (no conflict) | < 100ms | 60-80ms | ✅ Good |
| Diff (100 lines) | < 50ms | 20-30ms | ✅ Excellent |

### Installation

#### NPM Package
```bash
npm install @browser-git/browser-git @browser-git/storage-adapters
```

#### Yarn
```bash
yarn add @browser-git/browser-git @browser-git/storage-adapters
```

#### CLI Tool
```bash
npm install -g @browser-git/git-cli
```

### Quick Start

```typescript
import { Repository } from '@browser-git/browser-git';

// Initialize a new repository
const repo = await Repository.init('/my-repo', {
  storage: 'indexeddb',
  initialBranch: 'main',
});

// Configure user
await repo.config('user.name', 'Your Name');
await repo.config('user.email', 'your.email@example.com');

// Create and commit a file
await repo.fs.writeFile('README.md', '# My Project');
await repo.add(['README.md']);
await repo.commit('Initial commit', {
  author: {
    name: 'Your Name',
    email: 'your.email@example.com',
  },
});

// Clone a repository
const clonedRepo = await Repository.clone(
  'https://github.com/user/repo.git',
  '/local-path',
  {
    storage: 'indexeddb',
    auth: {
      method: 'token',
      token: 'your-github-token',
    },
  }
);
```

### Use Cases

#### ✅ Recommended Use Cases
- Browser-based code editors
- Documentation sites with version control
- Educational Git demonstrations
- Offline-first applications
- Git protocol prototyping
- Testing Git workflows

#### ⚠️ Not Recommended For
- Production applications with sensitive code
- Large-scale repositories (>10,000 commits)
- High-security environments
- Enterprise compliance requirements
- Primary version control solution

### Breaking Changes from Native Git

1. **Authentication**: SSH keys are not supported; use tokens or OAuth
2. **Hooks**: Git hooks cannot be executed in the browser
3. **Performance**: Slower than native Git, especially for large operations
4. **Storage**: Limited by browser storage quotas
5. **CORS**: Requires server-side CORS configuration

### Upgrade Path

This is the initial release, so there is no upgrade path from previous versions.

### Future Roadmap

Planned features for upcoming releases:

#### v0.2.0 (Q1 2025)
- Enhanced URL validation
- Improved performance for large repositories
- Better error messages
- More authentication options
- Documentation improvements

#### v0.3.0 (Q2 2025)
- Credential encryption using Web Crypto API
- Interactive rebase support
- Git stash functionality
- Cherry-pick support
- Better conflict resolution tools

#### v0.4.0 (Q3 2025)
- GPG commit signing (limited)
- Submodule support
- Git LFS integration
- Performance optimizations
- Advanced merge strategies

### Documentation

- [Getting Started Guide](docs/getting-started.md)
- [API Reference](docs/api-reference/)
- [Architecture Overview](docs/architecture/overview.md)
- [Security Policy](SECURITY.md)
- [Browser Compatibility](docs/browser-compatibility.md)
- [CORS Setup Guide](docs/guides/cors-workarounds.md)
- [Authentication Guide](docs/guides/authentication.md)

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to contribute to BrowserGit.

### Security

See [SECURITY.md](SECURITY.md) for security policy and how to report vulnerabilities.

### License

[MIT License](LICENSE)

### Acknowledgments

- TinyGo project for WebAssembly support
- Git project for the protocol specification
- All contributors and early testers

### Support

- GitHub Issues: https://github.com/nseba/browser-git/issues
- Discussions: https://github.com/nseba/browser-git/discussions
- Documentation: https://github.com/nseba/browser-git/tree/main/docs

---

**Note**: This is pre-release software. APIs may change in future versions. Use at your own risk for production applications.
