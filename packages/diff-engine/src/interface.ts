import type {
  DiffResult,
  DiffOptions,
  BinaryDiffResult,
  FormatOptions,
  LineWithWordChanges,
} from './types.js';

/**
 * Interface for diff engine implementations
 *
 * This allows plugging in different diff algorithms while maintaining
 * a consistent API for consumers.
 */
export interface IDiffEngine {
  /**
   * Compute diff between two text strings
   *
   * @param oldText - Original text
   * @param newText - Modified text
   * @param options - Diff options
   * @returns Diff result with hunks and changes
   */
  diff(
    oldText: string,
    newText: string,
    options?: DiffOptions
  ): DiffResult | BinaryDiffResult;

  /**
   * Compute diff between two files (as Uint8Array for binary support)
   *
   * @param oldFile - Original file content
   * @param newFile - Modified file content
   * @param options - Diff options
   * @returns Diff result or binary diff result
   */
  diffFiles(
    oldFile: Uint8Array,
    newFile: Uint8Array,
    options?: DiffOptions
  ): DiffResult | BinaryDiffResult;

  /**
   * Compute word-level diff for a single line
   *
   * @param oldLine - Original line
   * @param newLine - Modified line
   * @param options - Diff options
   * @returns Line with word-level changes
   */
  diffWords(
    oldLine: string,
    newLine: string,
    options?: DiffOptions
  ): LineWithWordChanges;

  /**
   * Format diff result as a string
   *
   * @param diff - Diff result to format
   * @param options - Format options
   * @returns Formatted diff string (unified, side-by-side, etc.)
   */
  format(
    diff: DiffResult | BinaryDiffResult,
    options?: FormatOptions
  ): string;

  /**
   * Check if content is binary
   *
   * @param content - Content to check
   * @returns True if content appears to be binary
   */
  isBinary(content: Uint8Array): boolean;

  /**
   * Apply a patch to text
   *
   * @param oldText - Original text
   * @param diff - Diff to apply
   * @returns Patched text or null if patch cannot be applied
   */
  patch(oldText: string, diff: DiffResult): string | null;
}

/**
 * Factory for creating diff engine instances
 */
export interface IDiffEngineFactory {
  /**
   * Create a new diff engine instance
   *
   * @param name - Name of the diff algorithm (e.g., 'myers', 'patience')
   * @returns Diff engine instance
   */
  create(name?: string): IDiffEngine;

  /**
   * Register a custom diff engine implementation
   *
   * @param name - Name to register the engine under
   * @param engine - Diff engine instance or constructor
   */
  register(
    name: string,
    engine: IDiffEngine | (() => IDiffEngine)
  ): void;

  /**
   * List available diff engine names
   *
   * @returns Array of registered engine names
   */
  listEngines(): string[];
}
