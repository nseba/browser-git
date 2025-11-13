/**
 * Represents a type of change in a diff
 */
export enum ChangeType {
  /** Content was added */
  Add = 'add',
  /** Content was deleted */
  Delete = 'delete',
  /** Content was modified */
  Modify = 'modify',
  /** Content remained unchanged (context) */
  Equal = 'equal',
}

/**
 * Represents a single change within a diff
 */
export interface Change {
  /** Type of change */
  type: ChangeType;
  /** The content of the change */
  value: string;
  /** Line number in the old file (for deletes and modifications) */
  oldLineNumber?: number;
  /** Line number in the new file (for adds and modifications) */
  newLineNumber?: number;
  /** Number of lines this change spans */
  count?: number;
}

/**
 * Represents a hunk (contiguous block of changes) in a diff
 */
export interface DiffHunk {
  /** Starting line number in the old file */
  oldStart: number;
  /** Number of lines from the old file in this hunk */
  oldLines: number;
  /** Starting line number in the new file */
  newStart: number;
  /** Number of lines from the new file in this hunk */
  newLines: number;
  /** Array of changes within this hunk */
  changes: Change[];
  /** Optional header text for the hunk (e.g., function name) */
  header?: string;
}

/**
 * Result of a diff operation
 */
export interface DiffResult {
  /** Array of hunks representing the changes */
  hunks: DiffHunk[];
  /** Whether the files are binary */
  isBinary: boolean;
  /** Old file path (optional) */
  oldPath?: string;
  /** New file path (optional) */
  newPath?: string;
  /** Old file mode (optional, for Git compatibility) */
  oldMode?: string;
  /** New file mode (optional, for Git compatibility) */
  newMode?: string;
  /** Whether the file was added */
  isNew?: boolean;
  /** Whether the file was deleted */
  isDeleted?: boolean;
  /** Total lines added */
  linesAdded: number;
  /** Total lines deleted */
  linesDeleted: number;
}

/**
 * Options for diff operations
 */
export interface DiffOptions {
  /** Number of context lines to include around changes (default: 3) */
  context?: number;
  /** Whether to ignore whitespace changes */
  ignoreWhitespace?: boolean;
  /** Whether to ignore case changes */
  ignoreCase?: boolean;
  /** Whether to perform word-level diff instead of line-level */
  wordDiff?: boolean;
  /** Whether to include unchanged lines in the result */
  includeUnchanged?: boolean;
  /** Maximum file size to diff (in bytes, default: 10MB) */
  maxFileSize?: number;
}

/**
 * Formatted diff output options
 */
export interface FormatOptions {
  /** Format style (unified, side-by-side, etc.) */
  format?: 'unified' | 'side-by-side' | 'json';
  /** Number of context lines (default: 3) */
  context?: number;
  /** Whether to show line numbers */
  showLineNumbers?: boolean;
  /** Whether to colorize output (for terminal) */
  colorize?: boolean;
  /** Prefix for old lines (default: '-') */
  oldPrefix?: string;
  /** Prefix for new lines (default: '+') */
  newPrefix?: string;
}

/**
 * Word-level change for fine-grained diffs
 */
export interface WordChange {
  /** Type of change */
  type: ChangeType;
  /** The word or character sequence */
  value: string;
  /** Position in the line */
  position: number;
}

/**
 * Line with word-level changes
 */
export interface LineWithWordChanges {
  /** Original line number */
  lineNumber: number;
  /** Array of word-level changes */
  words: WordChange[];
  /** The complete line text */
  text: string;
}

/**
 * Binary file diff result
 */
export interface BinaryDiffResult {
  /** Indicates this is a binary file */
  isBinary: true;
  /** Old file size in bytes */
  oldSize: number;
  /** New file size in bytes */
  newSize: number;
  /** Whether sizes differ */
  sizeChanged: boolean;
  /** Optional message */
  message?: string;
}
