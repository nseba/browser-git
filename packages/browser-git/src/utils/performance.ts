/**
 * Performance monitoring and optimization utilities
 * Helps track and optimize Git operations to meet performance targets:
 * - Commit: < 50ms
 * - Checkout: < 200ms
 * - Clone (100 commits): < 5s
 */

export interface PerformanceMetrics {
  operation: string;
  duration: number;
  timestamp: number;
  metadata?: Record<string, any>;
}

export interface PerformanceStats {
  count: number;
  totalDuration: number;
  avgDuration: number;
  minDuration: number;
  maxDuration: number;
  p50: number;
  p95: number;
  p99: number;
}

export interface MemorySnapshot {
  timestamp: number;
  heapUsed?: number;
  heapTotal?: number;
  external?: number;
}

/**
 * Performance monitor for tracking Git operation performance
 */
export class PerformanceMonitor {
  private metrics: Map<string, PerformanceMetrics[]> = new Map();
  private memorySnapshots: MemorySnapshot[] = [];
  private enabled: boolean = true;

  constructor(enabled: boolean = true) {
    this.enabled = enabled;
  }

  /**
   * Enable performance monitoring
   */
  enable(): void {
    this.enabled = true;
  }

  /**
   * Disable performance monitoring
   */
  disable(): void {
    this.enabled = false;
  }

  /**
   * Measure operation performance
   */
  async measure<T>(
    operation: string,
    fn: () => Promise<T>,
    metadata?: Record<string, any>
  ): Promise<T> {
    if (!this.enabled) {
      return fn();
    }

    const start = this.now();
    try {
      const result = await fn();
      const duration = this.now() - start;
      this.recordMetric(operation, duration, metadata);
      return result;
    } catch (error) {
      const duration = this.now() - start;
      this.recordMetric(operation, duration, {
        ...metadata,
        error: true,
      });
      throw error;
    }
  }

  /**
   * Measure synchronous operation performance
   */
  measureSync<T>(
    operation: string,
    fn: () => T,
    metadata?: Record<string, any>
  ): T {
    if (!this.enabled) {
      return fn();
    }

    const start = this.now();
    try {
      const result = fn();
      const duration = this.now() - start;
      this.recordMetric(operation, duration, metadata);
      return result;
    } catch (error) {
      const duration = this.now() - start;
      this.recordMetric(operation, duration, {
        ...metadata,
        error: true,
      });
      throw error;
    }
  }

  /**
   * Start timing an operation
   */
  startTimer(operation: string): () => void {
    if (!this.enabled) {
      return () => {};
    }

    const start = this.now();
    return (metadata?: Record<string, any>) => {
      const duration = this.now() - start;
      this.recordMetric(operation, duration, metadata);
    };
  }

  /**
   * Record a metric
   */
  recordMetric(
    operation: string,
    duration: number,
    metadata?: Record<string, any>
  ): void {
    if (!this.enabled) {
      return;
    }

    const metric: PerformanceMetrics = {
      operation,
      duration,
      timestamp: Date.now(),
    };

    if (metadata) {
      metric.metadata = metadata;
    }

    if (!this.metrics.has(operation)) {
      this.metrics.set(operation, []);
    }

    this.metrics.get(operation)!.push(metric);

    // Warn if operation exceeds performance targets
    this.checkPerformanceTarget(operation, duration);
  }

  /**
   * Get statistics for an operation
   */
  getStats(operation: string): PerformanceStats | null {
    const metrics = this.metrics.get(operation);
    if (!metrics || metrics.length === 0) {
      return null;
    }

    const durations = metrics.map((m) => m.duration).sort((a, b) => a - b);
    const count = durations.length;
    const totalDuration = durations.reduce((sum, d) => sum + d, 0);
    const avgDuration = totalDuration / count;
    const minDuration = durations[0]!;
    const maxDuration = durations[count - 1]!;

    const p50 = this.percentile(durations, 50);
    const p95 = this.percentile(durations, 95);
    const p99 = this.percentile(durations, 99);

    return {
      count,
      totalDuration,
      avgDuration,
      minDuration,
      maxDuration,
      p50,
      p95,
      p99,
    };
  }

