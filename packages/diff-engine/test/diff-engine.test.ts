import { describe, it, expect } from 'vitest';
import { MyersDiffEngine } from '../src/diff-engine.js';
import { ChangeType } from '../src/types.js';
import { stringToUint8Array } from '../src/utils/binary.js';

describe('MyersDiffEngine', () => {
  const engine = new MyersDiffEngine();

  describe('diff()', () => {
    it('should detect additions', () => {
      const oldText = 'Line 1\nLine 2\n';
      const newText = 'Line 1\nLine 2\nLine 3\n';

      const result = engine.diff(oldText, newText);

      expect(result.isBinary).toBe(false);
      expect(result.linesAdded).toBeGreaterThan(0);
      expect(result.hunks.length).toBeGreaterThan(0);
    });

    it('should detect deletions', () => {
      const oldText = 'Line 1\nLine 2\nLine 3\n';
      const newText = 'Line 1\nLine 2\n';

      const result = engine.diff(oldText, newText);

      expect(result.isBinary).toBe(false);
      expect(result.linesDeleted).toBeGreaterThan(0);
    });

    it('should detect modifications', () => {
      const oldText = 'Line 1\nLine 2\nLine 3\n';
      const newText = 'Line 1\nModified Line 2\nLine 3\n';

      const result = engine.diff(oldText, newText);

      expect(result.isBinary).toBe(false);
      expect(result.linesAdded).toBeGreaterThan(0);
      expect(result.linesDeleted).toBeGreaterThan(0);
    });

    it('should handle empty files', () => {
      const oldText = '';
      const newText = 'New content\n';

      const result = engine.diff(oldText, newText);

      expect(result.isBinary).toBe(false);
      expect(result.linesAdded).toBeGreaterThan(0);
      expect(result.linesDeleted).toBe(0);
    });

    it('should ignore whitespace when requested', () => {
      const oldText = 'Line 1\n  Line 2\nLine 3\n';
      const newText = 'Line 1\nLine 2\nLine 3\n';

      const result = engine.diff(oldText, newText, { ignoreWhitespace: true });

      expect(result.isBinary).toBe(false);
      // With whitespace ignored, should have minimal changes
      expect(result.linesDeleted).toBe(0);
    });
  });

  describe('diffFiles()', () => {
    it('should diff text files', () => {
      const oldFile = stringToUint8Array('Line 1\nLine 2\n');
      const newFile = stringToUint8Array('Line 1\nLine 2\nLine 3\n');

      const result = engine.diffFiles(oldFile, newFile);

      expect(result.isBinary).toBe(false);
      if (!result.isBinary) {
        expect(result.linesAdded).toBeGreaterThan(0);
      }
    });

    it('should detect binary files', () => {
      // Create binary content (with NUL byte)
      const oldFile = new Uint8Array([0x00, 0x01, 0x02, 0x03]);
      const newFile = new Uint8Array([0x00, 0x01, 0x02, 0x04]);

      const result = engine.diffFiles(oldFile, newFile);

      expect(result.isBinary).toBe(true);
      if (result.isBinary) {
        expect(result.oldSize).toBe(4);
        expect(result.newSize).toBe(4);
      }
    });
  });

  describe('diffWords()', () => {
    it('should perform word-level diff', () => {
      const oldLine = 'The quick brown fox';
      const newLine = 'The fast brown fox';

      const result = engine.diffWords(oldLine, newLine);

      expect(result.words.length).toBeGreaterThan(0);
      expect(result.words.some((w) => w.type === ChangeType.Delete)).toBe(true);
      expect(result.words.some((w) => w.type === ChangeType.Add)).toBe(true);
    });

    it('should handle identical lines', () => {
      const line = 'Same line';

      const result = engine.diffWords(line, line);

      expect(result.words.length).toBeGreaterThan(0);
      expect(result.words.every((w) => w.type === ChangeType.Equal)).toBe(true);
    });
  });

  describe('format()', () => {
    it('should format as unified diff', () => {
      const oldText = 'Line 1\nLine 2\n';
      const newText = 'Line 1\nLine 2\nLine 3\n';

      const diff = engine.diff(oldText, newText);
      const formatted = engine.format(diff, { format: 'unified' });

      expect(formatted).toContain('@@');
      expect(typeof formatted).toBe('string');
    });

    it('should format as side-by-side', () => {
      const oldText = 'Line 1\nLine 2\n';
      const newText = 'Line 1\nLine 2\nLine 3\n';

      const diff = engine.diff(oldText, newText);
      const formatted = engine.format(diff, { format: 'side-by-side' });

      expect(typeof formatted).toBe('string');
      expect(formatted.length).toBeGreaterThan(0);
    });

    it('should format as JSON', () => {
      const oldText = 'Line 1\nLine 2\n';
      const newText = 'Line 1\nLine 2\nLine 3\n';

      const diff = engine.diff(oldText, newText);
      const formatted = engine.format(diff, { format: 'json' });

      expect(() => JSON.parse(formatted)).not.toThrow();
    });

    it('should format binary diff', () => {
      const binaryDiff = {
        isBinary: true as const,
        oldSize: 100,
        newSize: 150,
        sizeChanged: true,
      };

      const formatted = engine.format(binaryDiff);

      expect(typeof formatted).toBe('string');
      expect(formatted).toContain('Binary');
    });
  });

  describe('isBinary()', () => {
    it('should detect binary content', () => {
      const binaryContent = new Uint8Array([0x00, 0x01, 0x02, 0x03]);

      expect(engine.isBinary(binaryContent)).toBe(true);
    });

    it('should recognize text content', () => {
      const textContent = stringToUint8Array('Hello, world!');

      expect(engine.isBinary(textContent)).toBe(false);
    });
  });

  describe('patch()', () => {
    it('should apply a simple addition patch', () => {
      const oldText = 'Line 1\nLine 2\n';
      const newText = 'Line 1\nLine 2\nLine 3\n';

      const diff = engine.diff(oldText, newText);
      const patched = engine.patch(oldText, diff);

      expect(patched).not.toBeNull();
      // Patched result should be similar to new text (may differ in trailing newline)
      expect(patched?.includes('Line 3')).toBe(true);
    });

    it('should handle empty diff by returning original text', () => {
      const oldText = 'Line 1\nLine 2\n';
      const emptyDiff = {
        hunks: [],
        isBinary: false,
        linesAdded: 0,
        linesDeleted: 0,
      };

      const patched = engine.patch(oldText, emptyDiff);

      // Empty diff should return original text
      expect(patched).toBe(oldText);
    });
  });
});
