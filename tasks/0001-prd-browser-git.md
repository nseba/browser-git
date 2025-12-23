# PRD: Browser-Based Git Implementation (BrowserGit)

**Status:** ✅ Approved
**Last Updated:** 2025-11-12
**Implementation Timeline:** 20 weeks (5 phases)
**Target Release:** Q2 2026

## Executive Summary

BrowserGit is a full-featured, browser-native Git implementation built with Go + WebAssembly and TypeScript. It enables offline-first web applications to have complete version control capabilities with no server dependencies. The library targets browser-based IDEs and web applications requiring local version control.

**Key Differentiators:**

- **Complete Git Support:** All core commands (init, add, commit, branch, merge, clone, push, pull, etc.)
- **Multiple Storage Backends:** IndexedDB, OPFS, LocalStorage, in-memory
- **Dual API Layers:** Low-level FS API + high-level Git operations API
- **Future-Proof:** SHA-1 and SHA-256 hash algorithm support from day one
- **Pluggable Architecture:** Modular diff engine, storage adapters, and auth handlers
- **Production-Ready:** >80% test coverage, cross-browser compatibility, comprehensive documentation

**Technology Stack:** Go (TinyGo for WASM) + TypeScript + Yarn workspaces

## Introduction/Overview

BrowserGit is a full-featured Git implementation designed to run entirely in web browsers, enabling offline-first web applications to have complete version control capabilities without server dependencies. The library will be built using Go compiled to WebAssembly for performance and will provide a comprehensive storage abstraction layer that makes it easy for browser-based IDEs and applications to manage project files locally.

**Problem Statement:** Modern web applications, particularly browser-based IDEs and collaborative tools, need robust version control capabilities but are constrained by browser environments. Existing solutions either require server-side Git operations or provide limited Git functionality. This creates barriers for offline-first applications and increases infrastructure complexity.

**Solution:** A complete, browser-native Git implementation that stores repositories in browser storage (IndexedDB, OPFS) with full support for Git commands and an intuitive file system abstraction layer.

## Goals

1. **Full Git Command Support:** Implement all core Git commands (init, add, commit, branch, merge, checkout, clone, fetch, push, pull, log, diff, blame, status) with browser-optimized performance
2. **Cross-Browser Compatibility:** Support latest 2 versions of Chrome/Edge, Firefox, and Safari
3. **Easy Integration:** Provide simple, well-documented APIs that browser-based IDEs can integrate in hours, not days
4. **Pragmatic Git Compatibility:** Maintain compatibility with standard Git repositories where practical, while optimizing for browser constraints
5. **Multiple Storage Backends:** Support IndexedDB (primary), OPFS, LocalStorage (fallback), and in-memory storage
6. **Flexible Architecture:** Provide both low-level filesystem APIs and high-level Git operation APIs
7. **Production-Ready Quality:** Comprehensive test coverage (unit, integration, cross-browser) with performance benchmarks

## User Stories

**As a browser-based IDE developer**, I want to integrate Git version control so that my users can manage their code with familiar Git workflows without leaving the browser.

**As a web application developer**, I want to add offline version control to my app so that users can work without internet connectivity and sync changes when they reconnect.

**As an end user of a browser-based IDE**, I want to commit changes, create branches, and merge code just like I do with desktop Git, so that I have a consistent development experience.

**As a developer integrating BrowserGit**, I want clear documentation and examples so that I can understand how to read/write files and execute Git operations without deep knowledge of Git internals.

**As a performance-conscious developer**, I want Git operations to feel instant (sub-100ms for common operations) so that the user experience matches native applications.

## Functional Requirements

### Core Git Operations

