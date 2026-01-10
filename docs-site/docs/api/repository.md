---
sidebar_position: 1
---

# Repository API

The `Repository` class is the main entry point for Git operations in BrowserGit.

## Creating Repositories

### Repository.init()

Initialize a new Git repository.

```typescript
static async init(
  path: string,
  options?: RepositoryOptions
): Promise<Repository>
```

**Parameters:**

- `path` - The path where the repository will be created
- `options` - Optional configuration

**Options:**

```typescript
interface RepositoryOptions {
  storage?: "indexeddb" | "opfs" | "localstorage" | "memory";
  hashAlgorithm?: "sha1" | "sha256";
  defaultBranch?: string; // Default: 'main'
}
```

**Example:**

```typescript
const repo = await Repository.init("/my-project", {
  storage: "indexeddb",
  hashAlgorithm: "sha1",
  defaultBranch: "main",
});
```

### Repository.open()

Open an existing repository.

```typescript
static async open(path: string): Promise<Repository | null>
```

**Parameters:**

- `path` - Path to the repository

**Returns:** The repository, or `null` if not found.

**Example:**

```typescript
const repo = await Repository.open("/my-project");
if (!repo) {
  console.log("Repository not found");
}
```

### Repository.clone()

Clone a remote repository.

```typescript
static async clone(
  url: string,
  path: string,
  options?: CloneOptions
): Promise<Repository>
```

**Parameters:**

- `url` - Remote repository URL
- `path` - Local path for the clone
- `options` - Clone configuration

**Options:**

```typescript
interface CloneOptions extends RepositoryOptions {
  auth?: GitAuth;
  corsProxy?: string;
  depth?: number; // Shallow clone depth
  branch?: string; // Branch to clone
  onProgress?: (progress: CloneProgress) => void;
}

interface GitAuth {
  type: "token" | "basic" | "callback";
  token?: string;
  username?: string;
  password?: string;
  callback?: (url: string, operation: string) => Promise<GitAuth>;
}
```

**Example:**

```typescript
const repo = await Repository.clone(
  "https://github.com/user/repo.git",
  "/local-repo",
  {
    storage: "indexeddb",
    auth: { type: "token", token: "ghp_xxx" },
    onProgress: (p) => console.log(`${p.phase}: ${p.loaded}/${p.total}`),
  },
);
```

## Properties

### repo.path

```typescript
readonly path: string
```

The filesystem path of the repository.

### repo.fs

```typescript
readonly fs: FileSystem
```

The filesystem instance for this repository. See [FileSystem API](./filesystem).

## Basic Operations

### repo.add()

Stage files for commit.

```typescript
async add(paths: string[]): Promise<void>
```

**Parameters:**

- `paths` - Array of file paths or glob patterns

**Example:**

```typescript
// Add specific files
await repo.add(["src/index.ts", "README.md"]);

// Add all files
await repo.add(["."]);

// Add by pattern
await repo.add(["src/**/*.ts"]);
```

### repo.commit()

Create a commit with staged changes.

```typescript
async commit(
  message: string,
  options: CommitOptions
): Promise<Commit>
```

**Options:**

```typescript
interface CommitOptions {
  author: Author;
  committer?: Author; // Defaults to author
  allowEmpty?: boolean;
  amend?: boolean;
}

interface Author {
  name: string;
  email: string;
  date?: Date;
}
```

**Returns:**

```typescript
interface Commit {
  hash: string;
  message: string;
  author: Author;
  committer: Author;
  parentHashes: string[];
  treeHash: string;
}
```

**Example:**

```typescript
const commit = await repo.commit("Add new feature", {
  author: {
    name: "John Doe",
    email: "john@example.com",
  },
});

console.log("Created commit:", commit.hash);
```

### repo.status()

Get the current repository status.

```typescript
async status(): Promise<Status>
```

**Returns:**

```typescript
interface Status {
  current: string; // Current branch
  staged: FileStatus[];
  modified: FileStatus[];
  untracked: string[];
  deleted: string[];
  renamed: Array<{ from: string; to: string }>;
  conflicted: string[];
}

interface FileStatus {
  path: string;
  status: "added" | "modified" | "deleted" | "renamed";
  oldPath?: string; // For renamed files
}
```

**Example:**

