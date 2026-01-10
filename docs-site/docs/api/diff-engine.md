---
sidebar_position: 4
---

# Diff Engine API

The diff engine provides algorithms for computing differences between files and text content.

## DiffEngine Interface

All diff engines implement this interface:

```typescript
interface DiffEngine {
  diff(
    oldContent: string,
    newContent: string,
    options?: DiffOptions
  ): DiffResult;
}

interface DiffOptions {
  contextLines?: number; // Lines of context (default: 3)
  ignoreWhitespace?: boolean;
  ignoreCase?: boolean;
  algorithm?: 'myers' | 'patience' | 'histogram';
}
```

## MyersDiffEngine

The default diff algorithm, optimal for most use cases.

### Usage

```typescript
import { MyersDiffEngine } from '@browser-git/diff-engine';

const engine = new MyersDiffEngine();

const result = engine.diff(
  'line1\nline2\nline3',
  'line1\nmodified\nline3'
);

console.log(result.hunks);
```

### DiffResult

```typescript
interface DiffResult {
  hunks: DiffHunk[];
  additions: number;
  deletions: number;
  changes: number;
}

interface DiffHunk {
  oldStart: number;  // Starting line in old content
  oldLines: number;  // Number of lines in old content
  newStart: number;  // Starting line in new content
  newLines: number;  // Number of lines in new content
  changes: Change[];
}

interface Change {
  type: 'add' | 'delete' | 'context';
  content: string;
  oldLineNumber?: number;
  newLineNumber?: number;
}
```

### Example Output

```typescript
const oldContent = `function hello() {
  console.log("Hello");
}`;

const newContent = `function hello() {
  console.log("Hello, World!");
  return true;
}`;

const result = engine.diff(oldContent, newContent);

// result.hunks[0]:
// {
//   oldStart: 1,
//   oldLines: 3,
//   newStart: 1,
//   newLines: 4,
//   changes: [
//     { type: 'context', content: 'function hello() {' },
//     { type: 'delete', content: '  console.log("Hello");' },
//     { type: 'add', content: '  console.log("Hello, World!");' },
//     { type: 'add', content: '  return true;' },
//     { type: 'context', content: '}' }
//   ]
// }
```

## DiffEngineFactory

Create diff engines with specific configurations:

```typescript
import { DiffEngineFactory } from '@browser-git/diff-engine';

// Default Myers algorithm
const engine = DiffEngineFactory.create();

// With options
const customEngine = DiffEngineFactory.create({
  algorithm: 'patience',
  contextLines: 5
});
```

## Utility Functions

### computeDiff()

Convenience function for simple diffs:

```typescript
import { computeDiff } from '@browser-git/diff-engine';

const result = computeDiff(oldContent, newContent, {
  contextLines: 3
});
```

### formatPatch()

Generate unified diff format:

```typescript
import { formatPatch } from '@browser-git/diff-engine';

const patch = formatPatch({
  oldPath: 'a/file.txt',
  newPath: 'b/file.txt',
  oldContent,
  newContent
});

// Output:
// --- a/file.txt
// +++ b/file.txt
// @@ -1,3 +1,4 @@
//  function hello() {
// -  console.log("Hello");
// +  console.log("Hello, World!");
// +  return true;
//  }
```

### applyPatch()

Apply a patch to content:

```typescript
import { applyPatch } from '@browser-git/diff-engine';

const patched = applyPatch(originalContent, patch);
```

### parsePatch()

Parse unified diff format:

```typescript
import { parsePatch } from '@browser-git/diff-engine';

const patches = parsePatch(patchText);

for (const patch of patches) {
  console.log('File:', patch.oldPath, '->', patch.newPath);
  console.log('Hunks:', patch.hunks.length);
}
```

## Binary Diff

For binary files:

```typescript
import { BinaryDiff } from '@browser-git/diff-engine';

const differ = new BinaryDiff();

// Check if files are binary
const isBinary = differ.isBinary(content);

// Compute binary diff (returns whether files differ)
const result = differ.diff(oldBinary, newBinary);

console.log('Files differ:', result.different);
console.log('Old size:', result.oldSize);
console.log('New size:', result.newSize);
```

## Three-Way Merge

For merge operations:

```typescript
import { ThreeWayMerge } from '@browser-git/diff-engine';

const merger = new ThreeWayMerge();

const result = merger.merge({
  base: baseContent,
  ours: ourContent,
  theirs: theirContent
});

if (result.hasConflicts) {
  console.log('Conflicts found');
  console.log(result.conflictedContent);
} else {
  console.log('Merged successfully');
  console.log(result.mergedContent);
}
```

### Conflict Markers

When conflicts occur, the output includes markers:

```
Normal content here
<<<<<<< ours
Our changes
=======
Their changes
>>>>>>> theirs
More normal content
```

### Conflict Resolution

```typescript
interface MergeResult {
  hasConflicts: boolean;
  mergedContent: string;
  conflictedContent?: string;
  conflicts: ConflictRegion[];
}

interface ConflictRegion {
  startLine: number;
  endLine: number;
  base: string;
  ours: string;
  theirs: string;
}
```

## Word-Level Diff

For more granular diffs:

```typescript
import { wordDiff } from '@browser-git/diff-engine';

const result = wordDiff(
  'The quick brown fox',
  'The slow brown dog'
);

// result.changes:
// [
//   { type: 'equal', value: 'The ' },
//   { type: 'delete', value: 'quick' },
//   { type: 'add', value: 'slow' },
//   { type: 'equal', value: ' brown ' },
//   { type: 'delete', value: 'fox' },
//   { type: 'add', value: 'dog' }
// ]
```

## Character-Level Diff

For inline changes:

```typescript
import { charDiff } from '@browser-git/diff-engine';

const result = charDiff('hello', 'hallo');

// result.changes:
// [
//   { type: 'equal', value: 'h' },
//   { type: 'delete', value: 'e' },
//   { type: 'add', value: 'a' },
//   { type: 'equal', value: 'llo' }
// ]
```

## Custom Diff Engines

Implement your own diff algorithm:

```typescript
import { DiffEngine, DiffResult, DiffOptions } from '@browser-git/diff-engine';

class CustomDiffEngine implements DiffEngine {
  diff(
    oldContent: string,
    newContent: string,
    options?: DiffOptions
  ): DiffResult {
    // Your implementation
    const hunks = this.computeHunks(oldContent, newContent);

    return {
      hunks,
      additions: this.countAdditions(hunks),
      deletions: this.countDeletions(hunks),
      changes: hunks.length
    };
  }
}

// Register custom engine
DiffEngineFactory.register('custom', CustomDiffEngine);

// Use it
const engine = DiffEngineFactory.create({ algorithm: 'custom' });
```

## Performance Considerations

### Large Files

For large files, consider streaming:

```typescript
import { streamingDiff } from '@browser-git/diff-engine';

const result = await streamingDiff(oldStream, newStream, {
  chunkSize: 64 * 1024 // 64KB chunks
});
```

### Timeout Protection

Prevent runaway diffs:

```typescript
const result = await computeDiffWithTimeout(oldContent, newContent, {
  timeout: 5000 // 5 seconds max
});
```

## See Also

- [Repository API](./repository) - Using diff with repositories
- [Architecture Overview](../architecture/overview) - How diff fits in