1. **FR-1.1:** The library MUST implement `git init` to initialize a new repository in browser storage
2. **FR-1.2:** The library MUST implement `git add <paths>` to stage files for commit
3. **FR-1.3:** The library MUST implement `git commit -m <message>` with author/committer metadata
4. **FR-1.4:** The library MUST implement `git status` showing working directory and staging area state
5. **FR-1.5:** The library MUST implement `git log` with customizable output formats and filtering
6. **FR-1.6:** The library MUST implement `git diff` for comparing working directory, staging, and commits
7. **FR-1.7:** The library MUST implement `git branch` operations (create, list, delete, rename)
8. **FR-1.8:** The library MUST implement `git checkout` for switching branches and checking out files
9. **FR-1.9:** The library MUST implement `git merge` with automatic and manual conflict resolution
10. **FR-1.10:** The library MUST implement `git clone <url>` to clone remote repositories via HTTP(S)
11. **FR-1.11:** The library MUST implement `git fetch` to retrieve updates from remote repositories
12. **FR-1.12:** The library MUST implement `git pull` (fetch + merge/rebase)
13. **FR-1.13:** The library MUST implement `git push` to send commits to remote repositories
14. **FR-1.14:** The library MUST implement `git blame` to show line-by-line commit history
15. **FR-1.15:** The library MUST support `.gitignore` pattern matching for excluding files

### Storage Abstraction Layer

16. **FR-2.1:** The library MUST provide an IndexedDB storage backend as the primary storage mechanism
17. **FR-2.2:** The library MUST provide an OPFS (Origin Private File System) storage backend for modern browsers
18. **FR-2.3:** The library MUST provide a LocalStorage backend as a fallback for storage-limited scenarios
19. **FR-2.4:** The library MUST provide an in-memory storage backend for testing and ephemeral operations
20. **FR-2.5:** Storage backends MUST implement a common interface allowing runtime switching
21. **FR-2.6:** The library MUST efficiently handle binary files (images, compiled assets, etc.)
22. **FR-2.7:** The library MUST provide storage quota management and monitoring APIs

### File System API

23. **FR-3.1:** The library MUST expose a Node.js fs-like async API (readFile, writeFile, mkdir, readdir, unlink, etc.)
24. **FR-3.2:** The library MUST expose a high-level Git-aware API (repo.addFile(), repo.commit(), repo.checkout(), etc.)
25. **FR-3.3:** File operations MUST support both text (UTF-8) and binary data
26. **FR-3.4:** The FS API MUST support path operations (join, dirname, basename, normalize)
27. **FR-3.5:** The FS API MUST provide watch/notification capabilities for file changes

### Remote Operations & Authentication

28. **FR-4.1:** The library MUST support HTTP(S) Git protocol (smart protocol) for clone/fetch/push
29. **FR-4.2:** The library MUST support HTTP Basic Authentication with tokens
30. **FR-4.3:** The library MUST provide OAuth integration hooks for GitHub/GitLab authentication flows
31. **FR-4.4:** The library MUST support SSH key-based authentication (browser crypto APIs)
32. **FR-4.5:** The library MUST allow consumers to provide custom authentication handlers
33. **FR-4.6:** Remote operations MUST handle CORS properly and provide clear error messages for CORS issues

### Module & Distribution

34. **FR-5.1:** The library MUST be distributed as an ES Module (ESM) package
35. **FR-5.2:** The library MUST provide CommonJS compatibility
36. **FR-5.3:** The library MUST include TypeScript definition files (.d.ts)
37. **FR-5.4:** The WASM binary MUST be <2MB gzipped for fast loading
38. **FR-5.5:** The library MUST work with modern bundlers (webpack, Rollup, Vite, esbuild)
39. **FR-5.6:** The library MUST be published to npm registry

### Testing & Quality

40. **FR-6.1:** The library MUST have >80% unit test coverage for all modules
41. **FR-6.2:** The library MUST include integration tests covering complete Git workflows
42. **FR-6.3:** The library MUST run tests across Chrome, Firefox, and Safari
43. **FR-6.4:** The library MUST include performance benchmarks for critical operations
44. **FR-6.5:** The test suite MUST run in CI/CD on every commit

### Documentation

