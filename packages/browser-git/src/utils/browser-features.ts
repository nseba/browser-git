/**
 * Browser feature detection and capability checking
 * Provides graceful degradation for unsupported features
 */

export interface BrowserCapabilities {
  /**  Storage capabilities */
  indexedDB: boolean;
  opfs: boolean;
  localStorage: boolean;
  sessionStorage: boolean;

  /** WASM capabilities */
  webAssembly: boolean;
  wasmStreaming: boolean;
  wasmThreads: boolean;
  wasmSIMD: boolean;

  /** Compression capabilities */
  compressionStream: boolean;
  decompressionStream: boolean;

  /** Crypto capabilities */
  webCrypto: boolean;
  subtleCrypto: boolean;

  /** Worker capabilities */
  workers: boolean;
  sharedArrayBuffer: boolean;

  /** Storage quota */
  storageManager: boolean;
  persistentStorage: boolean;

  /** Performance APIs */
  performanceAPI: boolean;
  performanceMemory: boolean;
}

export interface StorageQuota {
  usage: number;
  quota: number;
  available: number;
  percentage: number;
}

export interface BrowserInfo {
  name: string;
  version: string;
  engine: string;
  platform: string;
  mobile: boolean;
}

/**
 * Detect all browser capabilities
 */
export async function detectBrowserCapabilities(): Promise<BrowserCapabilities> {
  return {
    // Storage
    indexedDB: checkIndexedDB(),
    opfs: await checkOPFS(),
    localStorage: checkLocalStorage(),
    sessionStorage: checkSessionStorage(),

    // WASM
    webAssembly: checkWebAssembly(),
    wasmStreaming: checkWasmStreaming(),
    wasmThreads: await checkWasmThreads(),
    wasmSIMD: await checkWasmSIMD(),

    // Compression
    compressionStream: checkCompressionStream(),
    decompressionStream: checkDecompressionStream(),

    // Crypto
    webCrypto: checkWebCrypto(),
    subtleCrypto: checkSubtleCrypto(),

    // Workers
    workers: checkWorkers(),
    sharedArrayBuffer: checkSharedArrayBuffer(),

    // Storage
    storageManager: checkStorageManager(),
    persistentStorage: await checkPersistentStorage(),

    // Performance
    performanceAPI: checkPerformanceAPI(),
    performanceMemory: checkPerformanceMemory(),
  };
}

/**
 * Check IndexedDB availability
 */
export function checkIndexedDB(): boolean {
  try {
    return typeof indexedDB !== 'undefined' && indexedDB !== null;
  } catch {
    return false;
  }
}

/**
 * Check OPFS (Origin Private File System) availability
 */
export async function checkOPFS(): Promise<boolean> {
  try {
    if (
      typeof navigator !== 'undefined' &&
      'storage' in navigator &&
      'getDirectory' in (navigator.storage as any)
    ) {
      // Try to actually access OPFS
      const root = await (navigator.storage as any).getDirectory();
      return root !== null;
    }
    return false;
  } catch {
    return false;
  }
}

/**
 * Check localStorage availability
 */
export function checkLocalStorage(): boolean {
  try {
    if (typeof localStorage === 'undefined') {
      return false;
    }
    // Try to actually use it
    const testKey = '__storage_test__';
    localStorage.setItem(testKey, 'test');
    localStorage.removeItem(testKey);
    return true;
  } catch {
    return false;
  }
}

/**
 * Check sessionStorage availability
 */
export function checkSessionStorage(): boolean {
  try {
    if (typeof sessionStorage === 'undefined') {
      return false;
    }
    // Try to actually use it
    const testKey = '__storage_test__';
    sessionStorage.setItem(testKey, 'test');
    sessionStorage.removeItem(testKey);
    return true;
  } catch {
    return false;
  }
}

/**
 * Check WebAssembly availability
 */
export function checkWebAssembly(): boolean {
  try {
    return typeof WebAssembly !== 'undefined' && typeof WebAssembly.instantiate === 'function';
  } catch {
    return false;
  }
}

/**
 * Check WASM streaming compilation
 */
export function checkWasmStreaming(): boolean {
  try {
    return (
      typeof WebAssembly !== 'undefined' &&
      typeof WebAssembly.instantiateStreaming === 'function' &&
      typeof WebAssembly.compileStreaming === 'function'
    );
  } catch {
    return false;
  }
}

/**
 * Check WASM threads support
 */
export async function checkWasmThreads(): Promise<boolean> {
  try {
    // Threads require SharedArrayBuffer
    if (typeof SharedArrayBuffer === 'undefined') {
      return false;
    }

    // Try to create a WASM module with threads
    // This is a simple test module with shared memory
    const source = new Uint8Array([
      0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x05, 0x04, 0x01, 0x03, 0x01, 0x01,
    ]);

    const module = await WebAssembly.instantiate(source);
    return module !== null;
  } catch {
    return false;
  }
}

/**
 * Check WASM SIMD support
 */
export async function checkWasmSIMD(): Promise<boolean> {
  try {
    // Try to compile a module with SIMD instructions
    const source = new Uint8Array([
      0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
    ]);

    await WebAssembly.instantiate(source);
    return true;
  } catch {
    return false;
  }
}

/**
 * Check CompressionStream API
 */
export function checkCompressionStream(): boolean {
  try {
    return typeof CompressionStream !== 'undefined';
  } catch {
    return false;
  }
}

/**
 * Check DecompressionStream API
 */
export function checkDecompressionStream(): boolean {
  try {
    return typeof DecompressionStream !== 'undefined';
  } catch {
    return false;
  }
}

/**
 * Check Web Crypto API
 */
export function checkWebCrypto(): boolean {
  try {
    return typeof crypto !== 'undefined' && crypto !== null;
  } catch {
    return false;
  }
}

