/**
 * Path manipulation utilities
 * Similar to Node.js path module
 */

/**
 * Path separator (always '/' for browser/Unix-style paths)
 */
export const sep = '/';

/**
 * Normalize a path, resolving '..' and '.' segments
 */
export function normalize(path: string): string {
  if (!path || path === '.') {
    return '.';
  }

  // Remove leading/trailing slashes and split
  const parts = path.replace(/^\/+|\/+$/g, '').split('/');
  const result: string[] = [];

  for (const part of parts) {
    if (!part || part === '.') {
      // Skip empty and current directory markers
      continue;
    } else if (part === '..') {
      // Go up one directory
      if (result.length > 0 && result[result.length - 1] !== '..') {
        result.pop();
      } else {
        result.push('..');
      }
    } else {
      result.push(part);
    }
  }

  if (result.length === 0) {
    return '.';
  }

  return result.join('/');
}

/**
 * Join path segments
 */
export function join(...segments: string[]): string {
  if (segments.length === 0) {
    return '.';
  }

  const joined = segments
    .filter((seg) => seg && seg.length > 0)
    .join('/');

  return normalize(joined);
}

/**
 * Get directory name of a path
 */
export function dirname(path: string): string {
  if (!path) {
    return '.';
  }

  const normalized = normalize(path);
  if (normalized === '.') {
    return '.';
  }

  const lastSlash = normalized.lastIndexOf('/');
  if (lastSlash === -1) {
    return '.';
  }

  if (lastSlash === 0) {
    return '/';
  }

  return normalized.substring(0, lastSlash);
}

/**
 * Get base name of a path
 */
export function basename(path: string, ext?: string): string {
  if (!path) {
    return '';
  }

  const normalized = normalize(path);
  const lastSlash = normalized.lastIndexOf('/');
  const base = lastSlash === -1 ? normalized : normalized.substring(lastSlash + 1);

  if (ext && base.endsWith(ext)) {
    return base.substring(0, base.length - ext.length);
  }

  return base;
}

/**
 * Get file extension
 */
export function extname(path: string): string {
  if (!path) {
    return '';
  }

  const base = basename(path);
  const lastDot = base.lastIndexOf('.');

  if (lastDot === -1 || lastDot === 0) {
    return '';
  }

  return base.substring(lastDot);
}

/**
 * Check if path is absolute
 */
export function isAbsolute(path: string): boolean {
  return path.startsWith('/');
}

/**
 * Resolve a sequence of paths to an absolute path
 */
export function resolve(...paths: string[]): string {
  let resolved = '';

  for (let i = paths.length - 1; i >= 0; i--) {
    const path = paths[i];
    if (!path) continue;

    if (isAbsolute(path)) {
      resolved = path;
      break;
    }

    resolved = path + (resolved ? '/' + resolved : '');
  }

  // If still relative, prepend current directory marker
  if (!isAbsolute(resolved)) {
    resolved = '/' + resolved;
  }

  return normalize(resolved);
}

/**
 * Get relative path from 'from' to 'to'
 */
export function relative(from: string, to: string): string {
  const fromParts = normalize(from).split('/');
  const toParts = normalize(to).split('/');

  // Find common prefix
  let commonLength = 0;
  const minLength = Math.min(fromParts.length, toParts.length);

  for (let i = 0; i < minLength; i++) {
    if (fromParts[i] !== toParts[i]) {
      break;
    }
    commonLength++;
  }

  // Build relative path
  const upCount = fromParts.length - commonLength;
  const relativeParts: string[] = [];

  for (let i = 0; i < upCount; i++) {
    relativeParts.push('..');
  }

  for (let i = commonLength; i < toParts.length; i++) {
    const part = toParts[i];
    if (part) {
      relativeParts.push(part);
    }
  }

  if (relativeParts.length === 0) {
    return '.';
  }

  return relativeParts.join('/');
}

/**
 * Parse a path into components
 */
export interface ParsedPath {
  /** Root of the path (always '' for relative paths) */
  root: string;
  /** Directory path */
  dir: string;
  /** Base name with extension */
  base: string;
  /** File extension including dot */
  ext: string;
  /** File name without extension */
  name: string;
}

export function parse(path: string): ParsedPath {
  const normalized = normalize(path);
  const isAbs = isAbsolute(path);

  const dir = dirname(normalized);
  const base = basename(normalized);
  const ext = extname(base);
  const name = ext ? base.substring(0, base.length - ext.length) : base;

  return {
    root: isAbs ? '/' : '',
    dir: dir === '.' ? '' : dir,
    base,
    ext,
    name,
  };
}

/**
 * Format path components into a path string
 */
export function format(pathObj: Partial<ParsedPath>): string {
  const dir = pathObj.dir || '';
  const base = pathObj.base || ((pathObj.name || '') + (pathObj.ext || ''));

  if (!dir) {
    return base;
  }

  return join(dir, base);
}
