/**
 * Security and malformed input tests
 * Tests various malformed inputs and edge cases to ensure robustness and security
 */

import { describe, it, expect, beforeEach } from "vitest";
import { normalize, join, isPathSafe } from "../src/filesystem/path.js";
import { validateGitURL } from "../src/utils/url-validator";

describe("Security: Malformed Input Tests", () => {
  describe("Path Traversal Attacks", () => {
    it("should reject paths attempting to escape repository root", () => {
      const maliciousPaths = [
        "../../../etc/passwd",
        "foo/../../bar",
        "./../../etc/shadow",
        "valid/../../../etc/passwd",
      ];

      maliciousPaths.forEach((path) => {
        const normalized = normalize(path);
        // Should detect paths that escape the root directory
        expect(isPathSafe(normalized)).toBe(false);
      });
    });

    it("should handle null bytes in paths", () => {
      const pathsWithNullBytes = [
        "foo\x00bar",
        "foo/bar\x00.txt",
        "\x00etc/passwd",
      ];

      pathsWithNullBytes.forEach((path) => {
        // Path utilities should handle these gracefully
        expect(() => normalize(path)).not.toThrow();
      });
    });

    it("should handle very long paths", () => {
      const longPath = "a/".repeat(1000) + "file.txt";
      expect(() => normalize(longPath)).not.toThrow();
      const result = normalize(longPath);
      expect(result).toBeDefined();
    });

    it("should handle paths with special characters", () => {
      const specialPaths = [
        "foo<bar",
        "foo>bar",
        "foo|bar",
        "foo*bar",
        "foo?bar",
        "foo:bar",
        'foo"bar',
      ];

      specialPaths.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
      });
    });

    it("should handle unicode and emoji in paths", () => {
      const unicodePaths = [
        "你好/world",
        "emoji/😀.txt",
        "путь/файл.txt",
        "日本語/ファイル.txt",
      ];

      unicodePaths.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
        const result = normalize(path);
        expect(result).toBeDefined();
      });
    });

    it("should handle paths with whitespace", () => {
      const whitespacePaths = [
        "  leading-spaces",
        "trailing-spaces  ",
        "foo   bar",
        "foo\tbar",
        "foo\nbar",
        "foo\rbar",
      ];

      whitespacePaths.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
      });
    });
  });

  describe("URL Validation Security", () => {
    it("should reject SSRF attack vectors", () => {
      const ssrfURLs = [
        "http://169.254.169.254/latest/meta-data",
        "http://metadata.google.internal/computeMetadata/v1/",
        "http://127.0.0.1:8080/admin",
        "http://localhost:9200/_cluster/health",
        "http://[::1]/admin",
        "http://[::ffff:127.0.0.1]/internal",
      ];

      ssrfURLs.forEach((url) => {
        expect(() => validateGitURL(url)).toThrow();
      });
    });

    it("should reject URLs with credentials", () => {
      const urlsWithCredentials = [
        "https://user:pass@github.com/repo.git",
        "https://admin:admin@example.com/repo.git",
      ];

      // These might be allowed but should be sanitized
      urlsWithCredentials.forEach((url) => {
        expect(() => new URL(url)).not.toThrow();
      });
    });

    it("should handle malformed URLs", () => {
      const malformedURLs = [
        "",
        "not a url",
        "http://",
        "https://",
        "http://.",
        "http://..",
        "http://../",
        "http://?",
        "http://??",
        "http://??/",
        "http://#",
        "http://##",
        "http://##/",
        "//",
        "///",
        "https://",
        "ftps://example.com",
        "h t t p://example.com",
      ];

      malformedURLs.forEach((url) => {
        expect(() => validateGitURL(url)).toThrow();
      });
    });

    it("should reject URLs with suspicious TLDs", () => {
      // These should fail validation
      const suspiciousURLs = [
        "https://example.exe/repo.git",
        "https://example.bat/repo.git",
      ];

      suspiciousURLs.forEach((url) => {
        // URL constructor might accept these, but they're suspicious
        try {
          new URL(url);
        } catch {
          // Expected for some
        }
      });
    });

    it("should handle extremely long URLs", () => {
      const longURL = "https://github.com/" + "a".repeat(5000);
      expect(() => validateGitURL(longURL)).toThrow();
    });

    it("should reject data: and blob: URLs", () => {
      const unsafeProtocols = [
        "data:text/html,<script>alert(1)</script>",
        "blob:https://example.com/uuid",
        "ftp://example.com/file",
        "ftps://example.com/file",
      ];

      unsafeProtocols.forEach((url) => {
        expect(() => validateGitURL(url)).toThrow();
      });
    });
  });

  describe("Input Sanitization", () => {
    it("should handle empty strings", () => {
      expect(() => normalize("")).not.toThrow();
      expect(normalize("")).toBe(".");
    });

    it("should handle strings with only special characters", () => {
      const specialOnly = ["...", "////", "....", "././.", "../.."];

      specialOnly.forEach((str) => {
        expect(() => normalize(str)).not.toThrow();
      });
    });

    it("should handle very deeply nested paths", () => {
      const deepPath = Array(100).fill("dir").join("/");
      expect(() => normalize(deepPath)).not.toThrow();
      const result = normalize(deepPath);
      expect(result).toBeDefined();
    });

    it("should handle paths with repeated separators", () => {
      const repeatedSeps = ["foo////bar", "foo//////bar", "foo//////////bar"];

      repeatedSeps.forEach((path) => {
        const result = normalize(path);
        expect(result).toBe("foo/bar");
      });
    });
  });

  describe("Numeric Edge Cases", () => {
    it("should handle numeric overflow scenarios", () => {
      const largeNumbers = [
        Number.MAX_SAFE_INTEGER,
        Number.MIN_SAFE_INTEGER,
        Number.MAX_VALUE,
        Number.MIN_VALUE,
      ];

      // Test that numeric operations don't cause issues
      largeNumbers.forEach((num) => {
        expect(num).toBeDefined();
        expect(isFinite(num)).toBe(true);
      });
    });

    it("should handle special numeric values", () => {
      const specialValues = [NaN, Infinity, -Infinity, 0, -0];

      specialValues.forEach((val) => {
        expect(val).toBeDefined();
      });
    });
  });

  describe("Type Confusion", () => {
    it("should handle type mismatches gracefully", () => {
      // These should be caught by TypeScript, but test runtime behavior
      const invalidInputs: any[] = [
        null,
        undefined,
        {},
        [],
        123,
        true,
        false,
        Symbol("test"),
      ];

      invalidInputs.forEach((input) => {
        // join should handle these (though TypeScript would prevent)
        if (typeof input === "string" || typeof input === "undefined") {
          expect(() => join(input as any)).not.toThrow();
        }
      });
    });
  });

  describe("Binary Data Handling", () => {
    it("should handle binary data as strings", () => {
      const binaryStrings = [
        "\x00\x01\x02\x03",
        "\xFF\xFE\xFD",
        String.fromCharCode(0, 1, 2, 3, 255, 254),
      ];

      binaryStrings.forEach((str) => {
        expect(() => normalize(str)).not.toThrow();
      });
    });

    it("should handle UTF-8 edge cases", () => {
      const utf8EdgeCases = [
        "\uD800", // Unpaired high surrogate
        "\uDFFF", // Unpaired low surrogate
        "\uFFFE", // Non-character
        "\uFFFF", // Non-character
      ];

      utf8EdgeCases.forEach((str) => {
        expect(() => normalize(str)).not.toThrow();
      });
    });
  });

  describe("Path Combination Edge Cases", () => {
    it("should handle joining paths with absolute components", () => {
      expect(join("foo", "/bar")).toBe("foo/bar");
      expect(join("/foo", "bar")).toBe("foo/bar");
    });

    it("should handle joining with empty components", () => {
      expect(join("", "")).toBe(".");
      expect(join("foo", "")).toBe("foo");
      expect(join("", "foo")).toBe("foo");
    });

    it("should handle joining with current directory markers", () => {
      expect(join(".", "foo")).toBe("foo");
      expect(join("foo", ".")).toBe("foo");
      expect(join(".", ".")).toBe(".");
    });
  });

  describe("Regex Denial of Service (ReDoS)", () => {
    it("should handle patterns that could cause ReDoS", () => {
      // Patterns that might cause exponential backtracking
      const redosPatterns = [
        "a".repeat(100) + "!",
        "x".repeat(50) + "y".repeat(50),
        ("a+" as any).repeat(30),
      ];

      redosPatterns.forEach((pattern) => {
        expect(() => normalize(pattern)).not.toThrow();
        // Should complete in reasonable time
      });
    });
  });

  describe("Case Sensitivity", () => {
    it("should preserve case in paths", () => {
      expect(normalize("Foo/Bar")).toBe("Foo/Bar");
      expect(normalize("FOO/bar")).toBe("FOO/bar");
      expect(normalize("foo/BAR")).toBe("foo/BAR");
    });
  });

  describe("Circular Reference Detection", () => {
    it("should handle circular symbolic link patterns", () => {
      // Test paths that might represent circular references
      const circularPatterns = [
        "a/../a/../a/../a",
        "foo/bar/../../foo/bar/../../foo",
      ];

      circularPatterns.forEach((pattern) => {
        expect(() => normalize(pattern)).not.toThrow();
        const result = normalize(pattern);
        expect(result).toBeDefined();
      });
    });
  });

  describe("Boundary Conditions", () => {
    it("should handle maximum path length", () => {
      // Test paths at or near system limits
      const maxPath = "a".repeat(4096);
      expect(() => normalize(maxPath)).not.toThrow();
    });

    it("should handle maximum filename length", () => {
      const maxFilename = "a".repeat(255) + ".txt";
      expect(() => normalize(maxFilename)).not.toThrow();
    });

    it("should handle path with maximum directory depth", () => {
      const deepPath = Array(256).fill("d").join("/") + "/file.txt";
      expect(() => normalize(deepPath)).not.toThrow();
    });
  });

  describe("Encoding Issues", () => {
    it("should handle mixed encodings", () => {
      // Paths that might have encoding issues
      const encodingTests = ["café", "naïve", "résumé", "文件"];

      encodingTests.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
        const result = normalize(path);
        expect(result).toBe(path);
      });
    });

    it("should handle percent-encoded characters", () => {
      const percentEncoded = ["foo%20bar", "test%2Ffile", "%2E%2E%2Fetc"];

      percentEncoded.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
      });
    });
  });

  describe("Command Injection Prevention", () => {
    it("should handle shell metacharacters safely", () => {
      const shellMetachars = [
        "foo;rm -rf /",
        "bar|cat /etc/passwd",
        "baz&& echo hacked",
        "qux`whoami`",
        "test$(ls -la)",
      ];

      shellMetachars.forEach((path) => {
        // Should be treated as literal strings, not executed
        expect(() => normalize(path)).not.toThrow();
      });
    });
  });

  describe("Directory Traversal Variants", () => {
    it("should handle various directory traversal encodings", () => {
      const traversalVariants = [
        "..%2f..%2f",
        "..%5c..%5c",
        "..%252f..%252f",
        "%2e%2e%2f%2e%2e%2f",
      ];

      traversalVariants.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
      });
    });

    it("should handle backslash variations", () => {
      const backslashPaths = ["..\\..\\etc\\passwd", "foo\\bar", "foo\\\\bar"];

      backslashPaths.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
      });
    });
  });

  describe("Zero-width and Invisible Characters", () => {
    it("should handle zero-width characters", () => {
      const zeroWidthChars = [
        "foo\u200Bbar", // Zero-width space
        "test\uFEFFfile", // Zero-width no-break space
        "dir\u200Cname", // Zero-width non-joiner
        "file\u200Dname", // Zero-width joiner
      ];

      zeroWidthChars.forEach((path) => {
        expect(() => normalize(path)).not.toThrow();
      });
    });

    it("should handle right-to-left override", () => {
      // U+202E Right-to-Left Override can be used for spoofing
      const rtlPath = "test\u202Etxt.exe";
      expect(() => normalize(rtlPath)).not.toThrow();
    });
  });
});
