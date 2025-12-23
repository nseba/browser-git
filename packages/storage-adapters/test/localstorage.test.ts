import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { LocalStorageAdapter } from "../src/localstorage.js";
import { StorageError, StorageErrorCode } from "../src/interface.js";

describe("LocalStorageAdapter", () => {
  let adapter: LocalStorageAdapter;
  const testDbName = "test-db";

  beforeEach(() => {
    adapter = new LocalStorageAdapter(testDbName);
    // Clear localStorage before each test
    localStorage.clear();
  });

  afterEach(async () => {
    await adapter.clear();
  });

  describe("isSupported", () => {
    it("should check if localStorage is supported", () => {
      expect(LocalStorageAdapter.isSupported()).toBe(true);
    });
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

      let keys = await adapter.list();
      expect(keys).toHaveLength(3);

      await adapter.clear();

      keys = await adapter.list();
      expect(keys).toHaveLength(0);
    });

    it("should only clear data with the adapter prefix", async () => {
      // Add data with this adapter
      await adapter.set("key1", new Uint8Array([1]));

      // Add data with different prefix (simulating another adapter)
      localStorage.setItem("other-prefix:key2", "value");

      await adapter.clear();

      // Our data should be cleared
      expect(await adapter.exists("key1")).toBe(false);

      // Other data should remain
      expect(localStorage.getItem("other-prefix:key2")).toBe("value");

      // Cleanup
      localStorage.removeItem("other-prefix:key2");
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

    it("should handle moderate binary data", async () => {
      const key = "moderate";
      const value = new Uint8Array(1000);
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

      expect(quota).not.toBeNull();
      expect(quota).toHaveProperty("usage");
      expect(quota).toHaveProperty("quota");
      expect(quota).toHaveProperty("percentage");
      expect(typeof quota!.usage).toBe("number");
      expect(typeof quota!.quota).toBe("number");
      expect(typeof quota!.percentage).toBe("number");
    });

    it("should update usage estimate after storing data", async () => {
      const quota1 = await adapter.getQuota();
      const initialUsage = quota1!.usage;

      // Add some data
      await adapter.set("test", new Uint8Array([1, 2, 3]));

      const quota2 = await adapter.getQuota();
      const newUsage = quota2!.usage;

      expect(newUsage).toBeGreaterThan(initialUsage);
    });
  });

  describe("encoding/decoding", () => {
    it("should correctly encode and decode data", async () => {
      const testCases = [
        new Uint8Array([0]),
        new Uint8Array([255]),
        new Uint8Array([0, 128, 255]),
        new Uint8Array([1, 2, 3, 4, 5]),
      ];

      for (const testCase of testCases) {
        await adapter.set("test", testCase);
        const result = await adapter.get("test");
        expect(result).toEqual(testCase);
      }
    });
  });

  describe("isolation", () => {
    it("should isolate data between different adapter instances", async () => {
      const adapter1 = new LocalStorageAdapter("db1");
      const adapter2 = new LocalStorageAdapter("db2");

      await adapter1.set("key", new Uint8Array([1]));
      await adapter2.set("key", new Uint8Array([2]));

      const value1 = await adapter1.get("key");
      const value2 = await adapter2.get("key");

      expect(value1).toEqual(new Uint8Array([1]));
      expect(value2).toEqual(new Uint8Array([2]));

      // Cleanup
      await adapter1.clear();
      await adapter2.clear();
    });
  });
});