  /**
   * Get all metrics for an operation
   */
  getMetrics(operation: string): PerformanceMetrics[] {
    return this.metrics.get(operation) || [];
  }

  /**
   * Get all operations being tracked
   */
  getOperations(): string[] {
    return Array.from(this.metrics.keys());
  }

  /**
   * Clear metrics for an operation or all operations
   */
  clear(operation?: string): void {
    if (operation) {
      this.metrics.delete(operation);
    } else {
      this.metrics.clear();
    }
  }

  /**
   * Take a memory snapshot
   */
  takeMemorySnapshot(): MemorySnapshot {
    const snapshot: MemorySnapshot = {
      timestamp: Date.now(),
    };

    // Chrome/Edge specific
    if (typeof performance !== 'undefined' && 'memory' in performance) {
      const memory = (performance as any).memory;
      snapshot.heapUsed = memory.usedJSHeapSize;
      snapshot.heapTotal = memory.totalJSHeapSize;
      snapshot.external = memory.jsHeapSizeLimit;
    }

    this.memorySnapshots.push(snapshot);
    return snapshot;
  }

  /**
   * Get memory snapshots
   */
  getMemorySnapshots(): MemorySnapshot[] {
    return this.memorySnapshots;
  }

  /**
   * Clear memory snapshots
   */
  clearMemorySnapshots(): void {
    this.memorySnapshots = [];
  }

  /**
   * Generate performance report
   */
  generateReport(): string {
    const operations = this.getOperations();
    const lines: string[] = [];

    lines.push('Performance Report');
    lines.push('=================');
    lines.push('');

    for (const operation of operations) {
      const stats = this.getStats(operation);
      if (!stats) continue;

      lines.push(`Operation: ${operation}`);
      lines.push(`  Count: ${stats.count}`);
      lines.push(`  Total: ${stats.totalDuration.toFixed(2)}ms`);
      lines.push(`  Average: ${stats.avgDuration.toFixed(2)}ms`);
      lines.push(`  Min: ${stats.minDuration.toFixed(2)}ms`);
      lines.push(`  Max: ${stats.maxDuration.toFixed(2)}ms`);
      lines.push(`  P50: ${stats.p50.toFixed(2)}ms`);
      lines.push(`  P95: ${stats.p95.toFixed(2)}ms`);
      lines.push(`  P99: ${stats.p99.toFixed(2)}ms`);
      lines.push('');
    }

    return lines.join('\n');
  }

  /**
   * Export metrics as JSON
   */
  exportMetrics(): Record<string, PerformanceMetrics[]> {
    const result: Record<string, PerformanceMetrics[]> = {};
    for (const [operation, metrics] of this.metrics) {
      result[operation] = metrics;
    }
    return result;
  }

  /**
   * Get current timestamp
   */
  private now(): number {
    return typeof performance !== 'undefined' && performance.now
      ? performance.now()
      : Date.now();
  }

  /**
   * Calculate percentile
   */
  private percentile(sortedValues: number[], percentile: number): number {
    const index = (percentile / 100) * (sortedValues.length - 1);
    const lower = Math.floor(index);
    const upper = Math.ceil(index);
    const weight = index - lower;

    if (lower === upper) {
      return sortedValues[lower]!;
    }

    return sortedValues[lower]! * (1 - weight) + sortedValues[upper]! * weight;
  }

  /**
   * Check if operation meets performance target
   */
  private checkPerformanceTarget(operation: string, duration: number): void {
    const targets: Record<string, number> = {
      commit: 50, // 50ms
      checkout: 200, // 200ms
      clone: 5000, // 5s
      add: 20, // 20ms
      status: 100, // 100ms
      diff: 50, // 50ms
      merge: 300, // 300ms
    };

    const normalizedOp = operation.toLowerCase();
    const target = targets[normalizedOp];

    if (target && duration > target) {
      console.warn(
        `Performance warning: ${operation} took ${duration.toFixed(2)}ms (target: ${target}ms)`
      );
    }
  }
}

