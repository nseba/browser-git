/**
 * Utility functions for data serialization.
 */

/**
 * Convert a string to Uint8Array using UTF-8 encoding.
 */
export function stringToBytes(str: string): Uint8Array {
  const encoder = new TextEncoder();
  return encoder.encode(str);
}

/**
 * Convert Uint8Array to string using UTF-8 decoding.
 */
export function bytesToString(bytes: Uint8Array): string {
  const decoder = new TextDecoder();
  return decoder.decode(bytes);
}

/**
 * Serialize JSON object to Uint8Array.
 */
export function jsonToBytes(obj: unknown): Uint8Array {
  const json = JSON.stringify(obj);
  return stringToBytes(json);
}

/**
 * Deserialize Uint8Array to JSON object.
 */
export function bytesToJson<T = unknown>(bytes: Uint8Array): T {
  const json = bytesToString(bytes);
  return JSON.parse(json) as T;
}

/**
 * Concatenate multiple Uint8Arrays.
 */
export function concatBytes(...arrays: Uint8Array[]): Uint8Array {
  const totalLength = arrays.reduce((sum, arr) => sum + arr.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;

  for (const arr of arrays) {
    result.set(arr, offset);
    offset += arr.length;
  }

  return result;
}

/**
 * Compare two Uint8Arrays for equality.
 */
export function areEqual(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) {
    return false;
  }

  for (let i = 0; i < a.length; i++) {
    if (a[i] !== b[i]) {
      return false;
    }
  }

  return true;
}
