/**
 * Common test utilities shared across all test types
 */

/**
 * Delays execution for a specified number of milliseconds
 */
export function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Retries a function until it succeeds or max attempts reached
 */
export async function retry<T>(
  fn: () => Promise<T>,
  maxAttempts: number = 3,
  delayMs: number = 100
): Promise<T> {
  let lastError: Error | undefined;

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error as Error;
      if (attempt < maxAttempts) {
        await delay(delayMs);
      }
    }
  }

  throw lastError;
}

/**
 * Generates random data for testing
 */
export function randomBytes(size: number): Uint8Array {
  const bytes = new Uint8Array(size);
  for (let i = 0; i < size; i++) {
    bytes[i] = Math.floor(Math.random() * 256);
  }
  return bytes;
}

/**
 * Generates a random string
 */
export function randomString(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

/**
 * Generates a random file name
 */
export function randomFileName(extension: string = 'txt'): string {
  return `test_${randomString(8)}.${extension}`;
}

/**
 * Compares two Uint8Arrays for equality
 */
export function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
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

/**
 * Converts string to Uint8Array
 */
export function stringToBytes(str: string): Uint8Array {
  const encoder = new TextEncoder();
  return encoder.encode(str);
}

/**
 * Converts Uint8Array to string
 */
export function bytesToString(bytes: Uint8Array): string {
  const decoder = new TextDecoder();
  return decoder.decode(bytes);
}

/**
 * Creates a deep copy of an object
 */
export function deepCopy<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj));
}

/**
 * Checks if two objects are deeply equal
 */
export function deepEqual(a: any, b: any): boolean {
  if (a === b) return true;

  if (typeof a !== 'object' || typeof b !== 'object' || a === null || b === null) {
    return false;
  }

  const keysA = Object.keys(a);
  const keysB = Object.keys(b);

  if (keysA.length !== keysB.length) {
    return false;
  }

  for (const key of keysA) {
    if (!keysB.includes(key)) {
      return false;
    }

    if (!deepEqual(a[key], b[key])) {
      return false;
    }
  }

  return true;
}

/**
 * Formats bytes as human-readable string
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${units[i]}`;
}

/**
 * Formats milliseconds as human-readable string
 */
export function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms.toFixed(2)}ms`;
  } else if (ms < 60000) {
    return `${(ms / 1000).toFixed(2)}s`;
  } else {
    const minutes = Math.floor(ms / 60000);
    const seconds = ((ms % 60000) / 1000).toFixed(2);
    return `${minutes}m ${seconds}s`;
  }
}

/**
 * Creates a mock timer for testing time-dependent code
 */
export class MockTimer {
  private currentTime: number = 0;
  private timers: Map<number, { callback: () => void; time: number }> = new Map();
  private nextId: number = 1;

  now(): number {
    return this.currentTime;
  }

  setTimeout(callback: () => void, delay: number): number {
    const id = this.nextId++;
    this.timers.set(id, { callback, time: this.currentTime + delay });
    return id;
  }

  clearTimeout(id: number): void {
    this.timers.delete(id);
  }

  tick(ms: number): void {
    this.currentTime += ms;
    const expired = Array.from(this.timers.entries())
      .filter(([_, timer]) => timer.time <= this.currentTime)
      .sort((a, b) => a[1].time - b[1].time);

    for (const [id, timer] of expired) {
      this.timers.delete(id);
      timer.callback();
    }
  }

  reset(): void {
    this.currentTime = 0;
    this.timers.clear();
    this.nextId = 1;
  }
}

/**
 * Performance measurement utility
 */
export class PerformanceMeasure {
  private marks: Map<string, number> = new Map();

  mark(name: string): void {
    this.marks.set(name, performance.now());
  }

  measure(startMark: string, endMark?: string): number {
    const start = this.marks.get(startMark);
    if (start === undefined) {
      throw new Error(`Mark "${startMark}" not found`);
    }

    const end = endMark ? this.marks.get(endMark) : performance.now();
    if (end === undefined) {
      throw new Error(`Mark "${endMark}" not found`);
    }

    return end - start;
  }

  clear(): void {
    this.marks.clear();
  }
}
