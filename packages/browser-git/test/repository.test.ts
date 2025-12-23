/**
 * Repository API tests
 */

import { describe, it, expect, beforeEach } from "vitest";
import {
  Repository,
  CloneOptions,
  InitOptions,
  GitError,
  CloneError,
} from "../src/repository.js";
import { AuthMethod } from "../src/types/auth.js";

describe("Repository", () => {
  describe("CloneOptions", () => {
    it("should have correct default values", () => {
      const opts: CloneOptions = {};

      expect(opts.bare).toBeUndefined();
      expect(opts.depth).toBeUndefined();
      expect(opts.branch).toBeUndefined();
      expect(opts.remote).toBeUndefined();
    });

    it("should accept custom values", () => {
      const opts: CloneOptions = {
        bare: true,
        depth: 1,
        branch: "develop",
        remote: "upstream",
        auth: {
          method: AuthMethod.Token,
          token: "test-token",
        },
      };

      expect(opts.bare).toBe(true);
      expect(opts.depth).toBe(1);
      expect(opts.branch).toBe("develop");
      expect(opts.remote).toBe("upstream");
      expect(opts.auth).toBeDefined();
      expect(opts.auth?.method).toBe(AuthMethod.Token);
    });

    it("should accept progress callback", () => {
      const messages: string[] = [];
      const opts: CloneOptions = {
        onProgress: (msg) => messages.push(msg),
      };

      expect(opts.onProgress).toBeDefined();
      opts.onProgress!("test message");
      expect(messages).toContain("test message");
    });
  });

  describe("InitOptions", () => {
    it("should accept all valid options", () => {
      const opts: InitOptions = {
        bare: false,
        initialBranch: "main",
        hashAlgorithm: "sha1",
      };

      expect(opts.bare).toBe(false);
      expect(opts.initialBranch).toBe("main");
      expect(opts.hashAlgorithm).toBe("sha1");
    });

    it("should accept sha256 hash algorithm", () => {
      const opts: InitOptions = {
        hashAlgorithm: "sha256",
      };

      expect(opts.hashAlgorithm).toBe("sha256");
    });
  });

  describe("Repository.clone", () => {
    it("should reject clone with WASM not integrated error", async () => {
      const url = "https://github.com/user/repo.git";
      const path = "./test-repo";

      await expect(Repository.clone(url, path)).rejects.toThrow(CloneError);
      await expect(Repository.clone(url, path)).rejects.toThrow(
        "WASM not yet integrated",
      );
    });

    it("should accept valid clone options", async () => {
      const url = "https://github.com/user/repo.git";
      const path = "./test-repo";
      const opts: CloneOptions = {
        bare: false,
        depth: 0,
        branch: "main",
        remote: "origin",
        auth: {
          method: AuthMethod.Token,
          token: "test-token",
        },
        onProgress: (msg) => console.log(msg),
      };

      // Will fail with WASM not integrated, but validates options are accepted
      await expect(Repository.clone(url, path, opts)).rejects.toThrow();
    });
  });

  describe("Repository.init", () => {
    it("should reject init with WASM not integrated error", async () => {
      const path = "./test-repo";

      await expect(Repository.init(path)).rejects.toThrow(GitError);
      await expect(Repository.init(path)).rejects.toThrow(
        "WASM not yet integrated",
      );
    });

    it("should accept valid init options", async () => {
      const path = "./test-repo";
      const opts: InitOptions = {
        bare: false,
        initialBranch: "main",
        hashAlgorithm: "sha1",
      };

      // Will fail with WASM not integrated, but validates options are accepted
      await expect(Repository.init(path, opts)).rejects.toThrow();
    });
  });

  describe("Repository.open", () => {
    it("should reject open with WASM not integrated error", async () => {
      const path = "./test-repo";

      await expect(Repository.open(path)).rejects.toThrow(GitError);
      await expect(Repository.open(path)).rejects.toThrow(
        "WASM not yet integrated",
      );
    });
  });

  describe("GitError", () => {
    it("should create error with message", () => {
      const error = new GitError("Test error");

      expect(error).toBeInstanceOf(Error);
      expect(error.name).toBe("GitError");
      expect(error.message).toBe("Test error");
      expect(error.cause).toBeUndefined();
    });

    it("should create error with cause", () => {
      const cause = new Error("Original error");
      const error = new GitError("Test error", cause);

      expect(error.cause).toBe(cause);
    });
  });

  describe("CloneError", () => {
    it("should create error with URL", () => {
      const url = "https://github.com/user/repo.git";
      const error = new CloneError("Test error", url);

      expect(error).toBeInstanceOf(GitError);
      expect(error).toBeInstanceOf(Error);
      expect(error.name).toBe("CloneError");
      expect(error.message).toBe("Test error");
      expect(error.url).toBe(url);
    });

    it("should create error with cause", () => {
      const url = "https://github.com/user/repo.git";
      const cause = new Error("Original error");
      const error = new CloneError("Test error", url, cause);

      expect(error.cause).toBe(cause);
    });
  });
});
