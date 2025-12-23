/**
 * Performance benchmarks for filesystem operations
 */

import { bench, describe } from "vitest";
import { FileSystem } from "@browser-git/browser-git";
import { MemoryAdapter } from "@browser-git/storage-adapters";

describe("FileSystem - File Operations", () => {
  bench("write small file (1KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
  });

  bench("write medium file (10KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(10 * 1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
  });

  bench("write large file (100KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(100 * 1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
  });

  bench("read small file (1KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
    await fs.readFile("test.txt", { encoding: "utf8" });
  });

  bench("read medium file (10KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(10 * 1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
    await fs.readFile("test.txt", { encoding: "utf8" });
  });

  bench("read large file (100KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(100 * 1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
    await fs.readFile("test.txt", { encoding: "utf8" });
  });
});

describe("FileSystem - Directory Operations", () => {
  bench("create directory", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.mkdir("test-dir");
  });

  bench("create nested directories (3 levels)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.mkdir("level1/level2/level3", { recursive: true });
  });

  bench("create nested directories (10 levels)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const path = Array.from({ length: 10 }, (_, i) => `level${i}`).join("/");
    await fs.mkdir(path, { recursive: true });
  });

  bench("read directory with 10 files", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    for (let i = 0; i < 10; i++) {
      await fs.writeFile(`file${i}.txt`, "content", { encoding: "utf8" });
    }
    await fs.readdir(".");
  });

  bench("read directory with 100 files", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    for (let i = 0; i < 100; i++) {
      await fs.writeFile(`file${i}.txt`, "content", { encoding: "utf8" });
    }
    await fs.readdir(".");
  });
});

describe("FileSystem - Stat Operations", () => {
  bench("stat file", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.writeFile("test.txt", "content", { encoding: "utf8" });
    await fs.stat("test.txt");
  });

  bench("stat directory", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.mkdir("test-dir");
    await fs.stat("test-dir");
  });

  bench("exists check (file exists)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.writeFile("test.txt", "content", { encoding: "utf8" });
    await fs.exists("test.txt");
  });

  bench("exists check (file does not exist)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.exists("non-existent.txt");
  });
});

describe("FileSystem - Delete Operations", () => {
  bench("delete small file", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.writeFile("test.txt", "content", { encoding: "utf8" });
    await fs.unlink("test.txt");
  });

  bench("delete large file (100KB)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    const content = "x".repeat(100 * 1024);
    await fs.writeFile("test.txt", content, { encoding: "utf8" });
    await fs.unlink("test.txt");
  });

  bench("delete directory", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.mkdir("test-dir");
    await fs.rmdir("test-dir");
  });

  bench("delete directory tree (3 levels, 10 files)", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    await fs.mkdir("level1/level2/level3", { recursive: true });
    for (let i = 0; i < 10; i++) {
      await fs.writeFile(`level1/level2/level3/file${i}.txt`, "content", {
        encoding: "utf8",
      });
    }
    await fs.rmdir("level1", { recursive: true });
  });
});

describe("FileSystem - Bulk Operations", () => {
  bench("write 100 small files", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    for (let i = 0; i < 100; i++) {
      await fs.writeFile(`file${i}.txt`, "content", { encoding: "utf8" });
    }
  });

  bench("read 100 small files", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    for (let i = 0; i < 100; i++) {
      await fs.writeFile(`file${i}.txt`, "content", { encoding: "utf8" });
    }
    for (let i = 0; i < 100; i++) {
      await fs.readFile(`file${i}.txt`, { encoding: "utf8" });
    }
  });

  bench("stat 100 files", async () => {
    const adapter = new MemoryAdapter();
    const fs = new FileSystem(adapter);
    for (let i = 0; i < 100; i++) {
      await fs.writeFile(`file${i}.txt`, "content", { encoding: "utf8" });
    }
    for (let i = 0; i < 100; i++) {
      await fs.stat(`file${i}.txt`);
    }
  });
});
