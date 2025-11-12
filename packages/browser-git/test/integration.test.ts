/**
 * Integration tests combining filesystem with different storage adapters
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { FileSystem } from '../src/filesystem/fs.js';
import {
  MemoryAdapter,
  IndexedDBAdapter,
  LocalStorageAdapter,
} from '@browser-git/storage-adapters';

// Test with multiple storage adapters
const adapters = [
  { name: 'MemoryAdapter', create: () => new MemoryAdapter() },
  { name: 'IndexedDBAdapter', create: () => new IndexedDBAdapter('test-fs-db') },
  {
    name: 'LocalStorageAdapter',
    create: () => new LocalStorageAdapter('test-fs-'),
  },
];

adapters.forEach(({ name, create }) => {
  describe(`FileSystem with ${name}`, () => {
    let fs: FileSystem;

    beforeEach(async () => {
      const adapter = create();
      await adapter.clear();
      fs = new FileSystem(adapter);
    });

    it('should create and read a complete directory structure', async () => {
      // Create a complex directory structure
      await fs.mkdir('project', { recursive: true });
      await fs.mkdir('project/src', { recursive: true });
      await fs.mkdir('project/tests', { recursive: true });
      await fs.mkdir('project/docs', { recursive: true });

      // Write files
      await fs.writeFile('project/README.md', '# Project', { encoding: 'utf8' });
      await fs.writeFile('project/src/index.ts', 'export {}', {
        encoding: 'utf8',
      });
      await fs.writeFile('project/src/utils.ts', 'export const foo = 1;', {
        encoding: 'utf8',
      });
      await fs.writeFile('project/tests/index.test.ts', 'test("works", () => {})', {
        encoding: 'utf8',
      });

      // Verify structure
      const rootEntries = await fs.readdir('project');
      expect(rootEntries).toContain('README.md');
      expect(rootEntries).toContain('src');
      expect(rootEntries).toContain('tests');
      expect(rootEntries).toContain('docs');

      const srcEntries = await fs.readdir('project/src');
      expect(srcEntries).toHaveLength(2);
      expect(srcEntries).toContain('index.ts');
      expect(srcEntries).toContain('utils.ts');

      // Verify file contents
      const readme = await fs.readFile('project/README.md', { encoding: 'utf8' });
      expect(readme).toBe('# Project');

      const utils = await fs.readFile('project/src/utils.ts', { encoding: 'utf8' });
      expect(utils).toBe('export const foo = 1;');
    });

    it('should handle large files', async () => {
      // Create a large file (10KB)
      const largeContent = 'x'.repeat(10 * 1024);
      await fs.writeFile('large.txt', largeContent, { encoding: 'utf8' });

      const content = await fs.readFile('large.txt', { encoding: 'utf8' });
      expect(content).toBe(largeContent);

      const stats = await fs.stat('large.txt');
      expect(stats.size).toBe(10 * 1024);
    });

    it('should handle binary data correctly', async () => {
      const data = new Uint8Array(256);
      for (let i = 0; i < 256; i++) {
        data[i] = i;
      }

      await fs.writeFile('binary.dat', data);
      const read = await fs.readFile('binary.dat');

      expect(read).toEqual(data);
      expect(read.length).toBe(256);
    });

    it('should handle concurrent operations', async () => {
      // Perform multiple operations concurrently
      await Promise.all([
        fs.writeFile('file1.txt', 'content1'),
        fs.writeFile('file2.txt', 'content2'),
        fs.writeFile('file3.txt', 'content3'),
        fs.mkdir('dir1', { recursive: true }),
        fs.mkdir('dir2', { recursive: true }),
      ]);

      const [content1, content2, content3] = await Promise.all([
        fs.readFile('file1.txt', { encoding: 'utf8' }),
        fs.readFile('file2.txt', { encoding: 'utf8' }),
        fs.readFile('file3.txt', { encoding: 'utf8' }),
      ]);

      expect(content1).toBe('content1');
      expect(content2).toBe('content2');
      expect(content3).toBe('content3');
    });

    it('should update file timestamps correctly', async () => {
      await fs.writeFile('test.txt', 'initial');
      const stats1 = await fs.stat('test.txt');

      // Wait a bit
      await new Promise((resolve) => setTimeout(resolve, 10));

      // Modify file
      await fs.writeFile('test.txt', 'modified');
      const stats2 = await fs.stat('test.txt');

      expect(stats2.mtimeMs).toBeGreaterThan(stats1.mtimeMs);
      expect(stats2.ctimeMs).toBe(stats1.ctimeMs); // Creation time unchanged
    });

    it('should clean up properly when deleting directories', async () => {
      // Create a nested structure
      await fs.mkdir('root/sub1/sub2', { recursive: true });
      await fs.writeFile('root/file.txt', 'content');
      await fs.writeFile('root/sub1/file.txt', 'content');
      await fs.writeFile('root/sub1/sub2/file.txt', 'content');

      // Delete recursively
      await fs.rmdir('root', { recursive: true });

      // Verify everything is gone
      expect(await fs.exists('root')).toBe(false);
      expect(await fs.exists('root/file.txt')).toBe(false);
      expect(await fs.exists('root/sub1')).toBe(false);
      expect(await fs.exists('root/sub1/file.txt')).toBe(false);
      expect(await fs.exists('root/sub1/sub2')).toBe(false);
      expect(await fs.exists('root/sub1/sub2/file.txt')).toBe(false);
    });

    it('should handle file overwrites correctly', async () => {
      await fs.writeFile('test.txt', 'short');
      const stats1 = await fs.stat('test.txt');

      await fs.writeFile('test.txt', 'much longer content');
      const stats2 = await fs.stat('test.txt');

      const content = await fs.readFile('test.txt', { encoding: 'utf8' });
      expect(content).toBe('much longer content');
      expect(stats2.size).toBeGreaterThan(stats1.size);
    });

    it('should handle many small files', async () => {
      await fs.mkdir('manyfiles', { recursive: true });

      // Create 50 small files
      const promises = [];
      for (let i = 0; i < 50; i++) {
        promises.push(fs.writeFile(`manyfiles/file${i}.txt`, `content ${i}`));
      }
      await Promise.all(promises);

      const entries = await fs.readdir('manyfiles');
      expect(entries).toHaveLength(50);

      // Read a few random files
      const content10 = await fs.readFile('manyfiles/file10.txt', {
        encoding: 'utf8',
      });
      expect(content10).toBe('content 10');

      const content25 = await fs.readFile('manyfiles/file25.txt', {
        encoding: 'utf8',
      });
      expect(content25).toBe('content 25');
    });

    it('should handle different encodings', async () => {
      const text = 'Hello, World! 🌍';

      await fs.writeFile('utf8.txt', text, { encoding: 'utf8' });
      const utf8Read = await fs.readFile('utf8.txt', { encoding: 'utf8' });
      expect(utf8Read).toBe(text);

      const hexRead = await fs.readFile('utf8.txt', { encoding: 'hex' });
      expect(hexRead).toMatch(/^[0-9a-f]+$/);

      const base64Read = await fs.readFile('utf8.txt', { encoding: 'base64' });
      expect(base64Read).toMatch(/^[A-Za-z0-9+/]+=*$/);
    });

    it('should handle complex path normalization', async () => {
      // Create files with various path formats
      await fs.writeFile('./dir/../file.txt', 'content', { recursive: true });
      expect(await fs.exists('file.txt')).toBe(true);

      await fs.writeFile('dir1//dir2///file.txt', 'content2', { recursive: true });
      const content = await fs.readFile('dir1/dir2/file.txt', { encoding: 'utf8' });
      expect(content).toBe('content2');
    });

    it('should handle empty directories', async () => {
      await fs.mkdir('empty1');
      await fs.mkdir('empty2');
      await fs.mkdir('empty3');

      const entries1 = await fs.readdir('empty1');
      expect(entries1).toEqual([]);

      const stats = await fs.stat('empty2');
      expect(stats.isDirectory).toBe(true);
      expect(stats.size).toBe(0);
    });

    it('should track file sizes correctly', async () => {
      await fs.writeFile('small.txt', 'Hi');
      await fs.writeFile('medium.txt', 'x'.repeat(1000));
      await fs.writeFile('large.txt', 'y'.repeat(10000));

      const smallStats = await fs.stat('small.txt');
      const mediumStats = await fs.stat('medium.txt');
      const largeStats = await fs.stat('large.txt');

      expect(smallStats.size).toBe(2);
      expect(mediumStats.size).toBe(1000);
      expect(largeStats.size).toBe(10000);
    });
  });
});

// Integration test for file watchers with storage
describe('FileSystem watchers with storage', () => {
  let fs: FileSystem;

  beforeEach(async () => {
    const adapter = new MemoryAdapter();
    await adapter.clear();
    fs = new FileSystem(adapter);
  });

  it('should emit events for file operations across storage', async () => {
    const events: any[] = [];
    const watcher = fs.watch('test.txt', (event) => {
      events.push(event);
    });

    // Create file
    await fs.writeFile('test.txt', 'initial');
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('create');

    // Modify file
    await fs.writeFile('test.txt', 'modified');
    expect(events).toHaveLength(2);
    expect(events[1].type).toBe('modify');

    // Delete file
    await fs.unlink('test.txt');
    expect(events).toHaveLength(3);
    expect(events[2].type).toBe('delete');

    watcher.close();
  });

  it('should track directory events', async () => {
    const events: any[] = [];
    const watcher = fs.watch('testdir', (event) => {
      events.push(event);
    });

    // Create directory
    await fs.mkdir('testdir');
    expect(events).toHaveLength(1);
    expect(events[0].type).toBe('create');

    // Delete directory
    await fs.rmdir('testdir');
    expect(events).toHaveLength(2);
    expect(events[1].type).toBe('delete');

    watcher.close();
  });
});
