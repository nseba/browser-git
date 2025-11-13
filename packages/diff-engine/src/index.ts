/**
 * @browser-git/diff-engine
 *
 * Pluggable diff engine for BrowserGit with Myers algorithm implementation
 */

// Export types
export * from './types.js';
export * from './interface.js';

// Export implementations
export { MyersDiffEngine } from './diff-engine.js';
export { DiffEngineFactory, diffEngineFactory } from './factory.js';

// Export utilities
export * from './utils/index.js';

// Convenience export: default diff engine
import { MyersDiffEngine } from './diff-engine.js';
export default MyersDiffEngine;
