/**
 * Browser feature detection and compatibility utilities
 *
 * Detects browser capabilities and provides graceful fallbacks
 */

export interface BrowserCapabilities {
  /** Browser name and version */
  browser: {
    name: string;
    version: string;
    engine: string;
  };

  /** Core Web APIs */
  webAssembly: boolean;
  indexedDB: boolean;
  localStorage: boolean;
  sessionStorage: boolean;

  /** Advanced Storage */
  opfs: boolean; // Origin Private File System
  storageManager: boolean;
  persistentStorage: boolean;

  /** Crypto APIs */
  webCrypto: boolean;
  subtleCrypto: boolean;
  sha1Support: boolean;
  sha256Support: boolean;

  /** Compression APIs */
  compressionStream: boolean;
  decompressionStream: boolean;

  /** Worker APIs */
  webWorkers: boolean;
  sharedArrayBuffer: boolean;
  atomics: boolean;

  /** Performance APIs */
  performanceAPI: boolean;
  performanceMemory: boolean;
  performanceObserver: boolean;

  /** Storage Quota */
  storageQuota?: {
    usage: number;
    quota: number;
    available: number;
    percentage: number;
  };
}

export interface CompatibilityResult {
  compatible: boolean;
  capabilities: BrowserCapabilities;
  missingFeatures: string[];
  recommendations: string[];
  warnings: string[];
}

/**
 * Detects browser name and version
 */
export function detectBrowser(): { name: string; version: string; engine: string } {
  const ua = navigator.userAgent;

  // Detect browser
  let name = 'Unknown';
  let version = 'Unknown';
  let engine = 'Unknown';

  if (ua.includes('Chrome') && !ua.includes('Edg')) {
    name = 'Chrome';
    const match = ua.match(/Chrome\/(\d+)/);
    if (match) version = match[1];
    engine = 'Blink';
  } else if (ua.includes('Edg/')) {
    name = 'Edge';
    const match = ua.match(/Edg\/(\d+)/);
    if (match) version = match[1];
    engine = 'Blink';
  } else if (ua.includes('Firefox')) {
    name = 'Firefox';
    const match = ua.match(/Firefox\/(\d+)/);
    if (match) version = match[1];
    engine = 'Gecko';
  } else if (ua.includes('Safari') && !ua.includes('Chrome')) {
    name = 'Safari';
    const match = ua.match(/Version\/(\d+)/);
    if (match) version = match[1];
    engine = 'WebKit';
  }

  return { name, version, engine };
}

/**
 * Checks if WebAssembly is supported
 */
export function hasWebAssembly(): boolean {
  return typeof WebAssembly !== 'undefined';
}

/**
 * Checks if IndexedDB is supported
 */
export async function hasIndexedDB(): Promise<boolean> {
  if (typeof indexedDB === 'undefined') {
    return false;
  }

  try {
    const dbName = '__test_idb__';
    const request = indexedDB.open(dbName, 1);

    return new Promise<boolean>((resolve) => {
      request.onsuccess = () => {
        request.result.close();
        indexedDB.deleteDatabase(dbName);
        resolve(true);
      };

      request.onerror = () => {
        resolve(false);
      };

      setTimeout(() => resolve(false), 1000);
    });
  } catch {
    return false;
  }
}

/**
 * Checks if OPFS (Origin Private File System) is supported
 */
export function hasOPFS(): boolean {
  return (
    typeof navigator !== 'undefined' &&
    'storage' in navigator &&
    typeof navigator.storage === 'object' &&
    'getDirectory' in navigator.storage
  );
}

/**
 * Checks if localStorage is supported
 */
export function hasLocalStorage(): boolean {
  try {
    if (typeof localStorage === 'undefined') {
      return false;
    }

    const test = '__test__';
    localStorage.setItem(test, test);
    localStorage.removeItem(test);
    return true;
  } catch {
    return false;
  }
}

/**
 * Checks if sessionStorage is supported
 */
export function hasSessionStorage(): boolean {
  try {
    if (typeof sessionStorage === 'undefined') {
      return false;
    }

    const test = '__test__';
    sessionStorage.setItem(test, test);
    sessionStorage.removeItem(test);
    return true;
  } catch {
    return false;
  }
}

/**
 * Checks if Web Crypto API is supported
 */
export function hasWebCrypto(): boolean {
  return typeof crypto !== 'undefined' && 'subtle' in crypto;
}

/**
 * Checks if specific hash algorithm is supported
 */
