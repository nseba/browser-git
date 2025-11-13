/**
 * Text processing utilities for diff operations
 */

/**
 * Split text into lines, preserving line endings
 *
 * @param text - Text to split
 * @param preserveLineEndings - Whether to keep \n at end of lines (default: false)
 * @returns Array of lines
 */
export function splitLines(text: string, preserveLineEndings = false): string[] {
  if (!text) {
    return [];
  }

  const lines = text.split(/\r?\n/);

  if (preserveLineEndings && text.endsWith('\n')) {
    // Add empty line if text ends with newline
    return lines;
  }

  // Remove last empty line if it exists (from trailing newline)
  if (lines.length > 0 && lines[lines.length - 1] === '') {
    lines.pop();
  }

  return lines;
}

/**
 * Join lines back into text with consistent line endings
 *
 * @param lines - Array of lines
 * @param lineEnding - Line ending to use (default: \n)
 * @returns Joined text
 */
export function joinLines(lines: string[], lineEnding = '\n'): string {
  return lines.join(lineEnding);
}

/**
 * Normalize whitespace in text for comparison
 *
 * @param text - Text to normalize
 * @param options - Normalization options
 * @returns Normalized text
 */
export function normalizeWhitespace(
  text: string,
  options: {
    trimLines?: boolean;
    collapseSpaces?: boolean;
    ignoreTrailingWhitespace?: boolean;
  } = {}
): string {
  let result = text;

  if (options.trimLines) {
    const lines = splitLines(result);
    result = joinLines(lines.map((line) => line.trim()));
  }

  if (options.collapseSpaces) {
    result = result.replace(/[ \t]+/g, ' ');
  }

  if (options.ignoreTrailingWhitespace) {
    const lines = splitLines(result);
    result = joinLines(lines.map((line) => line.trimEnd()));
  }

  return result;
}

/**
 * Split text into words for word-level diffing
 *
 * @param text - Text to split
 * @returns Array of words with their positions
 */
export function splitWords(text: string): Array<{ word: string; position: number }> {
  const words: Array<{ word: string; position: number }> = [];
  const regex = /\S+|\s+/g;
  let match: RegExpExecArray | null;

  while ((match = regex.exec(text)) !== null) {
    words.push({
      word: match[0],
      position: match.index,
    });
  }

  return words;
}

/**
 * Count leading whitespace characters
 *
 * @param line - Line to analyze
 * @returns Number of leading whitespace characters
 */
export function countLeadingWhitespace(line: string): number {
  const match = line.match(/^[ \t]*/);
  return match ? match[0].length : 0;
}

/**
 * Count trailing whitespace characters
 *
 * @param line - Line to analyze
 * @returns Number of trailing whitespace characters
 */
export function countTrailingWhitespace(line: string): number {
  const match = line.match(/[ \t]*$/);
  return match ? match[0].length : 0;
}

/**
 * Check if a line is blank (only whitespace)
 *
 * @param line - Line to check
 * @returns True if line is blank
 */
export function isBlankLine(line: string): boolean {
  return /^\s*$/.test(line);
}

/**
 * Get line ending type from text
 *
 * @param text - Text to analyze
 * @returns Line ending type (\n, \r\n, or \r)
 */
export function detectLineEnding(text: string): '\n' | '\r\n' | '\r' | null {
  if (text.includes('\r\n')) {
    return '\r\n';
  }
  if (text.includes('\n')) {
    return '\n';
  }
  if (text.includes('\r')) {
    return '\r';
  }
  return null;
}

/**
 * Normalize line endings to a consistent format
 *
 * @param text - Text to normalize
 * @param lineEnding - Target line ending (default: \n)
 * @returns Text with normalized line endings
 */
export function normalizeLineEndings(text: string, lineEnding: '\n' | '\r\n' | '\r' = '\n'): string {
  return text.replace(/\r\n|\r|\n/g, lineEnding);
}

/**
 * Strip common leading whitespace from all lines
 *
 * @param text - Text to process
 * @returns Text with common indentation removed
 */
export function stripCommonIndent(text: string): string {
  const lines = splitLines(text);

  // Find minimum indentation (excluding blank lines)
  const nonBlankLines = lines.filter((line) => !isBlankLine(line));
  if (nonBlankLines.length === 0) {
    return text;
  }

  const minIndent = Math.min(
    ...nonBlankLines.map((line) => countLeadingWhitespace(line))
  );

  if (minIndent === 0) {
    return text;
  }

  // Remove common indentation
  const stripped = lines.map((line) =>
    isBlankLine(line) ? line : line.slice(minIndent)
  );

  return joinLines(stripped);
}

/**
 * Add indentation to all lines
 *
 * @param text - Text to indent
 * @param indent - Indentation string (default: two spaces)
 * @returns Indented text
 */
export function addIndent(text: string, indent = '  '): string {
  const lines = splitLines(text);
  return joinLines(lines.map((line) => (isBlankLine(line) ? line : indent + line)));
}

/**
 * Escape special characters for display
 *
 * @param text - Text to escape
 * @returns Escaped text
 */
export function escapeForDisplay(text: string): string {
  return text
    .replace(/\t/g, '→   ') // Tab
    .replace(/\r/g, '\\r') // Carriage return
    .replace(/ /g, '·'); // Space (optional, for visibility)
}

/**
 * Truncate long lines for display
 *
 * @param text - Text to truncate
 * @param maxLength - Maximum line length (default: 120)
 * @returns Truncated text
 */
export function truncateLongLines(text: string, maxLength = 120): string {
  const lines = splitLines(text);
  const truncated = lines.map((line) => {
    if (line.length > maxLength) {
      return line.slice(0, maxLength) + '...';
    }
    return line;
  });
  return joinLines(truncated);
}
