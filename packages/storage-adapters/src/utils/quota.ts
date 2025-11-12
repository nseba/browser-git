import type { StorageAdapter, StorageQuota } from '../interface.js';

/**
 * Utility functions for storage quota management.
 */

/**
 * Format bytes into a human-readable string.
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

/**
 * Check if storage quota is approaching limit (>80%).
 */
export function isQuotaNearLimit(quota: StorageQuota): boolean {
  return quota.percentage > 80;
}

/**
 * Check if storage quota is critical (>95%).
 */
export function isQuotaCritical(quota: StorageQuota): boolean {
  return quota.percentage > 95;
}

/**
 * Get available space in bytes.
 */
export function getAvailableSpace(quota: StorageQuota): number {
  return Math.max(0, quota.quota - quota.usage);
}

/**
 * Get quota status with formatted strings.
 */
export interface QuotaStatus {
  used: string;
  total: string;
  available: string;
  percentage: number;
  isNearLimit: boolean;
  isCritical: boolean;
}

export async function getQuotaStatus(
  adapter: StorageAdapter
): Promise<QuotaStatus | null> {
  if (!adapter.getQuota) {
    return null;
  }
  const quota = await adapter.getQuota();
  if (!quota) {
    return null;
  }

  return {
    used: formatBytes(quota.usage),
    total: formatBytes(quota.quota),
    available: formatBytes(getAvailableSpace(quota)),
    percentage: Math.round(quota.percentage * 100) / 100,
    isNearLimit: isQuotaNearLimit(quota),
    isCritical: isQuotaCritical(quota),
  };
}
