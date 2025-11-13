import type { IDiffEngine, IDiffEngineFactory } from './interface.js';
import { MyersDiffEngine } from './diff-engine.js';

/**
 * Factory for creating diff engine instances
 */
export class DiffEngineFactory implements IDiffEngineFactory {
  private engines: Map<string, IDiffEngine | (() => IDiffEngine)> = new Map();

  constructor() {
    // Register default engines
    this.register('myers', () => new MyersDiffEngine());
    this.register('default', () => new MyersDiffEngine());
  }

  /**
   * Create a new diff engine instance
   */
  create(name = 'default'): IDiffEngine {
    const engine = this.engines.get(name);

    if (!engine) {
      throw new Error(`Unknown diff engine: ${name}`);
    }

    if (typeof engine === 'function') {
      return engine();
    }

    return engine;
  }

  /**
   * Register a custom diff engine implementation
   */
  register(
    name: string,
    engine: IDiffEngine | (() => IDiffEngine)
  ): void {
    this.engines.set(name, engine);
  }

  /**
   * List available diff engine names
   */
  listEngines(): string[] {
    return Array.from(this.engines.keys());
  }
}

// Export singleton factory instance
export const diffEngineFactory = new DiffEngineFactory();
