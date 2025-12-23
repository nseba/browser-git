import { describe, it, expect, beforeEach } from "vitest";
import { MemoryAdapter } from "../src/memory.js";

describe("MemoryAdapter", () => {
  let adapter: MemoryAdapter;

  beforeEach(() => {
    adapter = new MemoryAdapter();
  });

  describe("basic operations", () => {
    it("should store and retrieve data", async () => {
      const key = "test-key";
      const value = new Uint8Array([1, 2, 3, 4, 5]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });

    it("should return null for non-existent key", async () => {
      const result = await adapter.get("non-existent");
      expect(result).toBeNull();
    });

    it("should delete data", async () => {
      const key = "test-key";
      const value = new Uint8Array([1, 2, 3]);

      await adapter.set(key, value);
      await adapter.delete(key);

      const retrieved = await adapter.get(key);
      expect(retrieved).toBeNull();
    });

    it("should check if key exists", async () => {
      const key = "test-key";
      const value = new Uint8Array([1, 2, 3]);

      expect(await adapter.exists(key)).toBe(false);

      await adapter.set(key, value);
      expect(await adapter.exists(key)).toBe(true);

      await adapter.delete(key);
      expect(await adapter.exists(key)).toBe(false);
    });
  });

  describe("list operations", () => {
    beforeEach(async () => {
      await adapter.set("file1.txt", new Uint8Array([1]));
      await adapter.set("file2.txt", new Uint8Array([2]));
      await adapter.set("dir/file3.txt", new Uint8Array([3]));
      await adapter.set("other.dat", new Uint8Array([4]));
    });

    it("should list all keys", async () => {
      const keys = await adapter.list();
      expect(keys).toHaveLength(4);
      expect(keys).toContain("file1.txt");
      expect(keys).toContain("file2.txt");
      expect(keys).toContain("dir/file3.txt");
      expect(keys).toContain("other.dat");
    });

    it("should list keys with prefix", async () => {
      const keys = await adapter.list("file");
      expect(keys).toHaveLength(2);
      expect(keys).toContain("file1.txt");
      expect(keys).toContain("file2.txt");
    });

    it("should list keys with directory prefix", async () => {
      const keys = await adapter.list("dir/");
      expect(keys).toHaveLength(1);
      expect(keys).toContain("dir/file3.txt");
    });

    it("should return empty array for non-matching prefix", async () => {
      const keys = await adapter.list("nonexistent");
      expect(keys).toHaveLength(0);
    });
  });

  describe("clear operation", () => {
    it("should clear all data", async () => {
      await adapter.set("key1", new Uint8Array([1]));
      await adapter.set("key2", new Uint8Array([2]));
      await adapter.set("key3", new Uint8Array([3]));

      expect(adapter.size()).toBe(3);

      await adapter.clear();

      expect(adapter.size()).toBe(0);
      const keys = await adapter.list();
      expect(keys).toHaveLength(0);
    });
  });

  describe("binary data", () => {
    it("should handle empty Uint8Array", async () => {
      const key = "empty";
      const value = new Uint8Array([]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
      expect(retrieved?.length).toBe(0);
    });

    it("should handle large binary data", async () => {
      const key = "large";
      const value = new Uint8Array(100000);
      for (let i = 0; i < value.length; i++) {
        value[i] = i % 256;
      }

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });

    it("should preserve binary data integrity", async () => {
      const key = "binary";
      const value = new Uint8Array([0, 255, 128, 1, 254, 127]);

      await adapter.set(key, value);
      const retrieved = await adapter.get(key);

      expect(retrieved).toEqual(value);
    });
  });

  describe("data isolation", () => {
    it("should return copies of stored data", async () => {
      const key = "test";
      const original = new Uint8Array([1, 2, 3]);

      await adapter.set(key, original);

      // Modify original
      original[0] = 99;

      // Stored value should be unchanged
      const retrieved = await adapter.get(key);
      expect(retrieved).toEqual(new Uint8Array([1, 2, 3]));
    });

    it("should prevent external modification of retrieved data", async () => {
      const key = "test";
      await adapter.set(key, new Uint8Array([1, 2, 3]));

      const retrieved = await adapter.get(key);
      retrieved![0] = 99;

      // Original stored value should be unchanged
      const retrieved2 = await adapter.get(key);
      expect(retrieved2).toEqual(new Uint8Array([1, 2, 3]));
    });
  });

  describe("quota management", () => {
    it("should get quota information", async () => {
      const quota = await adapter.getQuota();

      expect(quota).not.toBeNull();
      expect(quota).toHaveProperty("usage");
      expect(quota).toHaveProperty("quota");
      expect(quota).toHaveProperty("percentage");
    });

    it("should track usage", async () => {
      const quota1 = await adapter.getQuota();
      const initialUsage = quota1!.usage;

      await adapter.set("test", new Uint8Array([1, 2, 3, 4, 5]));

      const quota2 = await adapter.getQuota();
      const newUsage = quota2!.usage;

      expect(newUsage).toBeGreaterThan(initialUsage);
    });
  });

  describe("convenience methods", () => {
    beforeEach(async () => {
      await adapter.set("key1", new Uint8Array([1]));
      await adapter.set("key2", new Uint8Array([2]));
      await adapter.set("key3", new Uint8Array([3]));
    });

    it("should return size", () => {
      expect(adapter.size()).toBe(3);
    });

    it("should return all keys synchronously", () => {
      const keys = adapter.keys();
      expect(keys).toHaveLength(3);
      expect(keys).toContain("key1");
      expect(keys).toContain("key2");
      expect(keys).toContain("key3");
    });

    it("should return all entries synchronously", () => {
      const entries = adapter.entries();
      expect(entries).toHaveLength(3);

      const entryMap = new Map(entries);
      expect(entryMap.get("key1")).toEqual(new Uint8Array([1]));
      expect(entryMap.get("key2")).toEqual(new Uint8Array([2]));
      expect(entryMap.get("key3")).toEqual(new Uint8Array([3]));
    });
  });

  describe("concurrent operations", () => {
    it("should handle concurrent writes", async () => {
      const promises = [];
      for (let i = 0; i < 100; i++) {
        promises.push(adapter.set(`key${i}`, new Uint8Array([i])));
      }

      await Promise.all(promises);

      expect(adapter.size()).toBe(100);
    });

    it("should handle concurrent reads", async () => {
      // Setup
      for (let i = 0; i < 10; i++) {
        await adapter.set(`key${i}`, new Uint8Array([i]));
      }

      // Concurrent reads
      const promises = [];
      for (let i = 0; i < 10; i++) {
        promises.push(adapter.get(`key${i}`));
      }

      const results = await Promise.all(promises);
      expect(results).toHaveLength(10);
      results.forEach((result, i) => {
        expect(result).toEqual(new Uint8Array([i]));
      });
    });
  });
});
