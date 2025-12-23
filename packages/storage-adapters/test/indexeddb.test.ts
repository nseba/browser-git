import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { IndexedDBAdapter } from "../src/indexeddb.js";
import { StorageError, StorageErrorCode } from "../src/interface.js";

describe("IndexedDBAdapter", () => {
  let adapter: IndexedDBAdapter;
  const testDbName = "test-db";

  beforeEach(() => {
    adapter = new IndexedDBAdapter(testDbName);
  });

  afterEach(async () => {
    // Clean up
    await adapter.clear();
    adapter.close();

    // Delete the database
    if (typeof indexedDB !== "undefined") {
      indexedDB.deleteDatabase(`bg-${testDbName}`);
    }
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
      // Set up test data
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

      let keys = await adapter.list();
      expect(keys).toHaveLength(3);

      await adapter.clear();

      keys = await adapter.list();
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
      const value = new Uint8Array(10000);
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
      expect(Array.from(retrieved!)).toEqual([0, 255, 128, 1, 254, 127]);
    });
  });

  describe("quota management", () => {
    it("should get quota information", async () => {
      const quota = await adapter.getQuota();

      // Quota might not be available in all environments
      if (quota) {
        expect(quota).toHaveProperty("usage");
        expect(quota).toHaveProperty("quota");
        expect(quota).toHaveProperty("percentage");
        expect(typeof quota.usage).toBe("number");
        expect(typeof quota.quota).toBe("number");
        expect(typeof quota.percentage).toBe("number");
      }
    });
  });

  describe("concurrent operations", () => {
    it("should handle concurrent writes", async () => {
      const promises = [];
      for (let i = 0; i < 10; i++) {
        promises.push(adapter.set(`key${i}`, new Uint8Array([i])));
      }

      await Promise.all(promises);

      const keys = await adapter.list();
      expect(keys).toHaveLength(10);
    });

    it("should handle concurrent reads", async () => {
      // Set up data
      for (let i = 0; i < 5; i++) {
        await adapter.set(`key${i}`, new Uint8Array([i]));
      }

      // Concurrent reads
      const promises = [];
      for (let i = 0; i < 5; i++) {
        promises.push(adapter.get(`key${i}`));
      }

      const results = await Promise.all(promises);
      expect(results).toHaveLength(5);
      results.forEach((result, i) => {
        expect(result).toEqual(new Uint8Array([i]));
      });
    });
  });

  describe("error handling", () => {
    it("should handle invalid operations gracefully", async () => {
      await adapter.delete("non-existent-key"); // Should not throw
    });
  });

  describe("reinitialization", () => {
    it("should persist data across adapter instances", async () => {
      const key = "persistent-key";
      const value = new Uint8Array([1, 2, 3]);

      await adapter.set(key, value);
      adapter.close();

      // Create new adapter with same name
      const newAdapter = new IndexedDBAdapter(testDbName);
      const retrieved = await newAdapter.get(key);

      expect(retrieved).toEqual(value);

      newAdapter.close();
    });
  });
});
