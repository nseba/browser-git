/**
 * Path utilities unit tests
 */

import { describe, it, expect } from "vitest";
import {
  normalize,
  join,
  dirname,
  basename,
  extname,
  isAbsolute,
  resolve,
  relative,
  parse,
  format,
  sep,
} from "../src/filesystem/path.js";

describe("Path utilities", () => {
  describe("normalize", () => {
    it("should normalize simple paths", () => {
      expect(normalize("foo/bar")).toBe("foo/bar");
      expect(normalize("foo//bar")).toBe("foo/bar");
      expect(normalize("foo/./bar")).toBe("foo/bar");
    });

    it("should resolve .. segments", () => {
      expect(normalize("foo/bar/..")).toBe("foo");
      expect(normalize("foo/bar/../baz")).toBe("foo/baz");
      expect(normalize("foo/../bar")).toBe("bar");
    });

    it("should handle leading ..", () => {
      expect(normalize("../foo")).toBe("../foo");
      expect(normalize("../../foo")).toBe("../../foo");
    });

    it("should handle empty and current directory", () => {
      expect(normalize("")).toBe(".");
      expect(normalize(".")).toBe(".");
      expect(normalize("./")).toBe(".");
    });

    it("should remove trailing slashes", () => {
      expect(normalize("foo/bar/")).toBe("foo/bar");
      expect(normalize("foo/")).toBe("foo");
    });
  });

  describe("join", () => {
    it("should join path segments", () => {
      expect(join("foo", "bar")).toBe("foo/bar");
      expect(join("foo", "bar", "baz")).toBe("foo/bar/baz");
    });

    it("should normalize joined paths", () => {
      expect(join("foo", "../bar")).toBe("bar");
      expect(join("foo", "./bar")).toBe("foo/bar");
      expect(join("foo", "bar", "..")).toBe("foo");
    });

    it("should handle empty segments", () => {
      expect(join("foo", "", "bar")).toBe("foo/bar");
      expect(join("", "foo")).toBe("foo");
    });

    it("should return . for empty input", () => {
      expect(join()).toBe(".");
      expect(join("", "")).toBe(".");
    });
  });

  describe("dirname", () => {
    it("should return directory name", () => {
      expect(dirname("foo/bar/baz.txt")).toBe("foo/bar");
      expect(dirname("foo/bar")).toBe("foo");
      expect(dirname("foo")).toBe(".");
    });

    it("should handle root paths", () => {
      expect(dirname("/foo")).toBe(".");
      expect(dirname("/")).toBe(".");
    });

    it("should return . for empty path", () => {
      expect(dirname("")).toBe(".");
      expect(dirname(".")).toBe(".");
    });
  });

  describe("basename", () => {
    it("should return base name", () => {
      expect(basename("foo/bar/baz.txt")).toBe("baz.txt");
      expect(basename("foo/bar")).toBe("bar");
      expect(basename("foo")).toBe("foo");
    });

    it("should remove extension if provided", () => {
      expect(basename("foo/bar/baz.txt", ".txt")).toBe("baz");
      expect(basename("foo/bar.min.js", ".js")).toBe("bar.min");
    });

    it("should return empty for empty path", () => {
      expect(basename("")).toBe("");
    });
  });

  describe("extname", () => {
    it("should return file extension", () => {
      expect(extname("file.txt")).toBe(".txt");
      expect(extname("file.min.js")).toBe(".js");
      expect(extname("path/to/file.html")).toBe(".html");
    });

    it("should return empty for no extension", () => {
      expect(extname("file")).toBe("");
      expect(extname("path/to/file")).toBe("");
      expect(extname("")).toBe("");
    });

    it("should return empty for dotfiles", () => {
      expect(extname(".gitignore")).toBe("");
      expect(extname(".hidden")).toBe("");
    });
  });

  describe("isAbsolute", () => {
    it("should detect absolute paths", () => {
      expect(isAbsolute("/foo/bar")).toBe(true);
      expect(isAbsolute("/foo")).toBe(true);
      expect(isAbsolute("/")).toBe(true);
    });

    it("should detect relative paths", () => {
      expect(isAbsolute("foo/bar")).toBe(false);
      expect(isAbsolute("./foo")).toBe(false);
      expect(isAbsolute("../foo")).toBe(false);
      expect(isAbsolute("")).toBe(false);
    });
  });

  describe("resolve", () => {
    it("should resolve and normalize paths", () => {
      // Normalize removes leading slash, so paths become relative
      expect(resolve("foo", "bar")).toBe("foo/bar");
      expect(resolve("foo", "bar", "baz")).toBe("foo/bar/baz");
    });

    it("should handle absolute paths in sequence", () => {
      // When an absolute path is encountered, it becomes the base and iteration stops
      expect(resolve("foo", "/bar", "baz")).toBe("bar");
      expect(resolve("/foo", "/bar")).toBe("bar");
    });

    it("should normalize resolved paths", () => {
      expect(resolve("foo", "../bar")).toBe("bar");
      expect(resolve("foo", "./bar")).toBe("foo/bar");
    });

    it("should handle empty segments", () => {
      expect(resolve("foo", "", "bar")).toBe("foo/bar");
    });

    it("should handle relative paths", () => {
      expect(resolve("foo")).toBe("foo");
      expect(resolve("./foo")).toBe("foo");
    });
  });

  describe("relative", () => {
    it("should compute relative paths", () => {
      expect(relative("foo/bar", "foo/baz")).toBe("../baz");
      expect(relative("foo", "foo/bar")).toBe("bar");
      expect(relative("foo/bar/baz", "foo")).toBe("../..");
    });

    it("should return . for same paths", () => {
      expect(relative("foo", "foo")).toBe(".");
      expect(relative("foo/bar", "foo/bar")).toBe(".");
    });

    it("should handle paths with no common prefix", () => {
      expect(relative("foo/bar", "baz/qux")).toBe("../../baz/qux");
    });
  });

  describe("parse", () => {
    it("should parse file paths", () => {
      const parsed = parse("foo/bar/baz.txt");
      expect(parsed.root).toBe("");
      expect(parsed.dir).toBe("foo/bar");
      expect(parsed.base).toBe("baz.txt");
      expect(parsed.ext).toBe(".txt");
      expect(parsed.name).toBe("baz");
    });

    it("should parse absolute paths", () => {
      const parsed = parse("/foo/bar/baz.txt");
      expect(parsed.root).toBe("/");
      // After normalization, leading slash is removed
      expect(parsed.dir).toBe("foo/bar");
      expect(parsed.base).toBe("baz.txt");
    });

    it("should parse paths without extension", () => {
      const parsed = parse("foo/bar/file");
      expect(parsed.ext).toBe("");
      expect(parsed.name).toBe("file");
      expect(parsed.base).toBe("file");
    });

    it("should parse paths without directory", () => {
      const parsed = parse("file.txt");
      expect(parsed.dir).toBe("");
      expect(parsed.base).toBe("file.txt");
    });
  });

  describe("format", () => {
    it("should format path components", () => {
      expect(format({ dir: "foo/bar", base: "baz.txt" })).toBe(
        "foo/bar/baz.txt",
      );
      expect(format({ name: "file", ext: ".txt" })).toBe("file.txt");
    });

    it("should prefer base over name and ext", () => {
      expect(format({ base: "file.txt", name: "other", ext: ".js" })).toBe(
        "file.txt",
      );
    });

    it("should handle missing dir", () => {
      expect(format({ base: "file.txt" })).toBe("file.txt");
      expect(format({ name: "file", ext: ".txt" })).toBe("file.txt");
    });

    it("should handle empty components", () => {
      expect(format({})).toBe("");
      expect(format({ dir: "foo" })).toBe("foo");
    });
  });

  describe("sep", () => {
    it("should be forward slash", () => {
      expect(sep).toBe("/");
    });
  });

  describe("edge cases", () => {
    it("should handle complex paths", () => {
      const path = "foo/../bar/./baz/../qux.txt";
      expect(normalize(path)).toBe("bar/qux.txt");
    });

    it("should handle multiple consecutive slashes", () => {
      expect(normalize("foo///bar")).toBe("foo/bar");
      expect(join("foo", "///bar")).toBe("foo/bar");
    });

    it("should handle paths with only dots", () => {
      expect(normalize("./././")).toBe(".");
      expect(normalize("../..")).toBe("../..");
    });
  });
});
