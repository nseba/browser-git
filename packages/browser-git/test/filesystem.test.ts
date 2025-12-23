/**
 * FileSystem unit tests
 */

import { describe, it, expect, beforeEach } from "vitest";
import { FileSystem, FSError } from "../src/filesystem/fs.js";
import { MemoryAdapter } from "@browser-git/storage-adapters";

describe("FileSystem", () => {
  let fs: FileSystem;
  let storage: MemoryAdapter;

  beforeEach(() => {
    storage = new MemoryAdapter();
    fs = new FileSystem(storage);
  });

  describe("writeFile and readFile", () => {
    it("should write and read a text file", async () => {
      await fs.writeFile("test.txt", "Hello, World!");
      const content = await fs.readFile("test.txt", { encoding: "utf8" });
      expect(content).toBe("Hello, World!");
    });

    it("should write and read a binary file", async () => {
      const data = new Uint8Array([1, 2, 3, 4, 5]);
      await fs.writeFile("test.bin", data);
      const content = await fs.readFile("test.bin");
      expect(content).toEqual(data);
    });

    it("should support different encodings", async () => {
      await fs.writeFile("test.txt", "Hello", { encoding: "utf8" });

      const utf8 = await fs.readFile("test.txt", { encoding: "utf8" });
      expect(utf8).toBe("Hello");

      const hex = await fs.readFile("test.txt", { encoding: "hex" });
      expect(hex).toBe("48656c6c6f");

      const base64 = await fs.readFile("test.txt", { encoding: "base64" });
      expect(base64).toBe("SGVsbG8=");
    });

    it("should create parent directories with recursive option", async () => {
      await fs.writeFile("dir1/dir2/file.txt", "content", { recursive: true });
      const content = await fs.readFile("dir1/dir2/file.txt", {
        encoding: "utf8",
      });
      expect(content).toBe("content");
    });

    it("should throw ENOENT if file does not exist", async () => {
      await expect(fs.readFile("nonexistent.txt")).rejects.toThrow(FSError);
      await expect(fs.readFile("nonexistent.txt")).rejects.toThrow(/ENOENT/);
    });

    it("should throw EISDIR if trying to read a directory", async () => {
      await fs.mkdir("testdir");
      await expect(fs.readFile("testdir")).rejects.toThrow(FSError);
      await expect(fs.readFile("testdir")).rejects.toThrow(/EISDIR/);
    });

    it("should throw EISDIR if trying to write to a directory", async () => {
      await fs.mkdir("testdir");
      await expect(fs.writeFile("testdir", "data")).rejects.toThrow(FSError);
      await expect(fs.writeFile("testdir", "data")).rejects.toThrow(/EISDIR/);
    });

    it("should overwrite existing file", async () => {
      await fs.writeFile("test.txt", "first");
      await fs.writeFile("test.txt", "second");
      const content = await fs.readFile("test.txt", { encoding: "utf8" });
      expect(content).toBe("second");
    });
  });

  describe("mkdir", () => {
    it("should create a directory", async () => {
      await fs.mkdir("testdir");
      const exists = await fs.exists("testdir");
      expect(exists).toBe(true);

      const stat = await fs.stat("testdir");
      expect(stat.isDirectory).toBe(true);
    });

    it("should create nested directories with recursive option", async () => {
      await fs.mkdir("dir1/dir2/dir3", { recursive: true });

      expect(await fs.exists("dir1")).toBe(true);
      expect(await fs.exists("dir1/dir2")).toBe(true);
      expect(await fs.exists("dir1/dir2/dir3")).toBe(true);
    });

    it("should throw EEXIST if directory already exists without recursive", async () => {
      await fs.mkdir("testdir");
      await expect(fs.mkdir("testdir")).rejects.toThrow(FSError);
      await expect(fs.mkdir("testdir")).rejects.toThrow(/EEXIST/);
    });

    it("should not throw if directory exists with recursive option", async () => {
      await fs.mkdir("testdir", { recursive: true });
      await expect(
        fs.mkdir("testdir", { recursive: true }),
      ).resolves.not.toThrow();
    });

    it("should throw ENOENT if parent directory does not exist", async () => {
      await expect(fs.mkdir("nonexistent/child")).rejects.toThrow(FSError);
      await expect(fs.mkdir("nonexistent/child")).rejects.toThrow(/ENOENT/);
    });
  });

  describe("readdir", () => {
    it("should list files and directories", async () => {
      await fs.mkdir("testroot", { recursive: true });
      await fs.writeFile("testroot/file1.txt", "content1");
      await fs.writeFile("testroot/file2.txt", "content2");
      await fs.mkdir("testroot/subdir");
      await fs.writeFile("testroot/subdir/file3.txt", "content3");

      const entries = await fs.readdir("testroot");
      expect(entries).toContain("file1.txt");
      expect(entries).toContain("file2.txt");
      expect(entries).toContain("subdir");
      expect(entries).toHaveLength(3);
    });

    it("should list contents of subdirectory", async () => {
      await fs.mkdir("dir", { recursive: true });
      await fs.writeFile("dir/file1.txt", "content1");
      await fs.writeFile("dir/file2.txt", "content2");

      const entries = await fs.readdir("dir");
      expect(entries).toContain("file1.txt");
      expect(entries).toContain("file2.txt");
      expect(entries).toHaveLength(2);
    });

    it("should return empty array for empty directory", async () => {
      await fs.mkdir("emptydir");
      const entries = await fs.readdir("emptydir");
      expect(entries).toEqual([]);
    });

    it("should throw ENOENT if directory does not exist", async () => {
      await expect(fs.readdir("nonexistent")).rejects.toThrow(FSError);
      await expect(fs.readdir("nonexistent")).rejects.toThrow(/ENOENT/);
    });

    it("should throw ENOTDIR if path is not a directory", async () => {
      await fs.writeFile("file.txt", "content");
      await expect(fs.readdir("file.txt")).rejects.toThrow(FSError);
      await expect(fs.readdir("file.txt")).rejects.toThrow(/ENOTDIR/);
    });
  });

  describe("unlink", () => {
    it("should delete a file", async () => {
      await fs.writeFile("test.txt", "content");
      expect(await fs.exists("test.txt")).toBe(true);

      await fs.unlink("test.txt");
      expect(await fs.exists("test.txt")).toBe(false);
    });

    it("should throw ENOENT if file does not exist", async () => {
      await expect(fs.unlink("nonexistent.txt")).rejects.toThrow(FSError);
      await expect(fs.unlink("nonexistent.txt")).rejects.toThrow(/ENOENT/);
    });

    it("should throw EISDIR if trying to unlink a directory", async () => {
      await fs.mkdir("testdir");
      await expect(fs.unlink("testdir")).rejects.toThrow(FSError);
      await expect(fs.unlink("testdir")).rejects.toThrow(/EISDIR/);
    });
  });

  describe("rmdir", () => {
    it("should remove an empty directory", async () => {
      await fs.mkdir("testdir");
      expect(await fs.exists("testdir")).toBe(true);

      await fs.rmdir("testdir");
      expect(await fs.exists("testdir")).toBe(false);
    });

    it("should throw ENOTEMPTY if directory is not empty", async () => {
      await fs.mkdir("testdir");
      await fs.writeFile("testdir/file.txt", "content");

      await expect(fs.rmdir("testdir")).rejects.toThrow(FSError);
      await expect(fs.rmdir("testdir")).rejects.toThrow(/ENOTEMPTY/);
    });

    it("should remove directory and contents with recursive option", async () => {
      await fs.mkdir("testdir");
      await fs.writeFile("testdir/file1.txt", "content1");
      await fs.mkdir("testdir/subdir", { recursive: true });
      await fs.writeFile("testdir/subdir/file2.txt", "content2");

      await fs.rmdir("testdir", { recursive: true });
      expect(await fs.exists("testdir")).toBe(false);
    });

    it("should throw ENOENT if directory does not exist", async () => {
      await expect(fs.rmdir("nonexistent")).rejects.toThrow(FSError);
      await expect(fs.rmdir("nonexistent")).rejects.toThrow(/ENOENT/);
    });

    it("should throw ENOTDIR if path is not a directory", async () => {
      await fs.writeFile("file.txt", "content");
      await expect(fs.rmdir("file.txt")).rejects.toThrow(FSError);
      await expect(fs.rmdir("file.txt")).rejects.toThrow(/ENOTDIR/);
    });
  });

  describe("stat", () => {
    it("should return stats for a file", async () => {
      await fs.writeFile("test.txt", "Hello");
      const stats = await fs.stat("test.txt");

      expect(stats.path).toBe("test.txt");
      expect(stats.isFile).toBe(true);
      expect(stats.isDirectory).toBe(false);
      expect(stats.size).toBe(5);
      expect(stats.ctimeMs).toBeGreaterThan(0);
      expect(stats.mtimeMs).toBeGreaterThan(0);
      expect(stats.atimeMs).toBeGreaterThan(0);
    });

    it("should return stats for a directory", async () => {
      await fs.mkdir("testdir");
      const stats = await fs.stat("testdir");

      expect(stats.path).toBe("testdir");
      expect(stats.isFile).toBe(false);
      expect(stats.isDirectory).toBe(true);
      expect(stats.size).toBe(0);
    });

    it("should throw ENOENT if path does not exist", async () => {
      await expect(fs.stat("nonexistent")).rejects.toThrow(FSError);
      await expect(fs.stat("nonexistent")).rejects.toThrow(/ENOENT/);
    });

    it("should update access time when calling stat", async () => {
      await fs.writeFile("test.txt", "content");
      const stats1 = await fs.stat("test.txt");

      // Wait a bit
      await new Promise((resolve) => setTimeout(resolve, 10));

      const stats2 = await fs.stat("test.txt");
      expect(stats2.atimeMs).toBeGreaterThanOrEqual(stats1.atimeMs);
    });
  });

  describe("exists", () => {
    it("should return true for existing file", async () => {
      await fs.writeFile("test.txt", "content");
      expect(await fs.exists("test.txt")).toBe(true);
    });

    it("should return true for existing directory", async () => {
      await fs.mkdir("testdir");
      expect(await fs.exists("testdir")).toBe(true);
    });

    it("should return false for non-existing path", async () => {
      expect(await fs.exists("nonexistent")).toBe(false);
    });
  });

  describe("watch", () => {
    it("should emit events when file is created", async () => {
      const events: any[] = [];
      const watcher = fs.watch("test.txt", (event) => {
        events.push(event);
      });

      await fs.writeFile("test.txt", "content");

      expect(events).toHaveLength(1);
      expect(events[0].type).toBe("create");
      expect(events[0].path).toBe("test.txt");
      expect(events[0].timestamp).toBeGreaterThan(0);

      watcher.close();
    });

    it("should emit events when file is modified", async () => {
      await fs.writeFile("test.txt", "content1");

      const events: any[] = [];
      const watcher = fs.watch("test.txt", (event) => {
        events.push(event);
      });

      await fs.writeFile("test.txt", "content2");

      expect(events).toHaveLength(1);
      expect(events[0].type).toBe("modify");

      watcher.close();
    });

    it("should emit events when file is deleted", async () => {
      await fs.writeFile("test.txt", "content");

      const events: any[] = [];
      const watcher = fs.watch("test.txt", (event) => {
        events.push(event);
      });

      await fs.unlink("test.txt");

      expect(events).toHaveLength(1);
      expect(events[0].type).toBe("delete");

      watcher.close();
    });

    it("should stop emitting events after watcher is closed", async () => {
      const events: any[] = [];
      const watcher = fs.watch("test.txt", (event) => {
        events.push(event);
      });

      await fs.writeFile("test.txt", "content1");
      watcher.close();
      await fs.writeFile("test.txt", "content2");

      expect(events).toHaveLength(1);
    });

    it("should support multiple watchers on the same path", async () => {
      const events1: any[] = [];
      const events2: any[] = [];

      const watcher1 = fs.watch("test.txt", (event) => events1.push(event));
      const watcher2 = fs.watch("test.txt", (event) => events2.push(event));

      await fs.writeFile("test.txt", "content");

      expect(events1).toHaveLength(1);
      expect(events2).toHaveLength(1);

      watcher1.close();
      watcher2.close();
    });
  });

  describe("path normalization", () => {
    it("should normalize paths with multiple slashes", async () => {
      await fs.writeFile("dir//file.txt", "content", { recursive: true });
      const content = await fs.readFile("dir/file.txt", { encoding: "utf8" });
      expect(content).toBe("content");
    });

    it("should handle paths with . segments", async () => {
      await fs.writeFile("./dir/./file.txt", "content", { recursive: true });
      const content = await fs.readFile("dir/file.txt", { encoding: "utf8" });
      expect(content).toBe("content");
    });

    it("should handle paths with .. segments", async () => {
      await fs.mkdir("dir1/dir2", { recursive: true });
      await fs.writeFile("dir1/dir2/../file.txt", "content");
      const content = await fs.readFile("dir1/file.txt", { encoding: "utf8" });
      expect(content).toBe("content");
    });
  });

  describe("timestamps", () => {
    it("should update mtime when file is modified", async () => {
      await fs.writeFile("test.txt", "content1");
      const stats1 = await fs.stat("test.txt");

      await new Promise((resolve) => setTimeout(resolve, 10));

      await fs.writeFile("test.txt", "content2");
      const stats2 = await fs.stat("test.txt");

      expect(stats2.mtimeMs).toBeGreaterThan(stats1.mtimeMs);
    });

    it("should update atime when file is read", async () => {
      await fs.writeFile("test.txt", "content");
      const stats1 = await fs.stat("test.txt");

      await new Promise((resolve) => setTimeout(resolve, 10));

      await fs.readFile("test.txt");
      const stats2 = await fs.stat("test.txt");

      expect(stats2.atimeMs).toBeGreaterThan(stats1.atimeMs);
    });
  });
});