45. **FR-7.1:** The library MUST provide complete API reference documentation
46. **FR-7.2:** The library MUST include a comprehensive integration guide with code examples
47. **FR-7.3:** The library MUST include at least 3 example applications demonstrating common use cases
48. **FR-7.4:** The library MUST provide architecture documentation explaining internal design
49. **FR-7.5:** Documentation MUST include browser compatibility matrix
50. **FR-7.6:** Documentation MUST include CORS workarounds and remote operation setup guide

### Hash Algorithm & Future-Proofing

51. **FR-8.1:** The library MUST abstract hash algorithm with pluggable interface
52. **FR-8.2:** The library MUST support SHA-1 hash algorithm (current Git standard)
53. **FR-8.3:** The library MUST support SHA-256 hash algorithm (future Git standard)
54. **FR-8.4:** Repository configuration MUST allow specifying hash algorithm preference
55. **FR-8.5:** The library MUST correctly interoperate with repositories using different hash algorithms

### Diff Engine

56. **FR-9.1:** The library MUST provide a pluggable diff engine interface in separate `diff-engine` package
57. **FR-9.2:** The library MUST include default diff implementation using battle-tested library
58. **FR-9.3:** The diff API MUST support custom diff algorithm implementations
59. **FR-9.4:** Default diff implementation MUST handle both text and binary files appropriately

### Conflict Resolution

60. **FR-10.1:** Merge conflicts MUST return structured conflict objects with base, ours, and theirs content
61. **FR-10.2:** The library MUST provide utility method to generate Git-style conflict markers from structured conflicts
62. **FR-10.3:** Conflict objects MUST include file path, conflict regions, and metadata

### CLI Tool

63. **FR-11.1:** The library MUST include a full-featured CLI tool in `git-cli` package
64. **FR-11.2:** CLI MUST mirror standard Git commands for familiarity
65. **FR-11.3:** CLI MUST support testing and debugging workflows
66. **FR-11.4:** CLI MUST work with all supported storage backends

## Non-Goals (Out of Scope)

1. **NG-1:** Support for Git submodules (may be added in future versions)
2. **NG-2:** Support for Git LFS (Large File Storage) in initial release
3. **NG-3:** Git worktree functionality
4. **NG-4:** Support for Internet Explorer or legacy browsers
5. **NG-5:** Server-side Git operations or Git hosting capabilities
6. **NG-6:** Built-in UI components (library is headless, consumers build their own UI)
7. **NG-7:** Support for GitLab/Bitbucket-specific APIs beyond standard Git protocol
8. **NG-8:** Real-time collaboration features (operational transforms, CRDT)
9. **NG-9:** Git hooks execution (security concern in browser environment)
10. **NG-10:** Support for extremely large repositories (>10GB) in initial release
11. **NG-11:** Built-in CORS proxy service (document workarounds instead)
12. **NG-12:** File system import tool in initial release (may add later based on demand)
13. **NG-13:** Mobile browser optimization (focus on desktop browsers first)

## Design Considerations

### Architecture

**Monorepo Structure:**

```
browser-git/
├── packages/
│   ├── git-core/          # Go + WASM Git implementation
│   ├── browser-git/       # TypeScript wrapper & API layer
│   ├── storage-adapters/  # Storage backend implementations
│   ├── diff-engine/       # Pluggable diff algorithm implementation
│   └── git-cli/           # CLI tool for testing and debugging
├── examples/
│   ├── basic-demo/        # Simple demo app
│   ├── mini-ide/          # Minimal IDE example
│   └── offline-docs/      # Documentation site with version control
├── docs/                  # Documentation site
└── tests/
    ├── unit/
    ├── integration/
    └── browser/           # Cross-browser test suite
```

**Technology Stack:**

- **Core Implementation:** Go compiled to WebAssembly
- **API Layer:** TypeScript for type safety and developer experience
- **Build System:** Yarn workspaces, TinyGo for WASM compilation
- **Testing:** Vitest (unit), Playwright (integration & cross-browser)
- **Documentation:** TypeDoc for API docs, Docusaurus for guides

### Performance Strategy

