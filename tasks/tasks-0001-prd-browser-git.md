# Implementation Tasks: BrowserGit

Based on PRD: `0001-prd-browser-git.md`

**Status:** Complete task breakdown with sub-tasks
**Last Updated:** 2025-11-12

## Relevant Files

### Root Configuration
- `package.json` - Root workspace configuration for yarn workspaces
- `.gitignore` - Git ignore patterns
- `tsconfig.base.json` - Base TypeScript configuration shared across packages
- `README.md` - Project overview and getting started guide
- `.github/workflows/ci.yml` - GitHub Actions CI/CD pipeline
- `.github/workflows/release.yml` - Release automation workflow

### Package: git-core (Go + WASM)
- `packages/git-core/main.go` - Main WASM entry point and JS interop
- `packages/git-core/pkg/hash/interface.go` - Hash algorithm interface
- `packages/git-core/pkg/hash/sha1.go` - SHA-1 implementation
- `packages/git-core/pkg/hash/sha256.go` - SHA-256 implementation
- `packages/git-core/pkg/hash/hash_test.go` - Hash algorithm tests
- `packages/git-core/pkg/objects/blob.go` - Blob object implementation
- `packages/git-core/pkg/objects/tree.go` - Tree object implementation
- `packages/git-core/pkg/objects/commit.go` - Commit object implementation
- `packages/git-core/pkg/objects/tag.go` - Tag object implementation
- `packages/git-core/pkg/objects/objects_test.go` - Object model tests
- `packages/git-core/pkg/repository/repository.go` - Repository structure and initialization
- `packages/git-core/pkg/repository/repository_test.go` - Repository tests
- `packages/git-core/pkg/index/index.go` - Staging area (index) implementation
- `packages/git-core/pkg/index/index_test.go` - Index tests
- `packages/git-core/pkg/refs/refs.go` - Reference management (branches, tags)
- `packages/git-core/pkg/refs/refs_test.go` - Reference tests
- `packages/git-core/pkg/merge/merge.go` - Merge implementation
- `packages/git-core/pkg/merge/merge_test.go` - Merge tests
- `packages/git-core/pkg/protocol/http.go` - HTTP Git protocol implementation
- `packages/git-core/pkg/protocol/packfile.go` - Packfile reading/writing
- `packages/git-core/pkg/protocol/protocol_test.go` - Protocol tests
- `packages/git-core/Makefile` - Build automation for WASM compilation
- `packages/git-core/go.mod` - Go module definition
- `packages/git-core/README.md` - Package documentation

### Package: browser-git (TypeScript API wrapper)
- `packages/browser-git/src/index.ts` - Main entry point and API exports
- `packages/browser-git/src/repository.ts` - High-level repository API
- `packages/browser-git/src/wasm-loader.ts` - WASM binary loader and initialization
- `packages/browser-git/src/wasm-bridge.ts` - JS-WASM communication bridge
- `packages/browser-git/src/filesystem/index.ts` - File system API exports
- `packages/browser-git/src/filesystem/fs.ts` - Node.js-like fs API implementation
- `packages/browser-git/src/filesystem/path.ts` - Path manipulation utilities
- `packages/browser-git/src/filesystem/watch.ts` - File watching/notification API
- `packages/browser-git/src/types/index.ts` - TypeScript type definitions
- `packages/browser-git/src/types/git-objects.ts` - Git object type definitions
- `packages/browser-git/src/types/repository.ts` - Repository type definitions
- `packages/browser-git/src/types/storage.ts` - Storage interface type definitions
- `packages/browser-git/src/errors/index.ts` - Custom error classes
- `packages/browser-git/src/utils/compression.ts` - Compression utilities using browser APIs
- `packages/browser-git/src/utils/encoding.ts` - Text encoding/decoding utilities
- `packages/browser-git/test/repository.test.ts` - Repository API tests
- `packages/browser-git/test/filesystem.test.ts` - File system API tests
- `packages/browser-git/test/wasm-loader.test.ts` - WASM loader tests
- `packages/browser-git/package.json` - Package configuration
- `packages/browser-git/tsconfig.json` - TypeScript configuration
- `packages/browser-git/vitest.config.ts` - Vitest test configuration
- `packages/browser-git/README.md` - Package documentation

