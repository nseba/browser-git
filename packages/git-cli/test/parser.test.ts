import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  parseGlobPatterns,
  parseKeyValuePairs,
  parseAuthor,
  formatAuthor,
  parseDate,
  formatDate,
  formatRelativeDate,
  shortHash,
  truncate,
} from "../src/utils/parser.js";

describe("parser utilities", () => {
  describe("parseGlobPatterns", () => {
    it("should return empty array for empty input", () => {
      expect(parseGlobPatterns([])).toEqual([]);
    });

    it("should pass through simple file paths", () => {
      expect(parseGlobPatterns(["file.txt"])).toEqual(["file.txt"]);
      expect(parseGlobPatterns(["src/index.ts"])).toEqual(["src/index.ts"]);
    });

    it("should handle multiple paths", () => {
      expect(parseGlobPatterns(["file1.txt", "file2.txt"])).toEqual([
        "file1.txt",
        "file2.txt",
      ]);
    });

    it("should unescape asterisk characters", () => {
      expect(parseGlobPatterns(["\\*.txt"])).toEqual(["*.txt"]);
      expect(parseGlobPatterns(["src/\\*.ts"])).toEqual(["src/*.ts"]);
    });

    it("should handle multiple escaped asterisks", () => {
      expect(parseGlobPatterns(["\\*/\\*.ts"])).toEqual(["*/*.ts"]);
    });

    it("should handle patterns with no escaping needed", () => {
      expect(parseGlobPatterns(["*.txt"])).toEqual(["*.txt"]);
      expect(parseGlobPatterns(["**/*.ts"])).toEqual(["**/*.ts"]);
    });
  });

  describe("parseKeyValuePairs", () => {
    it("should return empty object for empty input", () => {
      expect(parseKeyValuePairs([])).toEqual({});
    });

    it("should parse single key=value pair", () => {
      expect(parseKeyValuePairs(["key=value"])).toEqual({ key: "value" });
    });

    it("should parse multiple key=value pairs", () => {
      expect(parseKeyValuePairs(["foo=bar", "baz=qux"])).toEqual({
        foo: "bar",
        baz: "qux",
      });
    });

    it("should handle values with equals signs", () => {
      expect(parseKeyValuePairs(["key=value=with=equals"])).toEqual({
        key: "value=with=equals",
      });
    });

    it("should ignore invalid entries without equals sign", () => {
      expect(parseKeyValuePairs(["noequals", "key=value"])).toEqual({
        key: "value",
      });
    });

    it("should ignore empty keys", () => {
      expect(parseKeyValuePairs(["=value"])).toEqual({});
    });

    it("should handle special characters in values", () => {
      expect(parseKeyValuePairs(["path=/usr/local/bin"])).toEqual({
        path: "/usr/local/bin",
      });
      expect(parseKeyValuePairs(["url=https://example.com"])).toEqual({
        url: "https://example.com",
      });
    });
  });

  describe("parseAuthor", () => {
    it("should parse valid author string", () => {
      expect(parseAuthor("John Doe <john@example.com>")).toEqual({
        name: "John Doe",
        email: "john@example.com",
      });
    });

    it("should handle extra whitespace", () => {
      expect(parseAuthor("  John Doe   <john@example.com>")).toEqual({
        name: "John Doe",
        email: "john@example.com",
      });
    });

    it("should handle whitespace inside brackets", () => {
      expect(parseAuthor("John Doe < john@example.com >")).toEqual({
        name: "John Doe",
        email: "john@example.com",
      });
    });

    it("should return null for invalid format - missing brackets", () => {
      expect(parseAuthor("John Doe john@example.com")).toBeNull();
    });

    it("should return null for missing email", () => {
      expect(parseAuthor("John Doe <>")).toBeNull();
    });

    it("should return null for missing name", () => {
      expect(parseAuthor("<john@example.com>")).toBeNull();
    });

    it("should return null for empty string", () => {
      expect(parseAuthor("")).toBeNull();
    });

    it("should handle names with special characters", () => {
      expect(parseAuthor("O'Brien <obrien@example.com>")).toEqual({
        name: "O'Brien",
        email: "obrien@example.com",
      });
    });
  });

  describe("formatAuthor", () => {
    it("should format author object as string", () => {
      expect(formatAuthor({ name: "John Doe", email: "john@example.com" })).toBe(
        "John Doe <john@example.com>"
      );
    });

    it("should handle empty name", () => {
      expect(formatAuthor({ name: "", email: "john@example.com" })).toBe(
        " <john@example.com>"
      );
    });

    it("should handle special characters", () => {
      expect(formatAuthor({ name: "O'Brien", email: "o@example.com" })).toBe(
        "O'Brien <o@example.com>"
      );
    });
  });

  describe("parseDate", () => {
    it("should parse ISO date string", () => {
      const date = parseDate("2024-01-15T12:00:00Z");
      expect(date).toBeInstanceOf(Date);
      expect(date.toISOString()).toBe("2024-01-15T12:00:00.000Z");
    });

    it("should parse date-only string", () => {
      const date = parseDate("2024-01-15");
      expect(date).toBeInstanceOf(Date);
    });

    it("should handle invalid date string", () => {
      const date = parseDate("not-a-date");
      expect(date.toString()).toBe("Invalid Date");
    });
  });

  describe("formatDate", () => {
    it("should format date to locale string", () => {
      const date = new Date("2024-01-15T12:00:00Z");
      const formatted = formatDate(date);
      expect(typeof formatted).toBe("string");
      expect(formatted.length).toBeGreaterThan(0);
    });
  });

  describe("formatRelativeDate", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('should format seconds ago (singular)', () => {
      const now = new Date("2024-01-15T12:00:01Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 second ago");
    });

    it('should format seconds ago (plural)', () => {
      const now = new Date("2024-01-15T12:00:30Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("30 seconds ago");
    });

    it('should format minutes ago (singular)', () => {
      const now = new Date("2024-01-15T12:01:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 minute ago");
    });

    it('should format minutes ago (plural)', () => {
      const now = new Date("2024-01-15T12:30:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("30 minutes ago");
    });

    it('should format hours ago (singular)', () => {
      const now = new Date("2024-01-15T13:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 hour ago");
    });

    it('should format hours ago (plural)', () => {
      const now = new Date("2024-01-15T17:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("5 hours ago");
    });

    it('should format days ago (singular)', () => {
      const now = new Date("2024-01-16T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 day ago");
    });

    it('should format days ago (plural)', () => {
      const now = new Date("2024-01-18T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("3 days ago");
    });

    it('should format weeks ago (singular)', () => {
      const now = new Date("2024-01-22T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 week ago");
    });

    it('should format weeks ago (plural)', () => {
      const now = new Date("2024-01-29T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("2 weeks ago");
    });

    it('should format months ago (singular)', () => {
      const now = new Date("2024-02-15T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 month ago");
    });

    it('should format months ago (plural)', () => {
      const now = new Date("2024-06-15T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("5 months ago");
    });

    it('should format years ago (singular)', () => {
      const now = new Date("2025-01-15T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("1 year ago");
    });

    it('should format years ago (plural)', () => {
      const now = new Date("2027-01-15T12:00:00Z");
      vi.setSystemTime(now);
      const date = new Date("2024-01-15T12:00:00Z");
      expect(formatRelativeDate(date)).toBe("3 years ago");
    });
  });

  describe("shortHash", () => {
    it("should truncate hash to 7 characters", () => {
      expect(shortHash("abc123456789def")).toBe("abc1234");
    });

    it("should handle exactly 7 character hash", () => {
      expect(shortHash("abc1234")).toBe("abc1234");
    });

    it("should handle shorter than 7 characters", () => {
      expect(shortHash("abc")).toBe("abc");
    });

    it("should handle empty string", () => {
      expect(shortHash("")).toBe("");
    });

    it("should handle full SHA-1 hash", () => {
      expect(shortHash("da39a3ee5e6b4b0d3255bfef95601890afd80709")).toBe(
        "da39a3e"
      );
    });

    it("should handle full SHA-256 hash", () => {
      expect(
        shortHash(
          "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
        )
      ).toBe("e3b0c44");
    });
  });

  describe("truncate", () => {
    it("should not truncate short strings", () => {
      expect(truncate("hello", 10)).toBe("hello");
    });

    it("should not truncate strings at exact length", () => {
      expect(truncate("hello", 5)).toBe("hello");
    });

    it("should truncate long strings with ellipsis", () => {
      expect(truncate("hello world", 8)).toBe("hello...");
    });

    it("should handle very short max length", () => {
      expect(truncate("hello world", 4)).toBe("h...");
    });

    it("should handle empty string", () => {
      expect(truncate("", 10)).toBe("");
    });

    it("should handle max length of 3 (minimum for ellipsis)", () => {
      expect(truncate("hello", 3)).toBe("...");
    });
  });
});