- Lazy-load WASM binary on first Git operation
- Use worker threads for heavy operations to avoid blocking UI
- Implement efficient pack file format for storage optimization
- Cache frequently accessed objects in memory
- **Compression Strategy:** Hybrid approach
  - Use Git delta compression for packfiles (maintains compatibility)
  - Use browser-native CompressionStream API for storage layer compression
  - This provides optimal balance between Git compatibility and implementation simplicity
- **Repository Size:** Start with medium target (500MB / 5k files), expand based on performance testing

### Compatibility Strategy

- Support standard Git object format (commits, trees, blobs) for interoperability
- **Hash Algorithm Support:** Design for both SHA-1 and SHA-256 from the start
  - Abstract hash algorithm interface for extensibility
  - Support both SHA-1 (current standard) and SHA-256 (future-proof)
  - Allow repository-level configuration of hash algorithm
- Optimize packfile format for browser storage if needed
- Maintain standard ref format for branches/tags
- Use standard Git config format

### Diff Implementation Strategy

- **Pluggable Architecture:** Create `diff-engine` package with pluggable interface
- **Initial Implementation:** Use existing battle-tested diff library (e.g., diff-match-patch or similar)
- **Separation of Concerns:** Keep diff engine as separate package to enforce modularity
- Consumers can provide custom diff implementations if needed
- Default implementation suitable for most code diffing use cases

### Conflict Resolution Strategy

- **Dual Approach:** Support both structured data and Git-style markers
- **API Design:**
  - Return structured conflict objects with metadata (base, ours, theirs)
  - Include utility methods to generate Git-style conflict markers
  - Allow consumers to choose their preferred conflict representation
- Provides flexibility for different UI implementations

### CORS & Remote Operations

- **No built-in proxy:** Document workarounds for CORS issues
- Provide clear documentation on:
  - Setting up custom CORS proxies
  - Configuring Git servers with CORS headers
  - Using browser extensions for development
- Focus library effort on core Git functionality

## Technical Considerations

1. **WASM Considerations:**
   - TinyGo will be used for smaller binary size vs standard Go compiler
   - Memory management between JS and Go requires careful interface design
   - Consider using `wasm-bindgen`-style approach for Go-JS interop

2. **Storage Considerations:**
   - IndexedDB transaction limits (browser-specific)
   - OPFS requires secure context (HTTPS)
   - LocalStorage 5-10MB limits
   - Quota management across storage APIs

3. **Security Considerations:**
   - No execution of arbitrary code (no Git hooks)
   - Sanitize remote URLs to prevent SSRF attacks
   - Secure credential storage using browser Credential Management API
   - CSP compatibility for WASM execution

4. **Browser Constraints:**
   - SharedArrayBuffer requires cross-origin isolation headers
   - Service Workers may be needed for advanced features
   - File System Access API requires user permission prompts

5. **Dependencies:**
   - Minimal external dependencies to reduce bundle size
   - Pure Go standard library for Git implementation
   - Consider using `go-git` as reference implementation

## Success Metrics

1. **Adoption:** 100+ GitHub stars within 3 months of launch
2. **Performance:**
   - Commit operation <50ms for <1000 files
   - Checkout operation <200ms for typical branch switch
   - Clone operation <5s for 100-commit repository
3. **Integration Success:** 3+ browser-based IDEs/tools integrate the library
4. **Developer Experience:** Integration takes <4 hours for experienced developer
5. **Reliability:** <5 bug reports per 1000 users in first 6 months
6. **Test Coverage:** Maintain >80% code coverage with passing tests on all target browsers
7. **Bundle Size:** WASM + JS wrapper <2MB gzipped

## Architecture Decisions (Resolved)

All key architectural decisions have been finalized:

1. **Compression Strategy (Q1):** ✅ Hybrid approach
   - Git delta compression for packfiles (compatibility)
   - Browser-native CompressionStream for storage layer (simplicity)

2. **Repository Size Target (Q2):** ✅ Test-driven approach
   - Start with medium target: 500MB / 5k files
   - Expand based on performance testing results

