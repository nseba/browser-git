# Contributing to BrowserGit

Thank you for your interest in contributing to BrowserGit! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone. Please:

- Be respectful and considerate
- Welcome newcomers and help them get started
- Accept constructive criticism gracefully
- Focus on what is best for the community
- Show empathy towards others

## Getting Started

### Prerequisites

- Node.js >= 18.0.0
- Yarn >= 4.0.0
- Go >= 1.21
- TinyGo >= 0.30.0

### Initial Setup

1. Fork the repository on GitHub
2. Clone your fork locally:

   ```bash
   git clone https://github.com/YOUR_USERNAME/browser-git.git
   cd browser-git
   ```

3. Add the upstream repository:

   ```bash
   git remote add upstream https://github.com/nseba/browser-git.git
   ```

4. Install dependencies:

   ```bash
   yarn install
   ```

5. Build the project:

   ```bash
   yarn build
   ```

6. Run tests to verify setup:
   ```bash
   yarn test
   ```

## Development Workflow

### Creating a Branch

Always create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

Branch naming conventions:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or modifications
- `chore/` - Build process or auxiliary tool changes

### Making Changes

1. Make your changes in the appropriate package
2. Write or update tests for your changes
3. Run tests locally:

   ```bash
   yarn test
   ```

4. Run linting and type checking:

   ```bash
   yarn lint
   yarn typecheck
   ```

5. Format your code:
   ```bash
   yarn format
   ```

### Development Commands

```bash
# Start development mode with hot reload
yarn dev

# Build WASM core
yarn build:wasm

# Build TypeScript packages
yarn build:packages

# Build everything
yarn build

# Run all tests
yarn test

# Run tests in watch mode
yarn test:watch

# Run specific test suite
yarn test:unit
yarn test:integration
yarn test:browser

# Generate coverage report
yarn test:coverage

# Clean build artifacts
yarn clean
```

## Project Structure

```
browser-git/
├── packages/
│   ├── git-core/          # Go + WASM implementation
│   ├── browser-git/       # TypeScript API wrapper
│   ├── storage-adapters/  # Storage backends
│   ├── diff-engine/       # Diff algorithms
│   └── git-cli/           # CLI tool
├── examples/              # Example applications
├── tests/                 # Integration and browser tests
├── docs/                  # Documentation
└── tasks/                 # Project planning documents
```

### Package Guidelines

- Each package should have its own `package.json`, `tsconfig.json`, and `README.md`
- Keep dependencies minimal and well-justified
- Use workspace dependencies for internal packages
- Export only public APIs from package entry points

## Coding Standards

### TypeScript

- Use TypeScript strict mode
- Prefer explicit types over implicit `any`
- Use interfaces for public APIs
- Document public functions with JSDoc comments
- Follow functional programming principles where appropriate

Example:

```typescript
/**
 * Initializes a new Git repository.
 *
 * @param path - The path where the repository will be created
 * @param options - Configuration options for the repository
 * @returns A promise that resolves to the initialized Repository
 * @throws {RepositoryError} If the directory already contains a repository
 */
export async function init(
  path: string,
  options: InitOptions,
): Promise<Repository> {
  // Implementation
}
```

### Go

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions small and focused
- Handle errors explicitly
- Write tests for exported functions

### Code Organization

- One feature per file when possible
- Group related functionality
- Keep files under 500 lines
- Extract reusable logic into utilities

### Naming Conventions

- **TypeScript**: camelCase for variables/functions, PascalCase for classes/types
- **Go**: Follow Go standard (camelCase for private, PascalCase for public)
- **Files**: kebab-case for file names
- **Directories**: kebab-case

## Testing

### Testing Strategy

- Write unit tests for all new code
- Maintain >80% code coverage
- Write integration tests for complete workflows
- Test across Chrome, Firefox, and Safari

### Writing Tests

Use Vitest for TypeScript tests:

```typescript
import { describe, it, expect, beforeEach } from "vitest";
import { Repository } from "../src/repository";

describe("Repository", () => {
  let repo: Repository;

  beforeEach(async () => {
    repo = await Repository.init("/test-repo", { storage: "memory" });
  });

  it("should initialize with correct default branch", () => {
    expect(repo.currentBranch()).toBe("main");
  });

  it("should commit changes successfully", async () => {
    await repo.add(["file.txt"]);
    const commit = await repo.commit("Test commit");
    expect(commit.message).toBe("Test commit");
  });
});
```

Use Go's testing package for WASM tests:

```go
func TestHashSHA1(t *testing.T) {
    hasher := NewSHA1()
    data := []byte("hello world")
    hash := hasher.Hash(data)

    expected := "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
    if hex.EncodeToString(hash) != expected {
        t.Errorf("Expected %s, got %s", expected, hex.EncodeToString(hash))
    }
}
```

### Running Tests

```bash
# Run all tests
yarn test

# Run specific package tests
cd packages/browser-git && yarn test

# Run with coverage
yarn test:coverage

# Run in watch mode during development
yarn test:watch
```

## Commit Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, missing semicolons, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes
- `perf`: Performance improvements

### Examples

```
feat(storage): add OPFS storage adapter

Implements Origin Private File System storage adapter for better
performance in modern browsers. Includes automatic fallback to
IndexedDB for unsupported browsers.

Closes #123
```

```
fix(merge): handle binary file conflicts correctly

Binary files were being corrupted during merge conflicts. This fix
ensures binary content is preserved and conflict resolution is
handled appropriately.

Fixes #456
```

### Commit Best Practices

- Write clear, concise commit messages
- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor to..." not "moves cursor to...")
- Limit subject line to 72 characters
- Reference issues and pull requests where appropriate
- Break large changes into smaller, logical commits

## Pull Request Process

### Before Submitting

1. Update your branch with latest upstream:

   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Ensure all tests pass:

   ```bash
   yarn test
   ```

3. Verify code quality:

   ```bash
   yarn lint
   yarn typecheck
   ```

4. Update documentation if needed
5. Add tests for new functionality
6. Update CHANGELOG.md if appropriate

### Submitting a Pull Request

1. Push your branch to your fork:

   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a pull request on GitHub

3. Fill out the PR template completely:
   - Clear description of changes
   - Link to related issues
   - Screenshots/videos if UI changes
   - Checklist completion

4. Request review from maintainers

### PR Requirements

- [ ] All tests pass
- [ ] Code coverage maintained (>80%)
- [ ] No linting errors
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (for significant changes)
- [ ] Commit messages follow conventions
- [ ] Branch is up to date with main

### Review Process

- Maintainers will review your PR
- Address feedback and make requested changes
- Push additional commits to the same branch
- Once approved, a maintainer will merge your PR

## Release Process

Releases are handled by maintainers:

1. Update version in all package.json files
2. Update CHANGELOG.md
3. Create a git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
4. Push tag: `git push upstream v1.0.0`
5. GitHub Actions will build and publish packages
6. Create GitHub release with changelog

## Getting Help

- 📖 Read the [documentation](./docs)
- 💬 Join [discussions](https://github.com/nseba/browser-git/discussions)
- 🐛 Check [existing issues](https://github.com/nseba/browser-git/issues)
- ❓ Ask questions in discussions

## Recognition

Contributors will be recognized in:

- Project README.md
- Release notes
- GitHub contributors page

Thank you for contributing to BrowserGit! 🎉
