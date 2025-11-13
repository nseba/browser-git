/**
 * Binary file detection utilities
 */

/**
 * Check if content appears to be binary
 *
 * Uses heuristics similar to Git's binary detection:
 * - Checks for NUL bytes
 * - Checks for high ratio of non-printable characters
 *
 * @param content - Content to check
 * @param sampleSize - Number of bytes to sample (default: 8000)
 * @returns True if content appears to be binary
 */
export function isBinary(content: Uint8Array, sampleSize = 8000): boolean {
  if (content.length === 0) {
    return false;
  }

  // Sample the beginning of the file
  const sample = content.slice(0, Math.min(sampleSize, content.length));

  // Check for NUL byte (strong indicator of binary)
  for (let i = 0; i < sample.length; i++) {
    if (sample[i] === 0) {
      return true;
    }
  }

  // Count non-printable characters
  let nonPrintable = 0;
  for (let i = 0; i < sample.length; i++) {
    const byte = sample[i];
    if (byte === undefined) continue;
    // Non-printable: not tab (9), not newline (10), not carriage return (13),
    // and not in printable ASCII range (32-126)
    if (
      byte !== 9 &&
      byte !== 10 &&
      byte !== 13 &&
      (byte < 32 || byte > 126)
    ) {
      nonPrintable++;
    }
  }

  // If more than 30% of characters are non-printable, consider it binary
  const ratio = nonPrintable / sample.length;
  return ratio > 0.3;
}

/**
 * Check if content is text
 *
 * @param content - Content to check
 * @param sampleSize - Number of bytes to sample
 * @returns True if content appears to be text
 */
export function isText(content: Uint8Array, sampleSize = 8000): boolean {
  return !isBinary(content, sampleSize);
}

/**
 * Convert Uint8Array to string assuming UTF-8 encoding
 *
 * @param content - Binary content
 * @returns Decoded string
 */
export function uint8ArrayToString(content: Uint8Array): string {
  const decoder = new TextDecoder('utf-8', { fatal: false });
  return decoder.decode(content);
}

/**
 * Convert string to Uint8Array using UTF-8 encoding
 *
 * @param text - Text to encode
 * @returns Encoded binary content
 */
export function stringToUint8Array(text: string): Uint8Array {
  const encoder = new TextEncoder();
  return encoder.encode(text);
}

/**
 * Detect file type from content and file extension
 *
 * @param content - File content
 * @param _filename - Optional filename for extension-based detection (currently unused)
 * @returns Detected file type
 */
export function detectFileType(
  content: Uint8Array,
  _filename?: string
): 'text' | 'binary' | 'image' | 'unknown' {
  // Check by content first
  if (isBinary(content)) {
    // Check for common image signatures
    if (isImage(content)) {
      return 'image';
    }
    return 'binary';
  }

  // It's text
  return 'text';
}

/**
 * Check if content is an image file
 *
 * @param content - File content
 * @returns True if content is an image
 */
export function isImage(content: Uint8Array): boolean {
  if (content.length < 4) {
    return false;
  }

  // Check for common image file signatures
  const signatures = [
    // PNG
    [0x89, 0x50, 0x4e, 0x47],
    // JPEG
    [0xff, 0xd8, 0xff],
    // GIF87a
    [0x47, 0x49, 0x46, 0x38, 0x37, 0x61],
    // GIF89a
    [0x47, 0x49, 0x46, 0x38, 0x39, 0x61],
    // BMP
    [0x42, 0x4d],
    // WebP
    [0x52, 0x49, 0x46, 0x46],
    // ICO
    [0x00, 0x00, 0x01, 0x00],
  ];

  for (const signature of signatures) {
    if (matchesSignature(content, signature)) {
      return true;
    }
  }

  return false;
}

/**
 * Check if content starts with a specific byte signature
 *
 * @param content - Content to check
 * @param signature - Byte signature to match
 * @returns True if signature matches
 */
function matchesSignature(content: Uint8Array, signature: number[]): boolean {
  if (content.length < signature.length) {
    return false;
  }

  for (let i = 0; i < signature.length; i++) {
    if (content[i] !== signature[i]) {
      return false;
    }
  }

  return true;
}

/**
 * Get a human-readable size string
 *
 * @param bytes - Number of bytes
 * @returns Formatted size string (e.g., "1.5 KB")
 */
export function formatSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(unitIndex > 0 ? 1 : 0)} ${units[unitIndex]}`;
}