3. **CORS Proxy (Q3):** ✅ Document workarounds
   - No built-in proxy service
   - Provide comprehensive documentation for CORS solutions

4. **File System Import (Q4):** ✅ Not in initial release
   - May add later based on user demand
   - Focus on core Git functionality first

5. **Hash Algorithm Support (Q5):** ✅ Support both SHA-1 and SHA-256
   - Abstract hash interface from the start
   - Future-proof for Git's evolution

6. **Diff Implementation (Q6):** ✅ Pluggable architecture
   - Separate `diff-engine` package
   - Use existing battle-tested library for default implementation
   - Allow custom implementations

7. **Conflict Resolution API (Q7):** ✅ Dual approach
   - Return structured conflict objects
   - Provide utility to generate Git-style markers
   - Maximum flexibility for consumers

8. **CLI Tool (Q8):** ✅ Full-featured CLI included
   - Separate `git-cli` package
   - Mirror standard Git commands
   - Essential for testing and debugging

## Remaining Open Questions

1. **Q-NEW-1:** Which specific diff library should we use for the default implementation?
   - Options: diff-match-patch, jsdiff, myers-diff
   - **Decision needed before:** Implementing diff-engine package

2. **Q-NEW-2:** Should we implement sparse checkout support?
   - **Impact:** Allows working with subset of large repositories
   - **Decision needed before:** Implementing checkout functionality

3. **Q-NEW-3:** What's the strategy for handling binary merge conflicts?
   - Options: Choose ours/theirs, manual resolution, external diff tools
   - **Decision needed before:** Implementing merge functionality

4. **Q-NEW-4:** Should we support partial clone (e.g., shallow clone, blob filters)?
   - **Impact:** Critical for large repository support
   - **Decision needed before:** Implementing clone functionality

## Implementation Phases

### Phase 1: Foundation (Weeks 1-3)

1. Create GitHub repository with monorepo structure
2. Set up yarn workspaces and build tooling
3. Configure TinyGo, TypeScript, and WASM build pipeline
4. Implement storage abstraction layer (IndexedDB, OPFS, LocalStorage, in-memory)
5. Create basic file system API layer
6. Set up testing infrastructure (Vitest, Playwright)
7. Write ADRs (Architecture Decision Records) for key choices

### Phase 2: Core Git Operations (Weeks 4-8)

1. Implement hash algorithm abstraction (SHA-1 & SHA-256)
2. Implement Git object model (blob, tree, commit, tag)
3. Implement basic operations: init, add, commit, status
4. Implement branch operations: branch, checkout
5. Implement log and history traversal
6. Create CLI tool skeleton
7. Write unit tests for core operations

### Phase 3: Diff & Merge (Weeks 9-12)

1. Implement pluggable diff-engine package
2. Integrate default diff library
3. Implement merge with conflict detection
4. Implement structured conflict resolution API
5. Add conflict marker generation utilities
6. Write integration tests for merge workflows

### Phase 4: Remote Operations (Weeks 13-16)

1. Implement HTTP Git protocol (smart protocol)
2. Implement authentication layer (Basic, OAuth hooks, SSH)
3. Implement clone, fetch, push, pull
4. Handle CORS scenarios and error messages
5. Write remote operation tests
6. Document CORS workarounds

### Phase 5: Polish & Documentation (Weeks 17-20)

1. Complete CLI tool with all commands
2. Create example applications (basic-demo, mini-ide, offline-docs)
3. Write comprehensive API documentation
4. Write integration guide
5. Write architecture documentation
6. Set up documentation site (Docusaurus)
7. Cross-browser testing and optimization
8. Performance benchmarking and optimization
9. Security audit
10. Prepare for npm publication

## Next Steps

**Immediate actions:**

1. ✅ Review and approve this PRD
2. Create GitHub repository for the project
3. Set up initial monorepo structure with yarn workspaces
4. Configure development environment (Go, TinyGo, Node.js, TypeScript)
5. Create initial package.json files for all packages
6. Set up CI/CD pipeline (GitHub Actions)
7. Begin Phase 1 implementation
