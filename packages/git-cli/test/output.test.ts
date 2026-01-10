import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  setOutputOptions,
  success,
  error,
  warning,
  info,
  header,
  section,
  dim,
  highlight,
  table,
  list,
  progress,
} from "../src/utils/output.js";

describe("output utilities", () => {
  let consoleLogSpy: ReturnType<typeof vi.spyOn>;
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;
  let consoleWarnSpy: ReturnType<typeof vi.spyOn>;
  let stdoutWriteSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
    consoleWarnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    stdoutWriteSpy = vi
      .spyOn(process.stdout, "write")
      .mockImplementation(() => true);
    // Disable color for predictable test output
    setOutputOptions({ color: false });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("setOutputOptions", () => {
    it("should update color option", () => {
      setOutputOptions({ color: true });
      success("test");
      // When color is enabled, chalk adds ANSI codes
      expect(consoleLogSpy).toHaveBeenCalled();
    });
  });

  describe("success", () => {
    it("should log success message with checkmark", () => {
      success("Operation completed");
      expect(consoleLogSpy).toHaveBeenCalledWith("✓", "Operation completed");
    });

    it("should handle empty message", () => {
      success("");
      expect(consoleLogSpy).toHaveBeenCalledWith("✓", "");
    });
  });

  describe("error", () => {
    it("should log error message with X mark", () => {
      error("Operation failed");
      expect(consoleErrorSpy).toHaveBeenCalledWith("✗", "Operation failed");
    });

    it("should handle empty message", () => {
      error("");
      expect(consoleErrorSpy).toHaveBeenCalledWith("✗", "");
    });
  });

  describe("warning", () => {
    it("should log warning message with warning symbol", () => {
      warning("Proceed with caution");
      expect(consoleWarnSpy).toHaveBeenCalledWith("⚠", "Proceed with caution");
    });

    it("should handle empty message", () => {
      warning("");
      expect(consoleWarnSpy).toHaveBeenCalledWith("⚠", "");
    });
  });

  describe("info", () => {
    it("should log info message with info symbol", () => {
      info("Informational message");
      expect(consoleLogSpy).toHaveBeenCalledWith("ℹ", "Informational message");
    });

    it("should handle empty message", () => {
      info("");
      expect(consoleLogSpy).toHaveBeenCalledWith("ℹ", "");
    });
  });

  describe("header", () => {
    it("should log header message", () => {
      header("Section Header");
      expect(consoleLogSpy).toHaveBeenCalledWith("Section Header");
    });
  });

  describe("section", () => {
    it("should log section title with underline when color disabled", () => {
      section("My Section");
      expect(consoleLogSpy).toHaveBeenCalledWith("My Section");
      expect(consoleLogSpy).toHaveBeenCalledWith("==========");
    });

    it("should create underline matching title length", () => {
      section("Title");
      expect(consoleLogSpy).toHaveBeenCalledWith("Title");
      expect(consoleLogSpy).toHaveBeenCalledWith("=====");
    });
  });

  describe("dim", () => {
    it("should log dim message", () => {
      dim("Dimmed text");
      expect(consoleLogSpy).toHaveBeenCalledWith("Dimmed text");
    });
  });

  describe("highlight", () => {
    it("should log highlighted message", () => {
      highlight("Important text");
      expect(consoleLogSpy).toHaveBeenCalledWith("Important text");
    });
  });

  describe("table", () => {
    it("should handle empty rows", () => {
      table([]);
      expect(consoleLogSpy).not.toHaveBeenCalled();
    });

    it("should format single row", () => {
      table([["column1", "column2"]]);
      expect(consoleLogSpy).toHaveBeenCalledWith("column1  column2");
    });

    it("should align columns properly", () => {
      table([
        ["short", "a"],
        ["verylongtext", "b"],
      ]);
      expect(consoleLogSpy).toHaveBeenCalledWith("short         a");
      expect(consoleLogSpy).toHaveBeenCalledWith("verylongtext  b");
    });

    it("should handle multiple columns", () => {
      table([
        ["a", "b", "c"],
        ["d", "e", "f"],
      ]);
      expect(consoleLogSpy).toHaveBeenCalledTimes(2);
    });

    it("should handle empty cells", () => {
      table([
        ["a", ""],
        ["", "b"],
      ]);
      expect(consoleLogSpy).toHaveBeenCalledTimes(2);
    });
  });

  describe("list", () => {
    it("should format items with default bullet", () => {
      list(["item1", "item2"]);
      expect(consoleLogSpy).toHaveBeenCalledWith("• item1");
      expect(consoleLogSpy).toHaveBeenCalledWith("• item2");
    });

    it("should use custom prefix", () => {
      list(["item1", "item2"], "-");
      expect(consoleLogSpy).toHaveBeenCalledWith("- item1");
      expect(consoleLogSpy).toHaveBeenCalledWith("- item2");
    });

    it("should handle empty list", () => {
      list([]);
      expect(consoleLogSpy).not.toHaveBeenCalled();
    });

    it("should handle numbered prefix", () => {
      list(["first", "second"], "1.");
      expect(consoleLogSpy).toHaveBeenCalledWith("1. first");
      expect(consoleLogSpy).toHaveBeenCalledWith("1. second");
    });
  });

  describe("progress", () => {
    it("should write progress bar to stdout", () => {
      progress(50, 100);
      expect(stdoutWriteSpy).toHaveBeenCalled();
      const output = stdoutWriteSpy.mock.calls[0][0] as string;
      expect(output).toContain("50%");
      expect(output).toContain("█");
      expect(output).toContain("░");
    });

    it("should include message when provided", () => {
      progress(25, 100, "Loading...");
      const output = stdoutWriteSpy.mock.calls[0][0] as string;
      expect(output).toContain("25%");
      expect(output).toContain("Loading...");
    });

    it("should show 0% at start", () => {
      progress(0, 100);
      const output = stdoutWriteSpy.mock.calls[0][0] as string;
      expect(output).toContain("0%");
    });

    it("should show 100% at completion", () => {
      progress(100, 100);
      const output = stdoutWriteSpy.mock.calls[0][0] as string;
      expect(output).toContain("100%");
    });

    it("should add newline at 100%", () => {
      progress(100, 100);
      expect(stdoutWriteSpy).toHaveBeenCalledWith("\n");
    });

    it("should not add newline before 100%", () => {
      progress(50, 100);
      expect(stdoutWriteSpy).toHaveBeenCalledTimes(1);
    });

    it("should handle fractional percentages", () => {
      progress(33, 100);
      const output = stdoutWriteSpy.mock.calls[0][0] as string;
      expect(output).toContain("33%");
    });
  });

  describe("color mode", () => {
    it("should respect color option setting", () => {
      // Test that setOutputOptions is called and function works
      setOutputOptions({ color: true });
      success("test");
      expect(consoleLogSpy).toHaveBeenCalled();

      consoleLogSpy.mockClear();

      setOutputOptions({ color: false });
      success("test");
      expect(consoleLogSpy).toHaveBeenCalledWith("✓", "test");
    });

    it("should output plain text with color disabled", () => {
      setOutputOptions({ color: false });
      success("test");
      expect(consoleLogSpy).toHaveBeenCalledWith("✓", "test");
    });
  });
});