### Package: storage-adapters
- `packages/storage-adapters/src/index.ts` - Storage adapter exports
- `packages/storage-adapters/src/interface.ts` - Common storage interface definition
- `packages/storage-adapters/src/indexeddb.ts` - IndexedDB storage adapter
- `packages/storage-adapters/src/opfs.ts` - OPFS storage adapter
- `packages/storage-adapters/src/localstorage.ts` - LocalStorage storage adapter
- `packages/storage-adapters/src/memory.ts` - In-memory storage adapter
- `packages/storage-adapters/src/utils/quota.ts` - Storage quota management utilities
- `packages/storage-adapters/src/utils/serialization.ts` - Data serialization helpers
- `packages/storage-adapters/test/indexeddb.test.ts` - IndexedDB adapter tests
- `packages/storage-adapters/test/opfs.test.ts` - OPFS adapter tests
- `packages/storage-adapters/test/localstorage.test.ts` - LocalStorage adapter tests
- `packages/storage-adapters/test/memory.test.ts` - Memory adapter tests
- `packages/storage-adapters/test/interface.test.ts` - Interface compliance tests
- `packages/storage-adapters/package.json` - Package configuration
- `packages/storage-adapters/tsconfig.json` - TypeScript configuration
- `packages/storage-adapters/README.md` - Package documentation

### Package: diff-engine
- `packages/diff-engine/src/index.ts` - Diff engine exports
- `packages/diff-engine/src/interface.ts` - Pluggable diff interface
- `packages/diff-engine/src/myers-diff.ts` - Default Myers diff implementation
- `packages/diff-engine/src/types.ts` - Diff result type definitions
- `packages/diff-engine/src/utils/text.ts` - Text processing utilities
- `packages/diff-engine/src/utils/binary.ts` - Binary diff utilities
- `packages/diff-engine/test/myers-diff.test.ts` - Myers diff tests
- `packages/diff-engine/test/interface.test.ts` - Interface tests
- `packages/diff-engine/package.json` - Package configuration
- `packages/diff-engine/tsconfig.json` - TypeScript configuration
- `packages/diff-engine/README.md` - Package documentation

### Package: git-cli
- `packages/git-cli/src/index.ts` - CLI entry point
- `packages/git-cli/src/commands/init.ts` - git init command
- `packages/git-cli/src/commands/add.ts` - git add command
- `packages/git-cli/src/commands/commit.ts` - git commit command
- `packages/git-cli/src/commands/status.ts` - git status command
- `packages/git-cli/src/commands/log.ts` - git log command
- `packages/git-cli/src/commands/diff.ts` - git diff command
- `packages/git-cli/src/commands/branch.ts` - git branch command
- `packages/git-cli/src/commands/checkout.ts` - git checkout command
- `packages/git-cli/src/commands/merge.ts` - git merge command
- `packages/git-cli/src/commands/clone.ts` - git clone command
- `packages/git-cli/src/commands/fetch.ts` - git fetch command
- `packages/git-cli/src/commands/pull.ts` - git pull command
- `packages/git-cli/src/commands/push.ts` - git push command
- `packages/git-cli/src/utils/parser.ts` - Argument parsing utilities
- `packages/git-cli/src/utils/output.ts` - Formatted output utilities
- `packages/git-cli/test/commands.test.ts` - CLI command tests
- `packages/git-cli/package.json` - Package configuration with bin entry
- `packages/git-cli/tsconfig.json` - TypeScript configuration
- `packages/git-cli/README.md` - CLI usage documentation

### Examples
- `examples/basic-demo/index.html` - Basic demo HTML page
- `examples/basic-demo/src/main.ts` - Basic demo application
- `examples/basic-demo/package.json` - Demo package configuration
- `examples/basic-demo/vite.config.ts` - Vite configuration
- `examples/mini-ide/src/App.tsx` - Mini IDE React application
- `examples/mini-ide/src/components/FileTree.tsx` - File tree component
- `examples/mini-ide/src/components/Editor.tsx` - Code editor component
- `examples/mini-ide/src/components/GitPanel.tsx` - Git operations panel
- `examples/mini-ide/package.json` - Mini IDE package configuration
- `examples/offline-docs/src/App.tsx` - Documentation site with version control
- `examples/offline-docs/package.json` - Offline docs package configuration