```typescript
const status = await repo.status();

console.log("On branch:", status.current);
console.log(
  "Staged files:",
  status.staged.map((f) => f.path),
);
console.log(
  "Modified files:",
  status.modified.map((f) => f.path),
);
console.log("Untracked files:", status.untracked);
```

## Branch Operations

### repo.createBranch()

Create a new branch.

```typescript
async createBranch(
  name: string,
  options?: CreateBranchOptions
): Promise<void>
```

**Options:**

```typescript
interface CreateBranchOptions {
  startPoint?: string; // Commit or branch to start from
  force?: boolean; // Overwrite existing branch
}
```

**Example:**

```typescript
// Create from current HEAD
await repo.createBranch("feature/new-feature");

// Create from specific commit
await repo.createBranch("hotfix", { startPoint: "abc1234" });
```

### repo.deleteBranch()

Delete a branch.

```typescript
async deleteBranch(
  name: string,
  options?: DeleteBranchOptions
): Promise<void>
```

**Options:**

```typescript
interface DeleteBranchOptions {
  force?: boolean; // Delete even if not fully merged
}
```

### repo.listBranches()

List all branches.

```typescript
async listBranches(
  options?: ListBranchesOptions
): Promise<Branch[]>
```

**Options:**

```typescript
interface ListBranchesOptions {
  remote?: boolean; // Include remote branches
  all?: boolean; // Include both local and remote
}
```

**Returns:**

```typescript
interface Branch {
  name: string;
  commit: string;
  isRemote: boolean;
  isHead: boolean; // Currently checked out
  upstream?: string; // Tracking branch
}
```

### repo.getCurrentBranch()

Get the current branch name.

```typescript
async getCurrentBranch(): Promise<string>
```

### repo.checkout()

Switch branches or restore files.

```typescript
async checkout(
  target: string,
  options?: CheckoutOptions
): Promise<void>
```

**Options:**

```typescript
interface CheckoutOptions {
  create?: boolean; // Create branch if doesn't exist
  force?: boolean; // Discard local changes
  paths?: string[]; // Checkout specific paths only
}
```

**Example:**

```typescript
// Switch branch
await repo.checkout("feature/new-feature");

// Create and switch
await repo.checkout("feature/another", { create: true });

// Restore file from HEAD
await repo.checkout("HEAD", { paths: ["src/index.ts"] });
```

## History Operations

### repo.log()

Get commit history.

```typescript
async log(options?: LogOptions): Promise<Commit[]>
```

**Options:**

```typescript
interface LogOptions {
  maxCount?: number;
  skip?: number;
  since?: Date;
  until?: Date;
  author?: string;
  path?: string; // Filter by file path
  ref?: string; // Starting ref (default: HEAD)
}
```

**Example:**

```typescript
// Recent commits
const commits = await repo.log({ maxCount: 10 });

// Commits for a specific file
const fileHistory = await repo.log({ path: "src/index.ts" });

// Commits by author
const myCommits = await repo.log({ author: "john@example.com" });
```

### repo.diff()

Compute differences between commits or working tree.

```typescript
async diff(options?: DiffOptions): Promise<DiffResult>
```

**Options:**

```typescript
interface DiffOptions {
  from?: string; // Starting commit (default: HEAD)
  to?: string; // Ending commit (default: working tree)
  paths?: string[]; // Filter by paths
  contextLines?: number; // Lines of context (default: 3)
  ignoreWhitespace?: boolean;
}
```

**Returns:**

```typescript
interface DiffResult {
  files: FileDiff[];
}

interface FileDiff {
  path: string;
  oldPath?: string; // For renames
  status: "added" | "modified" | "deleted" | "renamed";
  binary: boolean;
  hunks: DiffHunk[];
  additions: number;
  deletions: number;
}

interface DiffHunk {
  oldStart: number;
  oldLines: number;
  newStart: number;
  newLines: number;
  changes: Change[];
}

interface Change {
  type: "add" | "delete" | "context";
  content: string;
  oldLineNumber?: number;
  newLineNumber?: number;
}
```

**Example:**

```typescript
// Diff working tree against HEAD
const diff = await repo.diff();

// Diff between commits
const diff2 = await repo.diff({ from: "abc123", to: "def456" });

// Show diff for specific file
const fileDiff = await repo.diff({ paths: ["src/index.ts"] });
```

### repo.blame()

Show what revision and author last modified each line.

```typescript
async blame(path: string): Promise<BlameLine[]>
```

**Returns:**

