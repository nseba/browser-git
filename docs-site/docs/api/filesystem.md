---
sidebar_position: 2
---

# FileSystem API

The `FileSystem` class provides a Node.js-like filesystem interface for browser storage.

## Accessing the FileSystem

```typescript
import { Repository } from "@browser-git/browser-git";

const repo = await Repository.init("/project", { storage: "indexeddb" });
const fs = repo.fs;
```

Or standalone:

```typescript
import { FileSystem } from "@browser-git/browser-git";
import { IndexedDBAdapter } from "@browser-git/storage-adapters";

const adapter = new IndexedDBAdapter("my-fs");
await adapter.initialize();

const fs = new FileSystem(adapter);
```

## File Operations

### fs.readFile()

Read the contents of a file.

```typescript
async readFile(
  path: string,
  encoding?: 'utf-8' | null
): Promise<string | Uint8Array>
```

**Parameters:**

- `path` - Absolute path to the file
- `encoding` - If `'utf-8'`, returns string; otherwise returns `Uint8Array`

**Example:**

```typescript
// Read as string
const content = await fs.readFile("/project/README.md", "utf-8");

// Read as binary
const binary = await fs.readFile("/project/image.png");
```

### fs.writeFile()

Write data to a file.

```typescript
async writeFile(
  path: string,
  data: string | Uint8Array,
  options?: WriteFileOptions
): Promise<void>
```

**Options:**

```typescript
interface WriteFileOptions {
  encoding?: "utf-8";
  mode?: number; // Unix file mode (default: 0o644)
  flag?: "w" | "a"; // Write or append
}
```

**Example:**

```typescript
// Write string
await fs.writeFile("/project/file.txt", "Hello, World!");

// Write binary
await fs.writeFile("/project/data.bin", new Uint8Array([1, 2, 3, 4]));

// Append to file
await fs.writeFile("/project/log.txt", "New line\n", { flag: "a" });
```

### fs.appendFile()

Append data to a file.

```typescript
async appendFile(
  path: string,
  data: string | Uint8Array
): Promise<void>
```

### fs.unlink()

Delete a file.

```typescript
async unlink(path: string): Promise<void>
```

**Example:**

```typescript
await fs.unlink("/project/temp.txt");
```

### fs.rename()

Rename or move a file.

```typescript
async rename(oldPath: string, newPath: string): Promise<void>
```

**Example:**

```typescript
// Rename file
await fs.rename("/project/old.txt", "/project/new.txt");

// Move file
await fs.rename("/project/file.txt", "/project/subdir/file.txt");
```

### fs.copyFile()

Copy a file.

```typescript
async copyFile(src: string, dest: string): Promise<void>
```

## Directory Operations

### fs.mkdir()

Create a directory.

```typescript
async mkdir(
  path: string,
  options?: MkdirOptions
): Promise<void>
```

**Options:**

```typescript
interface MkdirOptions {
  recursive?: boolean; // Create parent directories
  mode?: number; // Unix directory mode (default: 0o755)
}
```

**Example:**

```typescript
// Create single directory
await fs.mkdir("/project/src");

// Create nested directories
await fs.mkdir("/project/src/components/ui", { recursive: true });
```

### fs.rmdir()

Remove a directory.

```typescript
async rmdir(
  path: string,
  options?: RmdirOptions
): Promise<void>
```

**Options:**

```typescript
interface RmdirOptions {
  recursive?: boolean; // Remove contents recursively
}
```

**Example:**

```typescript
// Remove empty directory
await fs.rmdir("/project/empty");

// Remove directory and contents
await fs.rmdir("/project/temp", { recursive: true });
```

### fs.readdir()

Read directory contents.

```typescript
async readdir(
  path: string,
  options?: ReaddirOptions
): Promise<string[] | Dirent[]>
```

**Options:**

```typescript
interface ReaddirOptions {
  withFileTypes?: boolean; // Return Dirent objects
  recursive?: boolean; // Include subdirectories
}
```

**Returns:**

```typescript
interface Dirent {
  name: string;
  isFile(): boolean;
  isDirectory(): boolean;
  isSymbolicLink(): boolean;
}
```

**Example:**

```typescript
// List file names
const files = await fs.readdir("/project/src");
// ['index.ts', 'utils', 'components']

// List with file types
const entries = await fs.readdir("/project/src", { withFileTypes: true });
for (const entry of entries) {
  if (entry.isDirectory()) {
    console.log(`📁 ${entry.name}`);
  } else {
    console.log(`📄 ${entry.name}`);
  }
}

// Recursive listing
const allFiles = await fs.readdir("/project", { recursive: true });
```

## File Information

### fs.stat()

Get file or directory information.

```typescript
async stat(path: string): Promise<Stats>
```

**Returns:**

```typescript
interface Stats {
  dev: number;
  ino: number;
  mode: number;
  nlink: number;
  uid: number;
  gid: number;
  size: number;
  atime: Date;
  mtime: Date;
  ctime: Date;
  birthtime: Date;

  isFile(): boolean;
  isDirectory(): boolean;
  isSymbolicLink(): boolean;
}
```

**Example:**