### Documentation
- `docs/README.md` - Documentation site index
- `docs/getting-started.md` - Getting started guide
- `docs/api-reference/repository.md` - Repository API reference
- `docs/api-reference/filesystem.md` - File system API reference
- `docs/api-reference/storage.md` - Storage adapters reference
- `docs/architecture/overview.md` - Architecture overview
- `docs/architecture/storage-layer.md` - Storage layer design
- `docs/architecture/wasm-bridge.md` - WASM bridge design
- `docs/guides/integration.md` - Integration guide for IDEs
- `docs/guides/cors-workarounds.md` - CORS handling guide
- `docs/guides/authentication.md` - Authentication setup guide
- `docs/browser-compatibility.md` - Browser compatibility matrix

### Testing
- `tests/integration/basic-workflow.test.ts` - Basic Git workflow integration tests
- `tests/integration/branching.test.ts` - Branch and merge integration tests
- `tests/integration/remote-operations.test.ts` - Clone, fetch, push integration tests
- `tests/browser/playwright.config.ts` - Playwright configuration
- `tests/browser/chrome.test.ts` - Chrome-specific tests
- `tests/browser/firefox.test.ts` - Firefox-specific tests
- `tests/browser/safari.test.ts` - Safari-specific tests
- `tests/performance/benchmarks.ts` - Performance benchmark suite

## Tasks

### Phase 1: Foundation & Infrastructure

- [x] 1.0 Set up monorepo infrastructure and development environment
  - [x] 1.1 Initialize Git repository with initial commit
  - [x] 1.2 Create root `package.json` with yarn workspaces configuration
  - [x] 1.3 Create directory structure for all packages (git-core, browser-git, storage-adapters, diff-engine, git-cli)
  - [x] 1.4 Create directory structure for examples and tests
  - [x] 1.5 Set up `.gitignore` with patterns for node_modules, dist, build artifacts, WASM files
  - [x] 1.6 Create base `tsconfig.base.json` with shared TypeScript configuration
  - [x] 1.7 Document monorepo setup in root `README.md`
  - [x] 1.8 Create `CONTRIBUTING.md` with development guidelines

- [x] 2.0 Implement storage abstraction layer with multiple backends
  - [x] 2.1 Create `storage-adapters` package structure with `package.json` and TypeScript config
  - [x] 2.2 Define common storage interface in `interface.ts` (methods: get, set, delete, list, exists, clear)
  - [x] 2.3 Implement IndexedDB storage adapter with proper database schema and transaction handling
  - [x] 2.4 Implement OPFS storage adapter with error handling for unsupported browsers
  - [x] 2.5 Implement LocalStorage adapter with size limit checking and fallback behavior
  - [x] 2.6 Implement in-memory storage adapter for testing purposes
  - [x] 2.7 Create storage quota management utilities to monitor and report storage usage
  - [x] 2.8 Implement data serialization helpers for binary data and JSON
  - [x] 2.9 Write unit tests for each storage adapter ensuring interface compliance
  - [x] 2.10 Write cross-storage adapter tests to verify consistent behavior
  - [x] 2.11 Add browser feature detection for OPFS and IndexedDB support

- [x] 3.0 Create file system API layer
  - [x] 3.1 Design and implement Node.js-like async fs API in `browser-git/src/filesystem/fs.ts`
  - [x] 3.2 Implement `readFile(path, encoding?)` with support for text and binary data
  - [x] 3.3 Implement `writeFile(path, data, encoding?)` with automatic directory creation
  - [x] 3.4 Implement `mkdir(path, recursive?)` for directory creation
  - [x] 3.5 Implement `readdir(path)` for listing directory contents
  - [x] 3.6 Implement `unlink(path)` for file deletion
  - [x] 3.7 Implement `rmdir(path, recursive?)` for directory deletion
  - [x] 3.8 Implement `stat(path)` for file metadata retrieval
  - [x] 3.9 Create path manipulation utilities (join, dirname, basename, normalize, relative)
  - [x] 3.10 Implement file watching API with event emitters for file changes
  - [x] 3.11 Write unit tests for all fs operations
  - [x] 3.12 Write integration tests combining fs operations with storage adapters

- [x] 4.0 Set up testing infrastructure and CI/CD pipeline
  - [x] 4.1 Install and configure Vitest for unit testing across all TypeScript packages
  - [x] 4.2 Create shared Vitest configuration for consistent test setup
  - [x] 4.3 Install and configure Playwright for integration and cross-browser testing
  - [x] 4.4 Create Playwright configuration for Chrome, Firefox, and Safari
  - [x] 4.5 Set up test utilities and helpers for common test scenarios
  - [x] 4.6 Create mock storage adapter for deterministic testing
  - [x] 4.7 Set up GitHub Actions workflow for CI (lint, type-check, unit tests)
  - [x] 4.8 Configure GitHub Actions to run Playwright tests on multiple browsers
  - [x] 4.9 Set up code coverage reporting with coverage thresholds (>80%)
  - [x] 4.10 Create performance benchmark infrastructure
  - [x] 4.11 Configure test scripts in root package.json (test, test:unit, test:integration, test:browser)

