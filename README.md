# BrowserGit

A full-featured Git implementation designed to run entirely in web browsers, built with Go + WebAssembly and TypeScript.

[![CI](https://github.com/nseba/browser-git/workflows/CI/badge.svg)](https://github.com/nseba/browser-git/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- ✅ **Complete Git Operations**: init, add, commit, branch, merge, checkout, clone, fetch, push, pull, log, diff, blame, status
- 🗄️ **Multiple Storage Backends**: IndexedDB (primary), OPFS, LocalStorage (fallback), in-memory
- 🔧 **Dual API Layers**: Low-level Node.js-like filesystem API + high-level Git operations API
- 🔐 **Flexible Authentication**: HTTP Basic Auth, OAuth hooks, SSH key support
- 🔮 **Future-Proof**: SHA-1 and SHA-256 hash algorithm support
- 🧩 **Pluggable Architecture**: Modular diff engine, storage adapters, and authentication handlers
- 🌐 **Cross-Browser**: Chrome, Firefox, Safari (latest 2 versions)
- ⚡ **High Performance**: Optimized for browser environments with WASM
- 📦 **Easy Integration**: Simple APIs for browser-based IDEs and applications
- 🧪 **Well Tested**: >80% code coverage with unit, integration, and cross-browser tests

## Quick Start

```bash
# Install dependencies
yarn install

# Build all packages
yarn build

# Run tests
yarn test

# Start development mode
yarn dev
```

## Installation

```bash
npm install @browser-git/core
# or
yarn add @browser-git/core
```

## Usage

```typescript
import { Repository } from '@browser-git/core';

// Initialize a new repository
const repo = await Repository.init('/my-project', {
  storage: 'indexeddb',
  hashAlgorithm: 'sha1'
});

// Add and commit files
await repo.add(['src/**/*.ts']);
await repo.commit('Initial commit', {
  author: { name: 'John Doe', email: 'john@example.com' }
});

// Create and switch branches
await repo.createBranch('feature/new-feature');
await repo.checkout('feature/new-feature');

// Clone a remote repository
const cloned = await Repository.clone(
  'https://github.com/user/repo.git',
  '/local-path',
  {
    storage: 'indexeddb',
    auth: { type: 'basic', token: 'ghp_...' }
  }
);

// Push changes
await repo.push('origin', 'main');
```

## Project Structure

This is a monorepo managed with Yarn workspaces:

```
browser-git/
├── packages/
│   ├── git-core/          # Go + WASM Git implementation
│   ├── browser-git/       # TypeScript wrapper & API layer
│   ├── storage-adapters/  # Storage backend implementations
│   ├── diff-engine/       # Pluggable diff algorithm
│   └── git-cli/           # CLI tool for testing and debugging
├── examples/
│   ├── basic-demo/        # Simple demo application
│   ├── mini-ide/          # Minimal IDE example
│   └── offline-docs/      # Documentation site with version control
├── tests/
│   ├── integration/       # Integration test suite
│   ├── browser/           # Cross-browser tests
│   └── performance/       # Performance benchmarks
└── docs/                  # Documentation site
```

## Development

### Prerequisites

- Node.js >= 18.0.0
- Yarn >= 4.0.0
- Go >= 1.21
- TinyGo >= 0.30.0 (for WASM compilation)

### Building

```bash
# Build WASM core
yarn build:wasm

# Build TypeScript packages
yarn build:packages

# Build everything
yarn build
```

### Testing

```bash
# Run all tests
yarn test

# Run unit tests only
yarn test:unit

# Run integration tests
yarn test:integration

# Run browser tests
yarn test:browser

# Run tests in watch mode
yarn test:watch

# Generate coverage report
yarn test:coverage
```

### Code Quality

```bash
# Lint code
yarn lint

# Fix linting issues
yarn lint:fix

# Type check
yarn typecheck

# Format code
yarn format

# Check formatting
yarn format:check
```

## Packages

### @browser-git/core
The main package that provides the high-level Git API for browser applications.

### @browser-git/storage
Storage adapter implementations for IndexedDB, OPFS, LocalStorage, and in-memory.

### @browser-git/diff
Pluggable diff engine with support for custom diff algorithms.

### @browser-git/cli
Command-line interface for testing and debugging Git operations.

## Examples

Check out the [examples](./examples) directory for complete working examples:

- **basic-demo**: Simple HTML/JS application demonstrating basic Git operations
- **mini-ide**: React-based minimal IDE with file editing and Git integration
- **offline-docs**: Documentation site with full version control

## Documentation

- [Getting Started Guide](./docs/getting-started.md)
- [API Reference](./docs/api-reference/)
- [Architecture Overview](./docs/architecture/overview.md)
- [Integration Guide](./docs/guides/integration.md)
- [Browser Compatibility](./docs/browser-compatibility.md)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## License

MIT © Sebastian Narvaez

## Roadmap

- [x] Phase 1: Foundation & Infrastructure
- [x] Phase 2: Core Git Operations
- [x] Phase 3: Diff & Merge
- [x] Phase 4: Remote Operations
- [ ] Phase 5: CLI, Examples & Documentation

See the [PRD](./tasks/0001-prd-browser-git.md) and [task list](./tasks/tasks-0001-prd-browser-git.md) for detailed implementation plans.

## Acknowledgments

- Built with [TinyGo](https://tinygo.org/) for efficient WASM compilation
- Inspired by [isomorphic-git](https://isomorphic-git.org/)
- Uses the Git protocol specification from [git-scm.com](https://git-scm.com/)

## Support

- 📖 [Documentation](./docs)
- 🐛 [Issue Tracker](https://github.com/nseba/browser-git/issues)
- 💬 [Discussions](https://github.com/nseba/browser-git/discussions)