export async function hasHashAlgorithm(algorithm: 'SHA-1' | 'SHA-256'): Promise<boolean> {
  if (!hasWebCrypto()) {
    return false;
  }

  try {
    const data = new Uint8Array([1, 2, 3]);
    await crypto.subtle.digest(algorithm, data);
    return true;
  } catch {
    return false;
  }
}

/**
 * Checks if CompressionStream is supported
 */
export function hasCompressionStream(): boolean {
  return typeof CompressionStream !== 'undefined';
}

/**
 * Checks if Web Workers are supported
 */
export function hasWebWorkers(): boolean {
  return typeof Worker !== 'undefined';
}

/**
 * Checks if SharedArrayBuffer is supported
 */
export function hasSharedArrayBuffer(): boolean {
  return typeof SharedArrayBuffer !== 'undefined';
}

/**
 * Gets storage quota information
 */
export async function getStorageQuota(): Promise<{
  usage: number;
  quota: number;
  available: number;
  percentage: number;
} | null> {
  if (!navigator.storage || !navigator.storage.estimate) {
    return null;
  }

  try {
    const estimate = await navigator.storage.estimate();
    const usage = estimate.usage || 0;
    const quota = estimate.quota || 0;
    const available = quota - usage;
    const percentage = quota > 0 ? (usage / quota) * 100 : 0;

    return { usage, quota, available, percentage };
  } catch {
    return null;
  }
}

/**
 * Checks if persistent storage is available
 */
export async function hasPersistentStorage(): Promise<boolean> {
  if (!navigator.storage || !navigator.storage.persist) {
    return false;
  }

  try {
    return await navigator.storage.persisted();
  } catch {
    return false;
  }
}

/**
 * Requests persistent storage
 */
export async function requestPersistentStorage(): Promise<boolean> {
  if (!navigator.storage || !navigator.storage.persist) {
    return false;
  }

  try {
    return await navigator.storage.persist();
  } catch {
    return false;
  }
}

/**
 * Detects all browser capabilities
 */
export async function detectCapabilities(): Promise<BrowserCapabilities> {
  const browser = detectBrowser();

  const [
    indexedDB,
    sha1,
    sha256,
    storageQuota,
    persistentStorage
  ] = await Promise.all([
    hasIndexedDB(),
    hasHashAlgorithm('SHA-1'),
    hasHashAlgorithm('SHA-256'),
    getStorageQuota(),
    hasPersistentStorage(),
  ]);

  return {
    browser,
    webAssembly: hasWebAssembly(),
    indexedDB,
    localStorage: hasLocalStorage(),
    sessionStorage: hasSessionStorage(),
    opfs: hasOPFS(),
    storageManager: typeof navigator.storage !== 'undefined',
    persistentStorage,
    webCrypto: hasWebCrypto(),
    subtleCrypto: hasWebCrypto(),
    sha1Support: sha1,
    sha256Support: sha256,
    compressionStream: hasCompressionStream(),
    decompressionStream: typeof DecompressionStream !== 'undefined',
    webWorkers: hasWebWorkers(),
    sharedArrayBuffer: hasSharedArrayBuffer(),
    atomics: typeof Atomics !== 'undefined',
    performanceAPI: typeof performance !== 'undefined',
    performanceMemory: typeof performance !== 'undefined' && 'memory' in performance,
    performanceObserver: typeof PerformanceObserver !== 'undefined',
    storageQuota: storageQuota || undefined,
  };
}

/**
 * Checks browser compatibility with BrowserGit
 */
export async function checkCompatibility(): Promise<CompatibilityResult> {
  const capabilities = await detectCapabilities();
  const missingFeatures: string[] = [];
  const recommendations: string[] = [];
  const warnings: string[] = [];

  // Required features
  if (!capabilities.webAssembly) {
    missingFeatures.push('WebAssembly');
  }

  if (!capabilities.webCrypto || !capabilities.subtleCrypto) {
    missingFeatures.push('Web Crypto API');
  }

  if (!capabilities.sha1Support) {
    missingFeatures.push('SHA-1 support');
  }

  // Recommended features
  if (!capabilities.indexedDB) {
    warnings.push('IndexedDB not available - only in-memory storage will work');
    recommendations.push('Use a modern browser with IndexedDB support for persistence');
  }

  if (!capabilities.opfs && capabilities.indexedDB) {
    recommendations.push('Use IndexedDB storage adapter for best performance in your browser');
  }

  if (capabilities.opfs) {
    recommendations.push('Use OPFS storage adapter for best performance');
  }

  if (!capabilities.compressionStream) {
    warnings.push('CompressionStream not available - compression will be slower');
  }

  if (capabilities.storageQuota) {
    const { available, quota, percentage } = capabilities.storageQuota;
    if (percentage > 80) {
      warnings.push(`Storage is ${percentage.toFixed(1)}% full - consider clearing data`);
    }
    if (available < 50 * 1024 * 1024) {
      warnings.push(`Only ${(available / 1024 / 1024).toFixed(2)}MB storage available`);
    }
  }

  // Browser-specific recommendations
  if (capabilities.browser.name === 'Safari') {
    recommendations.push('Safari has more restrictive storage quotas - monitor storage usage carefully');
    if (parseInt(capabilities.browser.version) < 15) {
      warnings.push('Safari version < 15 may have limited support');
    }
  }

  if (capabilities.browser.name === 'Firefox') {
    if (!capabilities.opfs) {
      recommendations.push('OPFS not available in this Firefox version - use IndexedDB');
    }
  }

  const compatible = missingFeatures.length === 0;

  return {
    compatible,
    capabilities,
    missingFeatures,
    recommendations,
    warnings,
  };
}

