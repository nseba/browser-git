import * as Diff from 'diff';
import type { IDiffEngine } from './interface.js';
import type {
  DiffResult,
  DiffOptions,
  BinaryDiffResult,
  FormatOptions,
  LineWithWordChanges,
  DiffHunk,
  Change,
  WordChange,
} from './types.js';
import { ChangeType } from './types.js';
import {
  splitLines,
  normalizeWhitespace,
} from './utils/text.js';
import {
  isBinary,
  uint8ArrayToString,
  formatSize,
} from './utils/binary.js';

/**
 * Myers diff algorithm implementation using the `diff` library
 */
export class MyersDiffEngine implements IDiffEngine {
  /**
   * Compute diff between two text strings
   */
  diff(
    oldText: string,
    newText: string,
    options: DiffOptions = {}
  ): DiffResult | BinaryDiffResult {
    // Apply whitespace normalization if requested
    let oldProcessed = oldText;
    let newProcessed = newText;

    if (options.ignoreWhitespace) {
      oldProcessed = normalizeWhitespace(oldProcessed, {
        trimLines: true,
        ignoreTrailingWhitespace: true,
      });
      newProcessed = normalizeWhitespace(newProcessed, {
        trimLines: true,
        ignoreTrailingWhitespace: true,
      });
    }

    // Perform line-level diff
    const diffResult = Diff.diffLines(oldProcessed, newProcessed, {
      ignoreWhitespace: options.ignoreWhitespace,
      ignoreCase: options.ignoreCase,
    });

    // Convert to our format
    const hunks = this.convertToHunks(diffResult, options);

    // Calculate statistics
    let linesAdded = 0;
    let linesDeleted = 0;

    for (const hunk of hunks) {
      for (const change of hunk.changes) {
        if (change.type === ChangeType.Add) {
          linesAdded += change.count || 1;
        } else if (change.type === ChangeType.Delete) {
          linesDeleted += change.count || 1;
        }
      }
    }

    return {
      hunks,
      isBinary: false,
      linesAdded,
      linesDeleted,
    };
  }

  /**
   * Compute diff between two files (binary-safe)
   */
  diffFiles(
    oldFile: Uint8Array,
    newFile: Uint8Array,
    options: DiffOptions = {}
  ): DiffResult | BinaryDiffResult {
    // Check for binary content
    if (isBinary(oldFile) || isBinary(newFile)) {
      return {
        isBinary: true,
        oldSize: oldFile.length,
        newSize: newFile.length,
        sizeChanged: oldFile.length !== newFile.length,
        message: `Binary files differ (${formatSize(oldFile.length)} vs ${formatSize(newFile.length)})`,
      };
    }

    // Convert to text and diff
    const oldText = uint8ArrayToString(oldFile);
    const newText = uint8ArrayToString(newFile);

    return this.diff(oldText, newText, options);
  }

  /**
   * Compute word-level diff for a single line
   */
  diffWords(
    oldLine: string,
    newLine: string,
    options: DiffOptions = {}
  ): LineWithWordChanges {
    const diffResult = Diff.diffWords(oldLine, newLine, {
      ignoreCase: options.ignoreCase,
    });

    const words: WordChange[] = [];
    let position = 0;

    for (const part of diffResult) {
      const type = part.added
        ? ChangeType.Add
        : part.removed
        ? ChangeType.Delete
        : ChangeType.Equal;

      words.push({
        type,
        value: part.value,
        position,
      });

      position += part.value.length;
    }

    return {
      lineNumber: 1,
      words,
      text: oldLine,
    };
  }

  /**
   * Format diff result as a string
   */
  format(
    diff: DiffResult | BinaryDiffResult,
    options: FormatOptions = {}
  ): string {
    if (diff.isBinary) {
      return this.formatBinaryDiff(diff as BinaryDiffResult);
    }

    const format = options.format || 'unified';

    switch (format) {
      case 'unified':
        return this.formatUnified(diff as DiffResult, options);
      case 'side-by-side':
        return this.formatSideBySide(diff as DiffResult, options);
      case 'json':
        return JSON.stringify(diff, null, 2);
      default:
        return this.formatUnified(diff as DiffResult, options);
    }
  }

  /**
   * Check if content is binary
   */
  isBinary(content: Uint8Array): boolean {
    return isBinary(content);
  }