```typescript
interface BlameLine {
  lineNumber: number;
  commit: string;
  author: Author;
  content: string;
  originalLineNumber: number;
}
```

## Merge Operations

### repo.merge()

Merge a branch into the current branch.

```typescript
async merge(
  branch: string,
  options?: MergeOptions
): Promise<MergeResult>
```

**Options:**

```typescript
interface MergeOptions {
  message?: string; // Custom merge commit message
  noCommit?: boolean; // Don't create merge commit
  strategy?: "recursive" | "ours" | "theirs";
}
```

**Returns:**

```typescript
interface MergeResult {
  success: boolean;
  commit?: string; // Merge commit hash
  conflicts: ConflictFile[];
}

interface ConflictFile {
  path: string;
  ours: string;
  theirs: string;
  ancestor: string;
}
```

**Example:**

```typescript
const result = await repo.merge("feature/new-feature");

if (result.conflicts.length > 0) {
  console.log(
    "Conflicts in:",
    result.conflicts.map((c) => c.path),
  );
} else {
  console.log("Merge successful:", result.commit);
}
```

## Remote Operations

### repo.addRemote()

Add a remote repository.

```typescript
async addRemote(name: string, url: string): Promise<void>
```

### repo.removeRemote()

Remove a remote repository.

```typescript
async removeRemote(name: string): Promise<void>
```

### repo.listRemotes()

List configured remotes.

```typescript
async listRemotes(): Promise<Remote[]>
```

**Returns:**

```typescript
interface Remote {
  name: string;
  fetchUrl: string;
  pushUrl: string;
}
```

### repo.fetch()

Download objects and refs from a remote.

```typescript
async fetch(
  remote: string,
  options?: FetchOptions
): Promise<FetchResult>
```

**Options:**

```typescript
interface FetchOptions {
  auth?: GitAuth;
  corsProxy?: string;
  prune?: boolean;
  depth?: number;
  onProgress?: (progress: TransferProgress) => void;
}
```

### repo.pull()

Fetch and merge remote changes.

```typescript
async pull(
  remote: string,
  branch: string,
  options?: PullOptions
): Promise<MergeResult>
```

### repo.push()

Upload local commits to remote.

```typescript
async push(
  remote: string,
  branch: string,
  options?: PushOptions
): Promise<void>
```

**Options:**

```typescript
interface PushOptions {
  auth?: GitAuth;
  corsProxy?: string;
  force?: boolean;
  setUpstream?: boolean;
  onProgress?: (progress: TransferProgress) => void;
}
```

**Example:**

```typescript
await repo.push("origin", "main", {
  auth: { type: "token", token: "ghp_xxx" },
  setUpstream: true,
  onProgress: (p) => console.log(`Pushing: ${p.loaded}/${p.total}`),
});
```

## Tag Operations

### repo.createTag()

Create a tag.

```typescript
async createTag(
  name: string,
  options?: CreateTagOptions
): Promise<void>
```

**Options:**

```typescript
interface CreateTagOptions {
  ref?: string; // Commit to tag (default: HEAD)
  message?: string; // For annotated tags
  force?: boolean;
}
```

### repo.deleteTag()

Delete a tag.

```typescript
async deleteTag(name: string): Promise<void>
```

### repo.listTags()

List all tags.

```typescript
async listTags(): Promise<Tag[]>
```

## Stash Operations

### repo.stash()

Stash current changes.

```typescript
async stash(options?: StashOptions): Promise<string>
```

### repo.stashPop()

Apply and remove the latest stash.

```typescript
async stashPop(): Promise<void>
```

### repo.stashList()

List stashed changes.

```typescript
async stashList(): Promise<Stash[]>
```

## Utility Methods

### repo.resolveRef()

Resolve a reference to a commit hash.

```typescript
async resolveRef(ref: string): Promise<string>
```

### repo.getCommit()

Get a commit by hash.

```typescript
async getCommit(hash: string): Promise<Commit>
```

### repo.getTree()

Get a tree object.

```typescript
async getTree(hash: string): Promise<TreeEntry[]>
```

### repo.getBlob()

Get file contents from a commit.

```typescript
async getBlob(hash: string): Promise<Uint8Array>
```

## See Also

- [FileSystem API](./filesystem) - File operations
- [Storage Adapters](./storage-adapters) - Storage backends
- [Diff Engine](./diff-engine) - Diff computation
