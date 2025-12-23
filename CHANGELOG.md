# Changelog

All notable changes to BrowserGit will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-18

### Initial Release

This is the first public release of BrowserGit, a complete Git implementation for browsers.

### Added

#### Core Git Operations

- Repository initialization and configuration
- File staging (`git add`) with glob patterns and .gitignore support
- Commit creation with author and committer metadata
- Branch creation, deletion, and renaming
- Checkout operations (branches and commits)
- Merge operations with three-way merge algorithm
- Conflict detection and resolution API
- Commit history and log operations
- Diff operations for file changes

#### Object Model

- Complete Git object model (blob, tree, commit, tag)
- SHA-1 and SHA-256 hash algorithm support
- Object serialization and deserialization
- Object compression using zlib
- Object database interface

#### Remote Operations

- Git HTTP smart protocol implementation
- Clone operation with progress tracking
- Fetch and pull operations
- Push operation with authentication
- Packfile format reading and writing
- Delta object encoding and decoding
- Pkt-line protocol support
- CORS detection and error handling

#### Storage Layer

- Pluggable storage adapter interface
- IndexedDB storage adapter (recommended)
- OPFS (Origin Private File System) adapter
- LocalStorage adapter
- In-memory adapter for testing
- Storage quota management utilities
- Data serialization helpers
- Browser feature detection

#### File System API

- Node.js-like async fs API
- Path manipulation utilities (join, dirname, basename, etc.)
- File reading and writing (text and binary)
- Directory operations (mkdir, readdir, rmdir)
- File watching API with event emitters

#### Diff Engine

- Pluggable diff interface
- Myers diff algorithm implementation
- Line-by-line diff support
- Word-level diff support
- Binary file diff detection
- Unified diff formatting

#### Authentication

- HTTP Basic Authentication
- OAuth flow integration hooks
- Custom authentication handlers
- Credential storage using browser APIs
- Token-based authentication (GitHub, GitLab)

#### CLI Tool

- Command-line interface (`bgit`)
- All major Git commands implemented
- Colored output and formatted tables
- Progress indicators
- Help documentation for all commands

#### Examples

- Basic demo application (HTML/JS)
- Mini IDE example (React)
- Offline documentation site example
- Integration examples for various use cases

#### Testing

- Comprehensive unit test suite
- Integration tests for complete workflows
- Cross-browser testing with Playwright
- Performance benchmarks
- Security and malformed input tests
- Mock adapters for testing

#### Documentation

- Architecture overview
- API reference documentation
- Getting started guide
- Integration guides
- CORS workarounds guide
- Authentication setup guide
- Browser compatibility matrix
- Performance optimization guide

#### Build System

- Monorepo structure with yarn workspaces
- TinyGo WASM compilation
- TypeScript builds for all packages
- Source maps for debugging
- Bundle size analysis
- CI/CD pipeline with GitHub Actions

### Security Features

- URL validation to prevent SSRF attacks
- Path sanitization to prevent directory traversal
- Input validation across all operations
- No arbitrary code execution (no eval)
- CSP compatibility
- Private IP address blocking
- Error message sanitization
- Secure credential handling

### Performance Optimizations

- Commit operations: < 50ms
- Checkout operations: < 200ms
- Clone operations: < 5s for 100-commit repos
- WASM bundle size: < 2MB gzipped
- Memory-efficient object storage
- Lazy loading of WASM modules

### Browser Compatibility

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Automatic feature detection
- Graceful degradation for older browsers

### Known Limitations

- Storage quotas vary by browser (typically 50MB-1GB)
- Large repositories (>500MB) may hit memory limits
- Remote operations require CORS-enabled servers
- OAuth flows require backend support
- Binary files may impact performance
- No support for Git LFS in this release
- No support for submodules in this release
- No support for rebase operations in this release

### Package Versions

All packages are released with version 0.1.0:

- `@browser-git/browser-git@0.1.0` - Main TypeScript API
- `@browser-git/storage-adapters@0.1.0` - Storage backend adapters
- `@browser-git/diff-engine@0.1.0` - Diff algorithm engine
- `@browser-git/git-cli@0.1.0` - Command-line interface

---

## [Unreleased]

### Planned for Future Releases

#### High Priority

- Interactive rebase operations
- Git stash functionality
- Shallow clone support
- Performance improvements for large repositories

#### Medium Priority

- Git submodule support
- Git LFS (Large File Storage) support
- Signed commits (GPG)
- Git hooks (client-side)
- Cherry-pick operations
- Advanced merge strategies

#### Low Priority

- Sparse checkout
- Git worktrees
- Git bisect
- Reflog support
- Bundle format support

### In Progress

- None

---

## Release Notes

### 0.1.0 Release Notes

**Release Date**: November 18, 2025

This initial release brings a complete, production-ready Git implementation to the browser. BrowserGit enables version control for browser-based applications without requiring server-side dependencies.

**Highlights:**

- Complete Git core functionality
- Multiple storage backends with IndexedDB as the recommended default
- Full remote operations support (clone, fetch, pull, push)
- Security-first design with input validation and SSRF protection
- Optimized for performance with sub-50ms commits
- Comprehensive test coverage
- Cross-browser compatibility (Chrome, Firefox, Safari)

**Getting Started:**

```bash
npm install @browser-git/browser-git
```

**Quick Example:**

```typescript
import { Repository } from "@browser-git/browser-git";

const repo = await Repository.init("/my-project");
await repo.writeFile("README.md", "# My Project");
await repo.add(["README.md"]);
await repo.commit("Initial commit");
```

**Documentation:**

- [Getting Started Guide](./docs/getting-started.md)
- [API Reference](./docs/api-reference/repository.md)
- [Architecture Overview](./docs/architecture/overview.md)

**Known Issues:**

- Large binary files (>10MB) may cause performance degradation
- Safari OPFS support requires macOS 13.4+ or iOS 16.4+
- Clone operations require CORS-enabled Git servers

**Upgrade Path:**
This is the initial release, so no upgrade from previous versions is necessary.

**Breaking Changes:**
None (initial release)

**Deprecations:**
None (initial release)

**Contributors:**
Special thanks to all contributors who made this release possible!

---

## Version Support

We follow semantic versioning (SemVer):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Support Policy

- **Current Release** (0.1.x): Full support with bug fixes and security updates
- **Previous Major**: Security updates only (when 1.0.0 is released)
- **Older Versions**: No official support

---

## Migration Guides

### Migrating to 0.1.0

This is the initial release, so no migration is necessary.

Future migration guides will be provided here for major version changes.

---

For detailed information about each release, see the [releases page](https://github.com/user/browser-git/releases).