```typescript
const stats = await fs.stat("/project/file.txt");

console.log("Size:", stats.size, "bytes");
console.log("Modified:", stats.mtime);
console.log("Is file:", stats.isFile());
```

### fs.lstat()

Like `stat()`, but doesn't follow symbolic links.

```typescript
async lstat(path: string): Promise<Stats>
```

### fs.exists()

Check if a path exists.

```typescript
async exists(path: string): Promise<boolean>
```

**Example:**

```typescript
if (await fs.exists("/project/config.json")) {
  const config = await fs.readFile("/project/config.json", "utf-8");
}
```

### fs.access()

Check file accessibility.

```typescript
async access(path: string, mode?: number): Promise<void>
```

**Modes:**

```typescript
import { constants } from "@browser-git/browser-git";

await fs.access(path, constants.F_OK); // File exists
await fs.access(path, constants.R_OK); // Readable
await fs.access(path, constants.W_OK); // Writable
```

## Symbolic Links

### fs.symlink()

Create a symbolic link.

```typescript
async symlink(target: string, path: string): Promise<void>
```

### fs.readlink()

Read the target of a symbolic link.

```typescript
async readlink(path: string): Promise<string>
```

## Path Operations

### fs.realpath()

Resolve a path to its absolute form.

```typescript
async realpath(path: string): Promise<string>
```

## Watching Files

### fs.watch()

Watch for file changes.

```typescript
watch(
  path: string,
  options?: WatchOptions,
  listener?: WatchListener
): FSWatcher
```

**Options:**

```typescript
interface WatchOptions {
  persistent?: boolean;
  recursive?: boolean;
}

type WatchListener = (eventType: "rename" | "change", filename: string) => void;
```

**Returns:**

```typescript
interface FSWatcher {
  close(): void;
}
```

**Example:**

```typescript
const watcher = fs.watch(
  "/project/src",
  { recursive: true },
  (event, filename) => {
    console.log(`${event}: ${filename}`);
  },
);

// Later, stop watching
watcher.close();
```

## Stream Operations

### fs.createReadStream()

Create a readable stream.

```typescript
createReadStream(
  path: string,
  options?: ReadStreamOptions
): ReadableStream<Uint8Array>
```

**Options:**

```typescript
interface ReadStreamOptions {
  start?: number;
  end?: number;
  highWaterMark?: number;
}
```

**Example:**

```typescript
const stream = fs.createReadStream("/project/large-file.bin");
const reader = stream.getReader();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  processChunk(value);
}
```

### fs.createWriteStream()

Create a writable stream.

```typescript
createWriteStream(
  path: string,
  options?: WriteStreamOptions
): WritableStream<Uint8Array>
```

**Example:**

```typescript
const stream = fs.createWriteStream("/project/output.bin");
const writer = stream.getWriter();

await writer.write(new Uint8Array([1, 2, 3]));
await writer.write(new Uint8Array([4, 5, 6]));
await writer.close();
```

## Utility Methods

### fs.glob()

Find files matching a pattern.

```typescript
async glob(pattern: string): Promise<string[]>
```

**Example:**

```typescript
// Find all TypeScript files
const tsFiles = await fs.glob("/project/**/*.ts");

// Find files in specific directory
const srcFiles = await fs.glob("/project/src/*.js");
```

### fs.readJSON()

Read and parse a JSON file.

```typescript
async readJSON<T>(path: string): Promise<T>
```

**Example:**

```typescript
interface Config {
  name: string;
  version: string;
}

const config = await fs.readJSON<Config>("/project/package.json");
console.log(config.name);
```

### fs.writeJSON()

Stringify and write JSON to a file.

```typescript
async writeJSON(
  path: string,
  data: unknown,
  options?: { spaces?: number }
): Promise<void>
```

**Example:**

```typescript
await fs.writeJSON("/project/config.json", { debug: true }, { spaces: 2 });
```

## Error Handling

FileSystem operations throw errors with specific codes:

```typescript
import { FileSystemError } from "@browser-git/browser-git";

try {
  await fs.readFile("/nonexistent");
} catch (error) {
  if (error instanceof FileSystemError) {
    switch (error.code) {
      case "ENOENT":
        console.log("File not found");
        break;
      case "EACCES":
        console.log("Permission denied");
        break;
      case "EISDIR":
        console.log("Is a directory");
        break;
      case "ENOTDIR":
        console.log("Not a directory");
        break;
      case "EEXIST":
        console.log("File already exists");
        break;
      case "ENOTEMPTY":
        console.log("Directory not empty");
        break;
    }
  }
}
```

## Path Utilities

Import path utilities:

```typescript
import { path } from "@browser-git/browser-git";

path.join("/project", "src", "index.ts");
// '/project/src/index.ts'

path.dirname("/project/src/index.ts");
// '/project/src'

path.basename("/project/src/index.ts");
// 'index.ts'

path.extname("/project/src/index.ts");
// '.ts'

path.normalize("/project//src/../src/./index.ts");
// '/project/src/index.ts'

path.isAbsolute("/project/src");
// true

path.relative("/project", "/project/src/index.ts");
// 'src/index.ts'
```

## See Also

- [Repository API](./repository) - High-level Git operations
- [Storage Adapters](./storage-adapters) - Storage backends