  /**
   * Apply a patch to text
   */
  patch(oldText: string, diff: DiffResult): string | null {
    try {
      const lines = splitLines(oldText);
      const result: string[] = [];
      let lineIndex = 0;

      for (const hunk of diff.hunks) {
        // Add lines before the hunk
        while (lineIndex < hunk.oldStart - 1 && lineIndex < lines.length) {
          const line = lines[lineIndex];
          if (line !== undefined) {
            result.push(line);
          }
          lineIndex++;
        }

        // Apply changes from the hunk
        for (const change of hunk.changes) {
          if (change.type === ChangeType.Equal) {
            // Keep unchanged lines
            const count = change.count || 1;
            for (let i = 0; i < count; i++) {
              const line = lines[lineIndex];
              if (line !== undefined) {
                result.push(line);
              }
              lineIndex++;
            }
          } else if (change.type === ChangeType.Delete) {
            // Skip deleted lines
            const count = change.count || 1;
            lineIndex += count;
          } else if (change.type === ChangeType.Add) {
            // Add new lines
            result.push(change.value);
          }
        }
      }

      // Add remaining lines
      while (lineIndex < lines.length) {
        const line = lines[lineIndex];
        if (line !== undefined) {
          result.push(line);
        }
        lineIndex++;
      }

      // Preserve trailing newline from original text
      const patchedText = result.join('\n');
      return oldText.endsWith('\n') ? patchedText + '\n' : patchedText;
    } catch (error) {
      return null;
    }
  }

  /**
   * Convert diff library output to our hunk format
   */
  private convertToHunks(
    diffResult: Diff.Change[],
    options: DiffOptions
  ): DiffHunk[] {
    const context = options.context ?? 3;
    const hunks: DiffHunk[] = [];
    let currentHunk: DiffHunk | null = null;
    let oldLine = 1;
    let newLine = 1;

    for (const part of diffResult) {
      const lines = splitLines(part.value);
      const type = part.added
        ? ChangeType.Add
        : part.removed
        ? ChangeType.Delete
        : ChangeType.Equal;

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];

        // Skip empty last line from split
        if (i === lines.length - 1 && line === '') {
          continue;
        }

        if (line === undefined) {
          continue;
        }

        // Create change
        const change: Change = {
          type,
          value: line,
          count: 1,
          ...(type !== ChangeType.Add && { oldLineNumber: oldLine }),
          ...(type !== ChangeType.Delete && { newLineNumber: newLine }),
        };

        // Start new hunk if needed
        if (!currentHunk || type !== ChangeType.Equal) {
          if (!currentHunk) {
            currentHunk = {
              oldStart: oldLine,
              oldLines: 0,
              newStart: newLine,
              newLines: 0,
              changes: [],
            };
            hunks.push(currentHunk);
          }

          currentHunk.changes.push(change);

          if (type === ChangeType.Delete || type === ChangeType.Equal) {
            currentHunk.oldLines++;
          }
          if (type === ChangeType.Add || type === ChangeType.Equal) {
            currentHunk.newLines++;
          }
        } else {
          // Context line - add to current hunk or start new one
          if (currentHunk.changes.length > 0) {
            const lastChange = currentHunk.changes[currentHunk.changes.length - 1];
            if (lastChange && lastChange.type === ChangeType.Equal && (lastChange.count ?? 1) < context * 2) {
              // Add to existing context
              lastChange.count = (lastChange.count ?? 1) + 1;
              currentHunk.oldLines++;
              currentHunk.newLines++;
            } else {
              // Close current hunk
              currentHunk = null;
              continue;
            }
          }
        }

        // Update line numbers
        if (type !== ChangeType.Add) {
          oldLine++;
        }
        if (type !== ChangeType.Delete) {
          newLine++;
        }
      }
    }

    return hunks;
  }

  /**
   * Format unified diff
   */
  private formatUnified(diff: DiffResult, options: FormatOptions): string {
    const lines: string[] = [];
    const oldPrefix = options.oldPrefix || '-';
    const newPrefix = options.newPrefix || '+';

    if (diff.oldPath && diff.newPath) {
      lines.push(`--- ${diff.oldPath}`);
      lines.push(`+++ ${diff.newPath}`);
    }

    for (const hunk of diff.hunks) {
      lines.push(
        `@@ -${hunk.oldStart},${hunk.oldLines} +${hunk.newStart},${hunk.newLines} @@`
      );

      for (const change of hunk.changes) {
        const prefix =
          change.type === ChangeType.Add
            ? newPrefix
            : change.type === ChangeType.Delete
            ? oldPrefix
            : ' ';

        const value = change.value ?? '';
        lines.push(`${prefix}${value}`);
      }
    }

    return lines.join('\n');
  }

  /**
   * Format side-by-side diff
   */
  private formatSideBySide(diff: DiffResult, _options: FormatOptions): string {
    const lines: string[] = [];
    const width = 80;

    for (const hunk of diff.hunks) {
      lines.push(`@@ -${hunk.oldStart},${hunk.oldLines} +${hunk.newStart},${hunk.newLines} @@`);
      lines.push('');

      for (const change of hunk.changes) {
        const line = change.value ? change.value.padEnd(width) : '';
        const indicator =
          change.type === ChangeType.Add
            ? '+'
            : change.type === ChangeType.Delete
            ? '-'
            : ' ';

        lines.push(`${indicator} ${line}`);
      }

      lines.push('');
    }

    return lines.join('\n');
  }

  /**
   * Format binary diff
   */
  private formatBinaryDiff(diff: BinaryDiffResult): string {
    return (
      diff.message ||
      `Binary files differ (${formatSize(diff.oldSize)} vs ${formatSize(diff.newSize)})`
    );
  }
}
