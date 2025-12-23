/**
 * Test helper utilities for browser-git tests
 * Provides common test scenarios and utilities
 */

import { FileSystem } from "../src/filesystem/fs";
import { StorageAdapter } from "@browser-git/storage-adapters";

/**
 * Creates a test directory structure with files
 */
export async function createTestStructure(fs: FileSystem): Promise<void> {
  // Create directory structure
  await fs.mkdir("project/src", { recursive: true });
  await fs.mkdir("project/tests", { recursive: true });
  await fs.mkdir("project/docs", { recursive: true });

  // Create files
  await fs.writeFile("project/README.md", "# Test Project\n", {
    encoding: "utf8",
  });
  await fs.writeFile("project/src/index.ts", "export {};\n", {
    encoding: "utf8",
  });
  await fs.writeFile(
    "project/src/utils.ts",
    "export const add = (a: number, b: number) => a + b;\n",
    {
      encoding: "utf8",
    },
  );
  await fs.writeFile(
    "project/tests/index.test.ts",
    'describe("test", () => {});\n',
    {
      encoding: "utf8",
    },
  );
  await fs.writeFile("project/docs/API.md", "# API Documentation\n", {
    encoding: "utf8",
  });
}

/**
 * Creates a large file for performance testing
 */
export async function createLargeFile(
  fs: FileSystem,
  path: string,
  sizeInKB: number,
): Promise<void> {
  const chunkSize = 1024; // 1KB
  const content = "x".repeat(chunkSize);

  for (let i = 0; i < sizeInKB; i++) {
    if (i === 0) {
      await fs.writeFile(path, content, { encoding: "utf8" });
    } else {
      const existing = await fs.readFile(path, { encoding: "utf8" });
      await fs.writeFile(path, existing + content, { encoding: "utf8" });
    }
  }
}

/**
 * Creates a deep nested directory structure
 */
export async function createDeepStructure(
  fs: FileSystem,
  depth: number,
): Promise<void> {
  let path = "deep";
  for (let i = 0; i < depth; i++) {
    path += `/level${i}`;
  }
  await fs.mkdir(path, { recursive: true });
  await fs.writeFile(`${path}/file.txt`, "deep file", { encoding: "utf8" });
}

/**
 * Asserts that a file exists with specific content
 */
export async function assertFileContent(
  fs: FileSystem,
  path: string,
  expectedContent: string,
): Promise<void> {
  const exists = await fs.exists(path);
  if (!exists) {
    throw new Error(`File does not exist: ${path}`);
  }

  const content = await fs.readFile(path, { encoding: "utf8" });
  if (content !== expectedContent) {
    throw new Error(
      `File content mismatch.\nExpected: ${expectedContent}\nActual: ${content}`,
    );
  }
}

/**
 * Asserts that a directory exists and contains specific entries
 */
export async function assertDirectoryContains(
  fs: FileSystem,
  path: string,
  expectedEntries: string[],
): Promise<void> {
  const exists = await fs.exists(path);
  if (!exists) {
    throw new Error(`Directory does not exist: ${path}`);
  }

  const stat = await fs.stat(path);
  if (!stat.isDirectory) {
    throw new Error(`Path is not a directory: ${path}`);
  }

  const entries = await fs.readdir(path);
  for (const expected of expectedEntries) {
    if (!entries.includes(expected)) {
      throw new Error(`Directory ${path} does not contain: ${expected}`);
    }
  }
}

/**
 * Recursively deletes all files and directories
 */
export async function cleanupFileSystem(
  fs: FileSystem,
  rootPath: string = ".",
): Promise<void> {
  try {
    const entries = await fs.readdir(rootPath);
    for (const entry of entries) {
      const fullPath = rootPath === "." ? entry : `${rootPath}/${entry}`;
      const stat = await fs.stat(fullPath);
      if (stat.isDirectory) {
        await cleanupFileSystem(fs, fullPath);
        await fs.rmdir(fullPath);
      } else {
        await fs.unlink(fullPath);
      }
    }
  } catch (error) {
    // Ignore errors during cleanup
  }
}

/**
 * Measures execution time of an async function
 */
export async function measureTime<T>(
  fn: () => Promise<T>,
): Promise<{ result: T; time: number }> {
  const start = performance.now();
  const result = await fn();
  const time = performance.now() - start;
  return { result, time };
}

/**
 * Creates a mock file tree for testing
 */
export interface FileTree {
  [key: string]: string | FileTree;
}

/**
 * Recursively creates files and directories from a file tree object
 */
export async function createFileTree(
  fs: FileSystem,
  tree: FileTree,
  basePath: string = "",
): Promise<void> {
  for (const [name, content] of Object.entries(tree)) {
    const path = basePath ? `${basePath}/${name}` : name;

    if (typeof content === "string") {
      // It's a file
      await fs.writeFile(path, content, { encoding: "utf8" });
    } else {
      // It's a directory
      await fs.mkdir(path, { recursive: true });
      await createFileTree(fs, content, path);
    }
  }
}

/**
 * Gets all files recursively from a directory
 */
export async function getAllFiles(
  fs: FileSystem,
  path: string = ".",
): Promise<string[]> {
  const files: string[] = [];
  const entries = await fs.readdir(path);

  for (const entry of entries) {
    const fullPath = path === "." ? entry : `${path}/${entry}`;
    const stat = await fs.stat(fullPath);

    if (stat.isDirectory) {
      const subFiles = await getAllFiles(fs, fullPath);
      files.push(...subFiles);
    } else {
      files.push(fullPath);
    }
  }

  return files;
}

/**
 * Waits for a condition to be true
 */
export async function waitFor(
  condition: () => boolean | Promise<boolean>,
  timeout: number = 5000,
  interval: number = 100,
): Promise<void> {
  const start = Date.now();

  while (Date.now() - start < timeout) {
    if (await condition()) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  throw new Error(`Timeout waiting for condition after ${timeout}ms`);
}

/**
 * Creates a spy function for testing callbacks
 */
export function createSpy<T extends (...args: any[]) => any>(): T & {
  calls: Array<Parameters<T>>;
  callCount: number;
  reset: () => void;
} {
  const calls: Array<any[]> = [];

  const spy = ((...args: any[]) => {
    calls.push(args);
  }) as any;

  Object.defineProperty(spy, "calls", {
    get: () => calls,
  });

  Object.defineProperty(spy, "callCount", {
    get: () => calls.length,
  });

  spy.reset = () => {
    calls.length = 0;
  };

  return spy;
}

/**
 * Asserts that an error is thrown with a specific message
 */
export async function assertThrows(
  fn: () => Promise<any>,
  expectedError: string | RegExp,
): Promise<void> {
  try {
    await fn();
    throw new Error("Expected function to throw an error");
  } catch (error) {
    if (error instanceof Error) {
      if (typeof expectedError === "string") {
        if (!error.message.includes(expectedError)) {
          throw new Error(
            `Expected error message to include "${expectedError}", got "${error.message}"`,
          );
        }
      } else {
        if (!expectedError.test(error.message)) {
          throw new Error(
            `Expected error message to match ${expectedError}, got "${error.message}"`,
          );
        }
      }
    } else {
      throw new Error("Caught error is not an Error instance");
    }
  }
}