/**
 * Check SubtleCrypto API
 */
export function checkSubtleCrypto(): boolean {
  try {
    return (
      typeof crypto !== 'undefined' &&
      crypto !== null &&
      'subtle' in crypto &&
      crypto.subtle !== null
    );
  } catch {
    return false;
  }
}

/**
 * Check Web Workers support
 */
export function checkWorkers(): boolean {
  try {
    return typeof Worker !== 'undefined';
  } catch {
    return false;
  }
}

/**
 * Check SharedArrayBuffer support
 */
export function checkSharedArrayBuffer(): boolean {
  try {
    return typeof SharedArrayBuffer !== 'undefined';
  } catch {
    return false;
  }
}

/**
 * Check Storage Manager API
 */
export function checkStorageManager(): boolean {
  try {
    return typeof navigator !== 'undefined' && 'storage' in navigator;
  } catch {
    return false;
  }
}

/**
 * Check persistent storage capability
 */
export async function checkPersistentStorage(): Promise<boolean> {
  try {
    if (!checkStorageManager()) {
      return false;
    }

    const persisted = await navigator.storage.persisted();
    return persisted;
  } catch {
    return false;
  }
}

/**
 * Check Performance API
 */
export function checkPerformanceAPI(): boolean {
  try {
    return typeof performance !== 'undefined' && typeof performance.now === 'function';
  } catch {
    return false;
  }
}

/**
 * Check Performance Memory API (Chrome-specific)
 */
export function checkPerformanceMemory(): boolean {
  try {
    return typeof performance !== 'undefined' && 'memory' in performance;
  } catch {
    return false;
  }
}

/**
 * Get storage quota information
 */
export async function getStorageQuota(): Promise<StorageQuota | null> {
  try {
    if (!checkStorageManager() || !navigator.storage.estimate) {
      return null;
    }

    const estimate = await navigator.storage.estimate();
    const usage = estimate.usage || 0;
    const quota = estimate.quota || 0;
    const available = quota - usage;
    const percentage = quota > 0 ? (usage / quota) * 100 : 0;

    return {
      usage,
      quota,
      available,
      percentage,
    };
  } catch {
    return null;
  }
}

/**
 * Request persistent storage
 */
export async function requestPersistentStorage(): Promise<boolean> {
  try {
    if (!checkStorageManager() || !navigator.storage.persist) {
      return false;
    }

    const granted = await navigator.storage.persist();
    return granted;
  } catch {
    return false;
  }
}

/**
 * Detect browser information
 */
export function detectBrowser(): BrowserInfo {
  const ua = typeof navigator !== 'undefined' ? navigator.userAgent : '';
  const platform = typeof navigator !== 'undefined' ? navigator.platform : '';

  // Detect browser name and version
  let name = 'Unknown';
  let version = 'Unknown';
  let engine = 'Unknown';

  // Chrome/Chromium
  if (ua.includes('Chrome') && !ua.includes('Edg')) {
    name = 'Chrome';
    const match = ua.match(/Chrome\/(\d+)/);
    if (match) version = match[1];
    engine = 'Blink';
  }
  // Edge
  else if (ua.includes('Edg/')) {
    name = 'Edge';
    const match = ua.match(/Edg\/(\d+)/);
    if (match) version = match[1];
    engine = 'Blink';
  }
  // Firefox
  else if (ua.includes('Firefox')) {
    name = 'Firefox';
    const match = ua.match(/Firefox\/(\d+)/);
    if (match) version = match[1];
    engine = 'Gecko';
  }
  // Safari
  else if (ua.includes('Safari') && !ua.includes('Chrome')) {
    name = 'Safari';
    const match = ua.match(/Version\/(\d+)/);
    if (match) version = match[1];
    engine = 'WebKit';
  }

  // Detect mobile
  const mobile =
    /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(ua) ||
    platform.includes('Mobile');

  return {
    name,
    version,
    engine,
    platform,
    mobile,
  };
}

/**
 * Get recommended storage adapter based on browser capabilities
 */
export async function getRecommendedStorageAdapter(
  capabilities?: BrowserCapabilities
): Promise<'opfs' | 'indexeddb' | 'localstorage' | 'memory'> {
  const caps = capabilities || (await detectBrowserCapabilities());

  // Prefer OPFS for best performance
  if (caps.opfs) {
    return 'opfs';
  }

  // IndexedDB for structured data
  if (caps.indexedDB) {
    return 'indexeddb';
  }

  // LocalStorage as fallback (limited capacity)
  if (caps.localStorage) {
    return 'localstorage';
  }

  // Memory only (no persistence)
  return 'memory';
}

/**
 * Check if browser meets minimum requirements for browser-git
 */
export async function checkMinimumRequirements(): Promise<{
  met: boolean;
  missing: string[];
}> {
  const capabilities = await detectBrowserCapabilities();
  const missing: string[] = [];

  // Required features
  if (!capabilities.webAssembly) {
    missing.push('WebAssembly');
  }

  if (!capabilities.subtleCrypto) {
    missing.push('SubtleCrypto (for SHA-1 hashing)');
  }

  // At least one storage option
  if (
    !capabilities.indexedDB &&
    !capabilities.opfs &&
    !capabilities.localStorage
  ) {
    missing.push('Storage (IndexedDB, OPFS, or localStorage)');
  }

  return {
    met: missing.length === 0,
    missing,
  };
}

/**
 * Format bytes to human-readable string
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

/**
 * Format storage quota for display
 */
export function formatStorageQuota(quota: StorageQuota): string {
  return `${formatBytes(quota.usage)} / ${formatBytes(quota.quota)} (${quota.percentage.toFixed(1)}% used)`;
}