/**
 * Generates a compatibility report as text
 */
export function generateCompatibilityReport(result: CompatibilityResult): string {
  const lines: string[] = [];

  lines.push('=== Browser Compatibility Report ===');
  lines.push('');

  // Browser info
  lines.push(`Browser: ${result.capabilities.browser.name} ${result.capabilities.browser.version}`);
  lines.push(`Engine: ${result.capabilities.browser.engine}`);
  lines.push('');

  // Compatibility status
  if (result.compatible) {
    lines.push('✅ Compatible - All required features are supported');
  } else {
    lines.push('❌ Not Compatible - Missing required features');
  }
  lines.push('');

  // Missing features
  if (result.missingFeatures.length > 0) {
    lines.push('Missing Required Features:');
    result.missingFeatures.forEach(feature => {
      lines.push(`  ❌ ${feature}`);
    });
    lines.push('');
  }

  // Core features
  lines.push('Core Features:');
  lines.push(`  ${result.capabilities.webAssembly ? '✅' : '❌'} WebAssembly`);
  lines.push(`  ${result.capabilities.webCrypto ? '✅' : '❌'} Web Crypto API`);
  lines.push(`  ${result.capabilities.sha1Support ? '✅' : '❌'} SHA-1`);
  lines.push(`  ${result.capabilities.sha256Support ? '✅' : '❌'} SHA-256`);
  lines.push('');

  // Storage features
  lines.push('Storage Features:');
  lines.push(`  ${result.capabilities.indexedDB ? '✅' : '❌'} IndexedDB`);
  lines.push(`  ${result.capabilities.opfs ? '✅' : '❌'} OPFS`);
  lines.push(`  ${result.capabilities.localStorage ? '✅' : '❌'} localStorage`);
  lines.push(`  ${result.capabilities.sessionStorage ? '✅' : '❌'} sessionStorage`);
  lines.push('');

  // Storage quota
  if (result.capabilities.storageQuota) {
    const { usage, quota, available, percentage } = result.capabilities.storageQuota;
    lines.push('Storage Quota:');
    lines.push(`  Total: ${(quota / 1024 / 1024 / 1024).toFixed(2)} GB`);
    lines.push(`  Used: ${(usage / 1024 / 1024).toFixed(2)} MB (${percentage.toFixed(1)}%)`);
    lines.push(`  Available: ${(available / 1024 / 1024).toFixed(2)} MB`);
    lines.push('');
  }

  // Warnings
  if (result.warnings.length > 0) {
    lines.push('Warnings:');
    result.warnings.forEach(warning => {
      lines.push(`  ⚠️  ${warning}`);
    });
    lines.push('');
  }

  // Recommendations
  if (result.recommendations.length > 0) {
    lines.push('Recommendations:');
    result.recommendations.forEach(rec => {
      lines.push(`  💡 ${rec}`);
    });
    lines.push('');
  }

  return lines.join('\n');
}

/**
 * Logs compatibility report to console
 */
export async function logCompatibilityReport(): Promise<void> {
  const result = await checkCompatibility();
  const report = generateCompatibilityReport(result);
  console.log(report);
}

/**
 * Throws error if browser is not compatible
 */
export async function assertCompatibility(): Promise<void> {
  const result = await checkCompatibility();
  if (!result.compatible) {
    throw new Error(
      `Browser is not compatible with BrowserGit.\nMissing features: ${result.missingFeatures.join(', ')}`
    );
  }
}
