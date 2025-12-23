# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

browser-git is a full-featured Git implementation for browsers using Go + WebAssembly and TypeScript. It enables Git operations directly in web browsers with multiple storage backends.

## Commands

### Development

```bash
npm install                      # Install dependencies
npm run build                    # Build all packages
npm run build:wasm               # Build WASM module (requires TinyGo)
npm run dev                      # Watch mode for all packages
```

### Testing

```bash
npm run test                     # Run unit tests (all packages)
npm run test:watch               # Watch mode for unit tests
npm run test:browser             # Run Playwright browser tests
npm run test:browser:chromium    # Run browser tests in Chromium only
npm run test:browser:headed      # Run browser tests with visible browser
npm run test:browser:debug       # Run browser tests in debug mode
npm run test:coverage            # Run tests with coverage

# Run tests for a specific package
cd packages/storage-adapters && npm test
cd packages/diff-engine && npm test
cd packages/browser-git && npm test
```

### Linting & Type Checking

```bash
npm run lint                     # ESLint check
npm run lint:fix                 # ESLint auto-fix
npm run type-check               # TypeScript type check (all packages)
npm run format:check             # Prettier check
npm run format                   # Prettier fix
```

### Benchmarks

```bash
npm run bench                    # Run all benchmarks
npm run bench:run                # Run benchmarks once (no watch)
npm run bench -- benchmarks/storage-adapters.bench.ts  # Specific benchmark
```

### CI

```bash
npm run ci                       # lint + type-check + test:unit + build
npm run ci:full                  # Full CI including browser tests and benchmarks
```

## Claude Code Skills

This project includes custom skills for structured feature development:

- `/create-prd` - Generate Product Requirements Documents. Use when planning a new feature. Asks clarifying questions and creates PRD in `/tasks/` directory.
- `/generate-tasks` - Convert PRD into executable task list. Creates parent tasks with sub-tasks and quality gates. Saves to `/tasks/tasks-[prd-name].md`.
- `/execute-tasks` - Execute task lists automatically. Runs all sub-tasks within a parent task without pausing, commits after each parent task, pauses for user confirmation before next parent.

**Workflow:** `/create-prd` → `/generate-tasks` → `/execute-tasks`

## Architecture

### Package Structure

The monorepo uses npm workspaces with 4 main packages:

**@browser-git/storage-adapters** (`packages/storage-adapters/`)

- Storage backend abstraction layer
- Implementations: `IndexedDBAdapter`, `OPFSAdapter`, `LocalStorageAdapter`, `MemoryAdapter`, `MockAdapter`
- All adapters implement `StorageAdapter` interface (get/set/delete/list/exists/clear)
- Uses jsdom for testing browser APIs

**@browser-git/diff-engine** (`packages/diff-engine/`)

- Pluggable diff engine with Myers algorithm
- `MyersDiffEngine` for text diffs, binary diff utilities
- `DiffEngineFactory` for creating diff engine instances
- Exports types: `DiffResult`, `DiffHunk`, `Change`, `DiffOptions`

**@browser-git/browser-git** (`packages/browser-git/`)

- Main library - high-level Git repository API
- `Repository` class: clone, init, open, add, commit, status, log, diff, merge, checkout, fetch, pull, push
- `FileSystem` class: browser-compatible filesystem API
- Browser feature detection and performance monitoring utilities
- Re-exports diff-engine for convenience
- Depends on storage-adapters and diff-engine

**@browser-git/git-cli** (`packages/git-cli/`)

- CLI wrapper around browser-git
- Commands: init, add, commit, status, log, diff, branch, checkout, merge, clone, fetch, pull, push
- Binaries: `browser-git`, `bgit`

**git-core** (`packages/git-core/`)

- Go/TinyGo WASM module for core Git operations
- Build with `make build` (production) or `make build-dev` (development)
- Outputs to `packages/browser-git/dist/git-core.wasm`

### Key Dependencies Between Packages

```
git-cli → browser-git → storage-adapters
                      → diff-engine
browser-git ←(loads)← git-core.wasm
```

### Test Organization

- Unit tests: `packages/*/test/*.test.ts` (vitest, jsdom environment)
- Browser tests: `tests/browser/*.spec.ts` (Playwright)
- Security tests: `tests/security/*.test.ts`
- Benchmarks: `benchmarks/*.bench.ts`

### TypeScript Configuration

- Base config: `tsconfig.base.json` (ES2020, ESNext modules, strict mode)
- Each package has its own `tsconfig.json` extending base
- Incremental builds with composite projects

### Storage Adapter Selection

- IndexedDB: Primary storage, recommended for most use cases
- OPFS: Best performance on modern browsers (Chrome 86+, Firefox 111+, Safari 15.2+)
- LocalStorage: Limited capacity fallback
- Memory: Testing only (ephemeral)