- [x] 5.0 Configure build system for Go/WASM and TypeScript
  - [x] 5.1 Install TinyGo and verify installation
  - [x] 5.2 Create `packages/git-core/go.mod` with module definition
  - [x] 5.3 Create `packages/git-core/Makefile` with WASM build targets
  - [x] 5.4 Configure TinyGo WASM build with optimization flags
  - [x] 5.5 Set up WASM binary output to `packages/browser-git/dist/` directory
  - [x] 5.6 Configure TypeScript build for each package (browser-git, storage-adapters, diff-engine, git-cli)
  - [x] 5.7 Set up build order dependencies (storage-adapters → diff-engine → git-core → browser-git → git-cli)
  - [x] 5.8 Create build scripts in root package.json (build, build:wasm, build:packages)
  - [x] 5.9 Set up watch mode for development (separate for WASM and TypeScript)
  - [x] 5.10 Configure source maps for debugging
  - [x] 5.11 Set up bundle size analysis and tracking
  - [x] 5.12 Create clean scripts to remove build artifacts

### Phase 2: Core Git Operations

- [ ] 6.0 Implement hash algorithm abstraction (SHA-1 and SHA-256)
  - [ ] 6.1 Create hash interface in `packages/git-core/pkg/hash/interface.go`
  - [ ] 6.2 Implement SHA-1 hasher with standard library
  - [ ] 6.3 Implement SHA-256 hasher with standard library
  - [ ] 6.4 Create hash factory function to select algorithm based on repository config
  - [ ] 6.5 Implement hash utilities (hex encoding, comparison, validation)
  - [ ] 6.6 Write unit tests for both hash implementations
  - [ ] 6.7 Verify hash outputs match standard Git for known inputs
  - [ ] 6.8 Export hash functionality to TypeScript through WASM bridge

- [ ] 7.0 Implement Git object model (blob, tree, commit, tag)
  - [ ] 7.1 Define object interface with serialize/deserialize methods
  - [ ] 7.2 Implement blob object with content storage
  - [ ] 7.3 Implement tree object with entries (mode, name, hash)
  - [ ] 7.4 Implement commit object with tree, parent(s), author, committer, message
  - [ ] 7.5 Implement tag object with target, tagger, message
  - [ ] 7.6 Implement object serialization to Git format
  - [ ] 7.7 Implement object deserialization from Git format
  - [ ] 7.8 Create object database interface for storage/retrieval
  - [ ] 7.9 Implement object compression using Go's compress/zlib
  - [ ] 7.10 Write unit tests for each object type with serialization round-trip tests
  - [ ] 7.11 Export object model to TypeScript with proper type definitions

- [ ] 8.0 Implement repository initialization and configuration
  - [ ] 8.1 Implement `git init` to create repository structure (.git directory structure)
  - [ ] 8.2 Create HEAD file pointing to initial branch (refs/heads/main)
  - [ ] 8.3 Create config file with default configuration
  - [ ] 8.4 Implement config parser for reading/writing repository configuration
  - [ ] 8.5 Add support for configuring hash algorithm (SHA-1 or SHA-256)
  - [ ] 8.6 Add support for configuring user name and email
  - [ ] 8.7 Implement repository detection (finding .git directory)
  - [ ] 8.8 Create TypeScript API: `Repository.init(path, options)`
  - [ ] 8.9 Write unit tests for repository initialization
  - [ ] 8.10 Write integration tests verifying complete repository structure

- [ ] 9.0 Implement staging and commit operations
  - [ ] 9.1 Implement index (staging area) data structure
  - [ ] 9.2 Implement `git add` to stage files (add to index)
  - [ ] 9.3 Implement `git add` with glob patterns and .gitignore support
  - [ ] 9.4 Implement .gitignore pattern matching using gitignore library
  - [ ] 9.5 Implement index serialization/deserialization
  - [ ] 9.6 Implement `git commit` to create commit objects from staged changes
  - [ ] 9.7 Update HEAD and branch references after commit
  - [ ] 9.8 Implement tree building from index entries
  - [ ] 9.9 Implement `git status` showing working tree vs index vs HEAD
  - [ ] 9.10 Create TypeScript APIs: `repo.add(paths)`, `repo.commit(message, options)`
  - [ ] 9.11 Write unit tests for staging operations
  - [ ] 9.12 Write integration tests for complete add-commit workflow