/**
 * Global performance monitor instance
 */
export const performanceMonitor = new PerformanceMonitor();

/**
 * Decorator for measuring method performance
 */
export function measured(operation?: string) {
  return function (
    target: any,
    propertyKey: string,
    descriptor: PropertyDescriptor
  ) {
    const originalMethod = descriptor.value;
    const opName = operation || `${target.constructor.name}.${propertyKey}`;

    descriptor.value = async function (...args: any[]) {
      return performanceMonitor.measure(
        opName,
        () => originalMethod.apply(this, args)
      );
    };

    return descriptor;
  };
}

/**
 * Batch operations for better performance
 */
export class BatchProcessor<T, R> {
  private batch: T[] = [];
  private timeout: number | null = null;
  private readonly batchSize: number;
  private readonly batchDelay: number;
  private readonly processor: (items: T[]) => Promise<R[]>;

  constructor(
    processor: (items: T[]) => Promise<R[]>,
    batchSize: number = 100,
    batchDelay: number = 10
  ) {
    this.processor = processor;
    this.batchSize = batchSize;
    this.batchDelay = batchDelay;
  }

  /**
   * Add item to batch
   */
  add(item: T): Promise<R> {
    return new Promise((resolve, reject) => {
      this.batch.push(item);

      // Store resolve/reject for this item
      (item as any).__resolve = resolve;
      (item as any).__reject = reject;

      // Process batch if it reaches size limit
      if (this.batch.length >= this.batchSize) {
        this.processBatch();
      } else {
        // Schedule batch processing
        if (this.timeout !== null) {
          clearTimeout(this.timeout);
        }
        this.timeout = setTimeout(() => this.processBatch(), this.batchDelay) as any;
      }
    });
  }

  /**
   * Process current batch
   */
  private async processBatch(): Promise<void> {
    if (this.timeout !== null) {
      clearTimeout(this.timeout);
      this.timeout = null;
    }

    if (this.batch.length === 0) {
      return;
    }

    const currentBatch = this.batch;
    this.batch = [];

    try {
      const results = await this.processor(currentBatch);

      // Resolve promises
      for (let i = 0; i < currentBatch.length; i++) {
        const item = currentBatch[i];
        const resolve = (item as any).__resolve;
        if (resolve) {
          resolve(results[i]);
        }
      }
    } catch (error) {
      // Reject all promises
      for (const item of currentBatch) {
        const reject = (item as any).__reject;
        if (reject) {
          reject(error);
        }
      }
    }
  }

  /**
   * Flush remaining items
   */
  async flush(): Promise<void> {
    await this.processBatch();
  }
}

/**
 * Debounce function execution
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: number | null = null;

  return function (this: any, ...args: Parameters<T>) {
    if (timeout !== null) {
      clearTimeout(timeout);
    }

    timeout = setTimeout(() => {
      func.apply(this, args);
    }, wait) as any;
  };
}

/**
 * Throttle function execution
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  limit: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean = false;

  return function (this: any, ...args: Parameters<T>) {
    if (!inThrottle) {
      func.apply(this, args);
      inThrottle = true;
      setTimeout(() => {
        inThrottle = false;
      }, limit);
    }
  };
}

/**
 * Memoize function results
 */
export function memoize<T extends (...args: any[]) => any>(
  func: T
): T {
  const cache = new Map<string, any>();

  return function (this: any, ...args: Parameters<T>): ReturnType<T> {
    const key = JSON.stringify(args);

    if (cache.has(key)) {
      return cache.get(key);
    }

    const result = func.apply(this, args);
    cache.set(key, result);
    return result;
  } as T;
}

/**
 * Format duration for display
 */
export function formatDuration(ms: number): string {
  if (ms < 1) {
    return `${(ms * 1000).toFixed(0)}μs`;
  } else if (ms < 1000) {
    return `${ms.toFixed(2)}ms`;
  } else {
    return `${(ms / 1000).toFixed(2)}s`;
  }
}
