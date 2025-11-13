/**
 * @browser-git/browser-git
 *
 * Browser-based Git implementation with file system API
 */

export * from './filesystem/index.js';
export * from './types/index.js';

// Re-export diff-engine for convenient access
export {
  MyersDiffEngine,
  DiffEngineFactory,
  diffEngineFactory,
  ChangeType,
  type IDiffEngine,
  type IDiffEngineFactory,
  type DiffResult,
  type BinaryDiffResult,
  type DiffOptions,
  type FormatOptions,
  type DiffHunk,
  type Change,
  type WordChange,
  type LineWithWordChanges,
} from '@browser-git/diff-engine';