- [ ] 10.0 Implement branch management
  - [ ] 10.1 Implement reference storage (refs/heads/*, refs/tags/*)
  - [ ] 10.2 Implement `git branch` to list branches
  - [ ] 10.3 Implement `git branch <name>` to create new branch
  - [ ] 10.4 Implement `git branch -d <name>` to delete branch
  - [ ] 10.5 Implement `git branch -m <old> <new>` to rename branch
  - [ ] 10.6 Implement symbolic reference handling (HEAD → refs/heads/main)
  - [ ] 10.7 Implement current branch detection
  - [ ] 10.8 Create TypeScript APIs: `repo.branch()`, `repo.createBranch(name)`, `repo.deleteBranch(name)`
  - [ ] 10.9 Write unit tests for branch operations
  - [ ] 10.10 Write integration tests for branch lifecycle

- [ ] 11.0 Implement checkout operations
  - [ ] 11.1 Implement `git checkout <branch>` to switch branches
  - [ ] 11.2 Implement working directory update (write files from tree)
  - [ ] 11.3 Implement index update to match checked out tree
  - [ ] 11.4 Implement HEAD update to point to new branch
  - [ ] 11.5 Implement `git checkout <commit>` for detached HEAD state
  - [ ] 11.6 Implement `git checkout -- <file>` to restore files from index
  - [ ] 11.7 Add safety checks for uncommitted changes
  - [ ] 11.8 Implement file conflict detection during checkout
  - [ ] 11.9 Create TypeScript APIs: `repo.checkout(target, options)`
  - [ ] 11.10 Write unit tests for checkout operations
  - [ ] 11.11 Write integration tests for branch switching scenarios

- [ ] 12.0 Implement history and log operations
  - [ ] 12.1 Implement commit graph traversal (parents, ancestors)
  - [ ] 12.2 Implement `git log` with commit history display
  - [ ] 12.3 Add filtering options (author, date range, path)
  - [ ] 12.4 Implement `git log --oneline` format
  - [ ] 12.5 Implement `git log --graph` for branch visualization
  - [ ] 12.6 Implement commit lookup by hash (full and abbreviated)
  - [ ] 12.7 Implement `git show <commit>` to display commit details
  - [ ] 12.8 Implement `git blame` for line-by-line history
  - [ ] 12.9 Create TypeScript APIs: `repo.log(options)`, `repo.getCommit(hash)`, `repo.blame(path)`
  - [ ] 12.10 Write unit tests for history traversal
  - [ ] 12.11 Write integration tests for log with various options

### Phase 3: Diff & Merge

- [ ] 13.0 Create pluggable diff-engine package
  - [ ] 13.1 Create `diff-engine` package structure
  - [ ] 13.2 Define pluggable diff interface in `interface.ts`
  - [ ] 13.3 Define diff result types (hunks, changes, line numbers)
  - [ ] 13.4 Research and select battle-tested diff library (e.g., diff-match-patch, jsdiff)
  - [ ] 13.5 Implement Myers diff algorithm wrapper
  - [ ] 13.6 Implement line-by-line diff for text files
  - [ ] 13.7 Implement word-level diff support
  - [ ] 13.8 Implement binary file diff detection
  - [ ] 13.9 Create text processing utilities (line splitting, whitespace handling)
  - [ ] 13.10 Implement diff formatting (unified diff, side-by-side)
  - [ ] 13.11 Write unit tests for diff engine with various input scenarios
  - [ ] 13.12 Export diff engine to git-core through WASM bridge

- [ ] 14.0 Implement merge operations with conflict detection
  - [ ] 14.1 Implement three-way merge algorithm
  - [ ] 14.2 Find merge base (common ancestor) between two branches
  - [ ] 14.3 Implement file-level merge (content merging for modified files)
  - [ ] 14.4 Implement tree merging (directory structure changes)
  - [ ] 14.5 Detect merge conflicts (both sides modified same lines)
  - [ ] 14.6 Handle binary file conflicts
  - [ ] 14.7 Implement fast-forward merge detection and execution
  - [ ] 14.8 Create merge commit with two parents
  - [ ] 14.9 Update working directory with merge results
  - [ ] 14.10 Create TypeScript API: `repo.merge(branch, options)`
  - [ ] 14.11 Write unit tests for merge scenarios (fast-forward, three-way, conflicts)
  - [ ] 14.12 Write integration tests for complete merge workflows

- [ ] 15.0 Implement conflict resolution API
  - [ ] 15.1 Define structured conflict data type (base, ours, theirs, path)
  - [ ] 15.2 Return structured conflict objects from merge operation
  - [ ] 15.3 Include conflict metadata (line ranges, change types)
  - [ ] 15.4 Implement utility to generate Git-style conflict markers (<<<<<<, ======, >>>>>>)
  - [ ] 15.5 Implement `repo.getConflicts()` to retrieve current conflicts
  - [ ] 15.6 Implement `repo.resolveConflict(path, resolution)` to mark conflicts as resolved
  - [ ] 15.7 Support resolution strategies (accept-ours, accept-theirs, manual)
  - [ ] 15.8 Update index after conflict resolution
  - [ ] 15.9 Write unit tests for conflict detection and resolution
  - [ ] 15.10 Write integration tests for conflict resolution workflows

### Phase 4: Remote Operations

- [ ] 16.0 Implement HTTP Git protocol (smart protocol)
  - [ ] 16.1 Research and document Git HTTP smart protocol specification
  - [ ] 16.2 Implement discovery phase (GET /info/refs?service=git-upload-pack)
  - [ ] 16.3 Parse advertisement of remote references
  - [ ] 16.4 Implement negotiation protocol (want/have exchange)
  - [ ] 16.5 Implement packfile format reader
  - [ ] 16.6 Implement packfile format writer
  - [ ] 16.7 Implement delta object decoding
  - [ ] 16.8 Implement delta object encoding
  - [ ] 16.9 Handle packfile compression/decompression
  - [ ] 16.10 Implement CORS detection and error handling with helpful messages
  - [ ] 16.11 Write unit tests for protocol parsing and packfile handling
  - [ ] 16.12 Write integration tests against real Git server (using test fixtures)

- [ ] 17.0 Implement authentication layer
  - [ ] 17.1 Define authentication interface with pluggable providers
  - [ ] 17.2 Implement HTTP Basic Authentication with username/token
  - [ ] 17.3 Implement OAuth flow integration hooks (callback handling)
  - [ ] 17.4 Research SSH key management in browser (Web Crypto API)
  - [ ] 17.5 Implement SSH key-based authentication (if feasible)
  - [ ] 17.6 Allow consumers to provide custom authentication handlers
  - [ ] 17.7 Implement credential storage using browser Credential Management API
  - [ ] 17.8 Add authentication to HTTP requests with proper headers
  - [ ] 17.9 Handle authentication errors with clear user-facing messages
  - [ ] 17.10 Create TypeScript API: `repo.setAuth(config)`
  - [ ] 17.11 Write unit tests for authentication providers
  - [ ] 17.12 Document authentication setup for GitHub, GitLab, etc.

- [ ] 18.0 Implement clone operation
  - [ ] 18.1 Implement `git clone <url> <path>` command structure
  - [ ] 18.2 Fetch remote references from repository
  - [ ] 18.3 Determine default branch (usually main or master)
  - [ ] 18.4 Download packfile with all objects
  - [ ] 18.5 Unpack objects into local object database
  - [ ] 18.6 Create local branches tracking remote branches
  - [ ] 18.7 Checkout default branch (HEAD) to working directory
  - [ ] 18.8 Set up remote configuration (origin)
  - [ ] 18.9 Handle shallow clone if requested (depth parameter)
  - [ ] 18.10 Show progress information during clone
  - [ ] 18.11 Create TypeScript API: `Repository.clone(url, path, options)`
  - [ ] 18.12 Write integration tests for cloning public repositories

- [ ] 19.0 Implement fetch and pull operations
  - [ ] 19.1 Implement `git fetch` to retrieve remote changes
  - [ ] 19.2 Update remote tracking branches (refs/remotes/origin/*)
  - [ ] 19.3 Download new objects not present locally
  - [ ] 19.4 Handle force updates and deletions
  - [ ] 19.5 Implement fetch refspec parsing and handling
  - [ ] 19.6 Implement `git pull` as fetch + merge
  - [ ] 19.7 Support pull with rebase option
  - [ ] 19.8 Handle fast-forward pulls
  - [ ] 19.9 Handle merge conflicts during pull
  - [ ] 19.10 Create TypeScript APIs: `repo.fetch(remote, options)`, `repo.pull(options)`
  - [ ] 19.11 Write unit tests for fetch protocol
  - [ ] 19.12 Write integration tests for fetch and pull workflows

- [ ] 20.0 Implement push operation
  - [ ] 20.1 Implement `git push` to send local commits to remote
  - [ ] 20.2 Determine which commits need to be sent
  - [ ] 20.3 Create packfile with objects to push
  - [ ] 20.4 Send packfile to remote using POST request
  - [ ] 20.5 Handle remote rejection (non-fast-forward)
  - [ ] 20.6 Implement force push with safety warnings
  - [ ] 20.7 Support pushing tags
  - [ ] 20.8 Support deleting remote branches/tags
  - [ ] 20.9 Handle authentication during push
  - [ ] 20.10 Show progress information during push
  - [ ] 20.11 Create TypeScript API: `repo.push(remote, branch, options)`
  - [ ] 20.12 Write integration tests for push operations

### Phase 5: CLI, Examples & Documentation

- [ ] 21.0 Create full-featured CLI tool
  - [ ] 21.1 Create `git-cli` package structure with bin entry point
  - [ ] 21.2 Set up command-line argument parsing (use commander or yargs)
  - [ ] 21.3 Implement `bgit init` command mirroring git init
  - [ ] 21.4 Implement `bgit add` command with glob support
  - [ ] 21.5 Implement `bgit commit` command with message flag
  - [ ] 21.6 Implement `bgit status` command with colored output
  - [ ] 21.7 Implement `bgit log` command with formatting options
  - [ ] 21.8 Implement `bgit diff` command with unified diff output
  - [ ] 21.9 Implement `bgit branch` command with create/delete/list
  - [ ] 21.10 Implement `bgit checkout` command
  - [ ] 21.11 Implement `bgit merge` command with conflict handling
  - [ ] 21.12 Implement `bgit clone` command with progress display
  - [ ] 21.13 Implement `bgit fetch`, `bgit pull`, `bgit push` commands
  - [ ] 21.14 Add help documentation for all commands
  - [ ] 21.15 Add version flag and about information
  - [ ] 21.16 Create formatted output utilities (tables, colors, symbols)
  - [ ] 21.17 Write unit tests for CLI command parsing
  - [ ] 21.18 Write integration tests for CLI workflows
  - [ ] 21.19 Make CLI executable with proper shebang and chmod

- [ ] 22.0 Create example applications
  - [ ] 22.1 Create `basic-demo` example with simple HTML/JS interface
  - [ ] 22.2 Implement basic operations in demo (init, add, commit, view log)
  - [ ] 22.3 Add Vite configuration for building basic demo
  - [ ] 22.4 Create `mini-ide` example with React
  - [ ] 22.5 Implement file tree component for browsing repository
  - [ ] 22.6 Implement code editor component (use Monaco or CodeMirror)
  - [ ] 22.7 Implement Git panel with staging, commit, branch switching
  - [ ] 22.8 Add diff viewer for file changes
  - [ ] 22.9 Create `offline-docs` example with Markdown editor
  - [ ] 22.10 Implement version history for documentation changes
  - [ ] 22.11 Add search across documentation versions
  - [ ] 22.12 Write README for each example with setup instructions
  - [ ] 22.13 Deploy examples to GitHub Pages or Vercel for live demos

- [ ] 23.0 Write comprehensive documentation
  - [ ] 23.1 Set up Docusaurus site in `docs/` directory
  - [ ] 23.2 Write getting started guide with installation and first steps
  - [ ] 23.3 Write API reference for Repository class and all methods
  - [ ] 23.4 Write API reference for file system API
  - [ ] 23.5 Write API reference for storage adapters
  - [ ] 23.6 Write architecture overview explaining high-level design
  - [ ] 23.7 Document storage layer architecture and trade-offs
  - [ ] 23.8 Document WASM bridge design and JS-Go communication
  - [ ] 23.9 Write integration guide for browser-based IDEs
  - [ ] 23.10 Write CORS workarounds guide with examples
  - [ ] 23.11 Write authentication setup guide for GitHub/GitLab
  - [ ] 23.12 Create browser compatibility matrix table
  - [ ] 23.13 Document limitations and known issues
  - [ ] 23.14 Add migration guide from other Git libraries (if applicable)
  - [ ] 23.15 Generate API documentation from TypeScript using TypeDoc
  - [ ] 23.16 Add code examples throughout documentation
  - [ ] 23.17 Set up documentation site deployment

- [ ] 24.0 Perform cross-browser testing and optimization
  - [ ] 24.1 Run integration tests on Chrome with Playwright
  - [ ] 24.2 Run integration tests on Firefox with Playwright
  - [ ] 24.3 Run integration tests on Safari with Playwright
  - [ ] 24.4 Identify and fix browser-specific issues
  - [ ] 24.5 Test storage adapter behavior across browsers
  - [ ] 24.6 Test WASM loading and performance on each browser
  - [ ] 24.7 Run performance benchmarks on all target browsers
  - [ ] 24.8 Optimize commit operations (<50ms target)
  - [ ] 24.9 Optimize checkout operations (<200ms target)
  - [ ] 24.10 Optimize clone operations (<5s for 100-commit repo)
  - [ ] 24.11 Optimize WASM bundle size (<2MB gzipped)
  - [ ] 24.12 Profile memory usage and optimize allocations
  - [ ] 24.13 Add browser feature detection and graceful degradation
  - [ ] 24.14 Update browser compatibility documentation with test results

- [ ] 25.0 Conduct security audit and prepare for release
  - [ ] 25.1 Review code for common security vulnerabilities (OWASP Top 10)
  - [ ] 25.2 Ensure no execution of arbitrary code (no eval, no Function constructor)
  - [ ] 25.3 Validate and sanitize all user inputs (paths, URLs, commands)
  - [ ] 25.4 Review remote URL handling to prevent SSRF attacks
  - [ ] 25.5 Review credential storage and ensure no plaintext passwords
  - [ ] 25.6 Verify CSP compatibility for WASM execution
  - [ ] 25.7 Review CORS handling and error messages
  - [ ] 25.8 Test with various malformed inputs and edge cases
  - [ ] 25.9 Create SECURITY.md with vulnerability reporting process
  - [ ] 25.10 Write release notes highlighting features and limitations
  - [ ] 25.11 Create npm publish scripts with version bumping
  - [ ] 25.12 Set up automated release workflow with GitHub Actions
  - [ ] 25.13 Prepare announcement blog post or documentation
  - [ ] 25.14 Publish all packages to npm registry
  - [ ] 25.15 Create GitHub release with binaries and changelog
  - [ ] 25.16 Share on relevant communities (Reddit, HackerNews, Twitter)

## Notes

- **Testing Strategy:** Unit tests should be written alongside implementation for each sub-task. Integration tests should be written after completing related parent tasks.
- **Development Order:** Follow the task order as listed. Phase 1 must be completed before starting Phase 2, and so on.
- **WASM Development:** Use `make build` in `packages/git-core/` to rebuild WASM. The output should automatically be copied to the TypeScript package.
- **Storage Backend:** For initial development, use in-memory storage for speed. Switch to IndexedDB for integration testing.
- **Git Protocol:** Test remote operations against a local Git server or public repositories on GitHub.
- **Performance:** Run benchmarks frequently during Phase 2-4 to catch performance regressions early.
- **Documentation:** Write documentation incrementally as features are implemented, not all at once in Phase 5.
- **Security:** Consider security implications at every step, especially when handling user input and remote operations.

## Running Tests

- **Unit Tests:** `yarn test:unit` (runs Vitest across all packages)
- **Integration Tests:** `yarn test:integration` (runs integration test suite)
- **Browser Tests:** `yarn test:browser` (runs Playwright tests on all browsers)
- **All Tests:** `yarn test` (runs all test suites)
- **Watch Mode:** `yarn test:watch` (runs tests in watch mode during development)
- **Coverage:** `yarn test:coverage` (generates coverage report)

## Build Commands

- **Build All:** `yarn build` (builds WASM + all TypeScript packages)
- **Build WASM:** `yarn build:wasm` (builds only Go/WASM)
- **Build Packages:** `yarn build:packages` (builds only TypeScript packages)
- **Watch Mode:** `yarn dev` (watches and rebuilds on changes)
- **Clean:** `yarn clean` (removes all build artifacts)
