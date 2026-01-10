/**
 * Integration tests for CLI workflows
 * These tests verify that CLI commands work correctly together in complete Git workflows
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { Command } from "commander";

// Mock the browser-git module
vi.mock("@browser-git/browser-git", () => ({
  Repository: {
    init: vi.fn(),
    open: vi.fn(),
    clone: vi.fn(),
  },
  AuthMethod: {
    Basic: "basic",
    Token: "token",
    OAuth: "oauth",
  },
}));

// Mock output utilities
vi.mock("../src/utils/output.js", () => ({
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  progress: vi.fn(),
  section: vi.fn(),
  dim: vi.fn(),
  header: vi.fn(),
  setOutputOptions: vi.fn(),
}));

import { Repository } from "@browser-git/browser-git";
import { success, error, warning, info } from "../src/utils/output.js";
import { initCommand } from "../src/commands/init.js";
import { addCommand } from "../src/commands/add.js";
import { commitCommand } from "../src/commands/commit.js";
import { statusCommand } from "../src/commands/status.js";
import { logCommand } from "../src/commands/log.js";
import { branchCommand } from "../src/commands/branch.js";
import { checkoutCommand } from "../src/commands/checkout.js";
import { mergeCommand } from "../src/commands/merge.js";
import { cloneCommand } from "../src/commands/clone.js";
import { fetchCommand } from "../src/commands/fetch.js";
import { pullCommand } from "../src/commands/pull.js";
import { pushCommand } from "../src/commands/push.js";

// Helper to reset Commander.js command options between tests
// Commander caches parsed option values on the command object
// We only reset options that were set by CLI (not defaults) to preserve default values
function resetCommandOptions(command: Command): void {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const cmd = command as any;
  const sources = cmd._optionValueSources || {};
  const values = cmd._optionValues || {};

  // Only remove values that were set by CLI, preserve defaults
  for (const key of Object.keys(sources)) {
    if (sources[key] === "cli") {
      delete values[key];
      delete sources[key];
    }
  }
}

// Helper to run a command with arguments
async function runCommand(command: Command, args: string[]): Promise<void> {
  // Reset cached options to prevent test pollution
  resetCommandOptions(command);

  const program = new Command();
  program.exitOverride();
  program.addCommand(command);
  await program.parseAsync(["node", "test", ...args]);
}

describe("CLI Integration Tests", () => {
  let mockRepo: any;
  let processExitSpy: ReturnType<typeof vi.spyOn>;
  let processCwdSpy: ReturnType<typeof vi.spyOn>;
  let consoleLogSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.clearAllMocks();

    // Create comprehensive mock repository
    mockRepo = {
      // Basic operations
      add: vi.fn().mockResolvedValue(undefined),
      commit: vi.fn().mockResolvedValue("abc123456789"),
      status: vi.fn().mockResolvedValue({
        branch: "main",
        modified: [],
        added: [],
        deleted: [],
        untracked: [],
        staged: [],
      }),

      // History operations
      log: vi.fn().mockResolvedValue([
        {
          hash: "abc123456789",
          message: "Initial commit",
          author: { name: "Test User", email: "test@example.com" },
          date: new Date("2024-01-15T12:00:00Z"),
        },
      ]),

      // Branch operations
      listBranches: vi.fn().mockResolvedValue(["main"]),
      getCurrentBranch: vi.fn().mockResolvedValue("main"),
      createBranch: vi.fn().mockResolvedValue(undefined),
      deleteBranch: vi.fn().mockResolvedValue(undefined),
      renameBranch: vi.fn().mockResolvedValue(undefined),

      // Checkout and merge
      checkout: vi.fn().mockResolvedValue(undefined),
      merge: vi.fn().mockResolvedValue({ commitHash: "merge123" }),
      mergeAbort: vi.fn().mockResolvedValue(undefined),

      // Remote operations
      fetch: vi.fn().mockResolvedValue(undefined),
      pull: vi.fn().mockResolvedValue({ alreadyUpToDate: false }),
      push: vi.fn().mockResolvedValue(undefined),
      setAuth: vi.fn().mockResolvedValue(undefined),
      listRemotes: vi.fn().mockResolvedValue([{ name: "origin" }]),

      // Diff operations
      diff: vi.fn().mockResolvedValue([]),
    };

    // Set up Repository mocks
    vi.mocked(Repository.open).mockResolvedValue(mockRepo);
    vi.mocked(Repository.init).mockResolvedValue(mockRepo);
    vi.mocked(Repository.clone).mockResolvedValue(mockRepo);

    // Mock process methods
    processExitSpy = vi.spyOn(process, "exit").mockImplementation((code) => {
      throw new Error(`process.exit(${code})`);
    });
    processCwdSpy = vi.spyOn(process, "cwd").mockReturnValue("/test/repo");
    consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("Basic Git Workflow: init → add → commit → status → log", () => {
    it("should complete a basic workflow from init to log", async () => {
      // Step 1: Initialize repository
      await runCommand(initCommand, ["init"]);
      expect(Repository.init).toHaveBeenCalledWith(".", {
        bare: undefined,
        initialBranch: "main",
        hashAlgorithm: "sha1",
      });
      expect(success).toHaveBeenCalledWith(
        "Initialized empty Git repository in ./.git/"
      );

      vi.clearAllMocks();

      // Step 2: Add files
      await runCommand(addCommand, ["add", "."]);
      expect(Repository.open).toHaveBeenCalled();
      expect(mockRepo.add).toHaveBeenCalledWith(["."], {
        force: undefined,
        update: undefined,
      });
      expect(success).toHaveBeenCalledWith("Added 1 file(s) to index");

      vi.clearAllMocks();

      // Step 3: Commit changes
      await runCommand(commitCommand, [
        "commit",
        "-m",
        "Initial commit",
      ]);
      expect(mockRepo.commit).toHaveBeenCalledWith(
        "Initial commit",
        expect.any(Object)
      );
      expect(success).toHaveBeenCalledWith(expect.stringContaining("Created commit"));

      vi.clearAllMocks();

      // Step 4: Check status
      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();

      vi.clearAllMocks();

      // Step 5: View log
      await runCommand(logCommand, ["log"]);
      expect(mockRepo.log).toHaveBeenCalled();
    });

    it("should handle workflow with modified files", async () => {
      // Setup: repo with modified files
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: ["file1.txt", "file2.txt"],
        added: [],
        deleted: [],
        untracked: ["new-file.txt"],
        staged: [],
      });

      // Check status shows modified and untracked
      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();

      vi.clearAllMocks();

      // Add specific files (not using -A which has different behavior)
      await runCommand(addCommand, ["add", "file1.txt", "file2.txt"]);
      expect(mockRepo.add).toHaveBeenCalled();

      vi.clearAllMocks();

      // Commit the changes
      await runCommand(commitCommand, [
        "commit",
        "-m",
        "Add modified files",
      ]);
      expect(mockRepo.commit).toHaveBeenCalled();
    });

    it("should support commit with --all flag", async () => {
      // The -a flag stages modified files and commits
      await runCommand(commitCommand, [
        "commit",
        "-a",
        "-m",
        "Auto-stage and commit",
      ]);

      expect(mockRepo.add).toHaveBeenCalledWith(["."]);
      expect(mockRepo.commit).toHaveBeenCalledWith(
        "Auto-stage and commit",
        expect.any(Object)
      );
    });
  });

  describe("Branch Workflow: create → checkout → modify → merge", () => {
    it("should complete a feature branch workflow", async () => {
      // Step 1: Create a new branch
      await runCommand(branchCommand, ["branch", "feature"]);
      expect(mockRepo.createBranch).toHaveBeenCalledWith("feature");
      expect(success).toHaveBeenCalledWith("Created branch feature");

      vi.clearAllMocks();

      // Step 2: Checkout the new branch
      await runCommand(checkoutCommand, ["checkout", "feature"]);
      expect(mockRepo.checkout).toHaveBeenCalledWith("feature", {
        force: undefined,
      });
      expect(success).toHaveBeenCalledWith("Switched to branch 'feature'");

      vi.clearAllMocks();

      // Step 3: Make changes and commit on feature branch
      await runCommand(commitCommand, [
        "commit",
        "-a",
        "-m",
        "Feature work",
      ]);
      expect(mockRepo.commit).toHaveBeenCalled();

      vi.clearAllMocks();

      // Step 4: Checkout main
      await runCommand(checkoutCommand, ["checkout", "main"]);
      expect(mockRepo.checkout).toHaveBeenCalledWith("main", {
        force: undefined,
      });

      vi.clearAllMocks();

      // Step 5: Merge feature into main
      mockRepo.merge.mockResolvedValue({ commitHash: "merge456" });
      await runCommand(mergeCommand, ["merge", "feature"]);
      expect(mockRepo.merge).toHaveBeenCalledWith("feature", expect.any(Object));
      expect(success).toHaveBeenCalledWith(
        expect.stringContaining("Merge completed")
      );
    });

    it("should create and checkout branch in one step with -b", async () => {
      await runCommand(checkoutCommand, ["checkout", "-b", "new-feature"]);

      expect(mockRepo.createBranch).toHaveBeenCalledWith("new-feature");
      expect(mockRepo.checkout).toHaveBeenCalledWith(
        "new-feature",
        expect.any(Object)
      );
      expect(success).toHaveBeenCalledWith(
        "Switched to a new branch 'new-feature'"
      );
    });

    it("should handle fast-forward merge", async () => {
      mockRepo.merge.mockResolvedValue({
        fastForward: true,
        commitHash: "ff123",
      });

      await runCommand(mergeCommand, ["merge", "feature"]);

      expect(success).toHaveBeenCalledWith(
        expect.stringContaining("Fast-forward merge")
      );
    });

    it("should handle merge conflicts", async () => {
      mockRepo.merge.mockResolvedValue({
        conflicts: [{ path: "file.txt" }],
      });

      try {
        await runCommand(mergeCommand, ["merge", "feature"]);
      } catch {
        // Expected - conflicts cause process.exit
      }

      expect(warning).toHaveBeenCalledWith(
        expect.stringContaining("Automatic merge failed")
      );
    });

    it("should support merge abort", async () => {
      // Note: merge command requires a branch argument, even for --abort
      // Pass a dummy branch (it will be ignored when --abort is used)
      await runCommand(mergeCommand, ["merge", "--abort", "dummy"]);

      expect(mockRepo.mergeAbort).toHaveBeenCalled();
      expect(success).toHaveBeenCalledWith("Merge aborted");
    });

    it("should delete branch after merge", async () => {
      // First merge the branch
      mockRepo.merge.mockResolvedValue({ commitHash: "merge789" });
      await runCommand(mergeCommand, ["merge", "feature"]);

      vi.clearAllMocks();

      // Then delete the merged branch
      await runCommand(branchCommand, ["branch", "-d", "feature"]);
      expect(mockRepo.deleteBranch).toHaveBeenCalledWith("feature", false);
      expect(success).toHaveBeenCalledWith("Deleted branch feature");
    });
  });

  describe("Remote Workflow: clone → fetch → pull → push", () => {
    it("should complete a remote workflow", async () => {
      // Step 1: Clone a repository
      await runCommand(cloneCommand, [
        "clone",
        "https://github.com/user/repo.git",
      ]);
      expect(Repository.clone).toHaveBeenCalledWith(
        "https://github.com/user/repo.git",
        "repo",
        expect.any(Object)
      );
      expect(success).toHaveBeenCalledWith(
        expect.stringContaining("Cloned repository")
      );

      vi.clearAllMocks();

      // Step 2: Fetch updates
      await runCommand(fetchCommand, ["fetch"]);
      expect(mockRepo.fetch).toHaveBeenCalledWith(expect.any(Object));
      expect(success).toHaveBeenCalledWith("Fetched from origin");

      vi.clearAllMocks();

      // Step 3: Pull changes
      mockRepo.pull.mockResolvedValue({ alreadyUpToDate: false });
      await runCommand(pullCommand, ["pull"]);
      expect(mockRepo.pull).toHaveBeenCalledWith(expect.any(Object));
      expect(success).toHaveBeenCalledWith("Pull completed");

      vi.clearAllMocks();

      // Step 4: Make local changes and push
      await runCommand(commitCommand, [
        "commit",
        "-a",
        "-m",
        "Local changes",
      ]);
      vi.clearAllMocks();

      await runCommand(pushCommand, ["push"]);
      expect(mockRepo.push).toHaveBeenCalledWith(expect.any(Object));
      expect(success).toHaveBeenCalledWith("Pushed to origin");
    });

    it("should clone to specific directory", async () => {
      await runCommand(cloneCommand, [
        "clone",
        "https://github.com/user/repo.git",
        "my-project",
      ]);

      expect(Repository.clone).toHaveBeenCalledWith(
        "https://github.com/user/repo.git",
        "my-project",
        expect.any(Object)
      );
    });

    it("should handle already up-to-date pull", async () => {
      mockRepo.pull.mockResolvedValue({ alreadyUpToDate: true });

      await runCommand(pullCommand, ["pull"]);

      expect(success).toHaveBeenCalledWith("Already up to date");
    });

    it("should fetch from specific remote", async () => {
      await runCommand(fetchCommand, ["fetch", "upstream"]);

      expect(mockRepo.fetch).toHaveBeenCalledWith(
        expect.objectContaining({ remote: "upstream" })
      );
    });

    it("should push with force option", async () => {
      await runCommand(pushCommand, ["push", "-f"]);

      expect(mockRepo.push).toHaveBeenCalledWith(
        expect.objectContaining({ force: true })
      );
    });

    it("should push to specific remote and refspec", async () => {
      await runCommand(pushCommand, ["push", "upstream", "feature"]);

      expect(mockRepo.push).toHaveBeenCalledWith(
        expect.objectContaining({
          remote: "upstream",
          refspec: "feature",
        })
      );
    });
  });

  describe("Combined Workflow: Feature Development Cycle", () => {
    it("should complete full feature development cycle", async () => {
      // Setup mocks for the entire workflow
      mockRepo.pull.mockResolvedValue({ alreadyUpToDate: false });
      mockRepo.merge.mockResolvedValue({ commitHash: "final123" });

      // 1. Fetch latest
      await runCommand(fetchCommand, ["fetch"]);
      expect(mockRepo.fetch).toHaveBeenCalled();

      // 2. Pull latest changes
      await runCommand(pullCommand, ["pull"]);
      expect(mockRepo.pull).toHaveBeenCalled();

      // 3. Create feature branch
      await runCommand(checkoutCommand, ["checkout", "-b", "feature/new-api"]);
      expect(mockRepo.createBranch).toHaveBeenCalledWith("feature/new-api");

      // 4. Make commits on feature branch
      await runCommand(commitCommand, [
        "commit",
        "-a",
        "-m",
        "Add new API endpoint",
      ]);
      expect(mockRepo.commit).toHaveBeenCalled();

      // 5. Checkout main
      await runCommand(checkoutCommand, ["checkout", "main"]);
      expect(mockRepo.checkout).toHaveBeenCalled();

      // 6. Merge feature branch
      await runCommand(mergeCommand, ["merge", "feature/new-api"]);
      expect(mockRepo.merge).toHaveBeenCalledWith(
        "feature/new-api",
        expect.any(Object)
      );

      // 7. Push merged changes
      await runCommand(pushCommand, ["push"]);
      expect(mockRepo.push).toHaveBeenCalled();

      // 8. Delete feature branch
      await runCommand(branchCommand, ["branch", "-d", "feature/new-api"]);
      expect(mockRepo.deleteBranch).toHaveBeenCalledWith(
        "feature/new-api",
        false
      );
    });
  });

  describe("Log and History Workflow", () => {
    it("should view log with various options", async () => {
      mockRepo.log.mockResolvedValue([
        {
          hash: "abc123",
          message: "Third commit",
          author: { name: "Dev", email: "dev@test.com" },
          date: new Date(),
        },
        {
          hash: "def456",
          message: "Second commit",
          author: { name: "Dev", email: "dev@test.com" },
          date: new Date(),
        },
        {
          hash: "ghi789",
          message: "First commit",
          author: { name: "Dev", email: "dev@test.com" },
          date: new Date(),
        },
      ]);

      // View recent commits
      await runCommand(logCommand, ["log", "-n", "3"]);
      expect(mockRepo.log).toHaveBeenCalledWith(
        expect.objectContaining({ maxCount: 3 })
      );

      vi.clearAllMocks();

      // View oneline format
      await runCommand(logCommand, ["log", "--oneline"]);
      expect(mockRepo.log).toHaveBeenCalled();
    });

    it("should view commit history after multiple commits", async () => {
      // Simulate making several commits
      const commits = ["feat: add feature", "fix: bug fix", "docs: update readme"];

      for (const msg of commits) {
        mockRepo.commit.mockResolvedValueOnce(`hash-${msg.slice(0, 4)}`);
        await runCommand(commitCommand, ["commit", "-a", "-m", msg]);
      }

      vi.clearAllMocks();

      // View the history
      mockRepo.log.mockResolvedValue(
        commits.map((msg, i) => ({
          hash: `hash${i}`,
          message: msg,
          author: { name: "Dev", email: "dev@test.com" },
          date: new Date(),
        }))
      );

      await runCommand(logCommand, ["log"]);
      expect(mockRepo.log).toHaveBeenCalled();
    });
  });

  describe("Status Throughout Workflow", () => {
    it("should show correct status at each workflow stage", async () => {
      // Initial clean state
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: [],
        added: [],
        deleted: [],
        untracked: [],
        staged: [],
      });

      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();
      vi.clearAllMocks();

      // After creating new files (untracked)
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: [],
        added: [],
        deleted: [],
        untracked: ["newfile.txt"],
        staged: [],
      });

      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();
      vi.clearAllMocks();

      // After staging (added)
      await runCommand(addCommand, ["add", "newfile.txt"]);
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: [],
        added: ["newfile.txt"],
        deleted: [],
        untracked: [],
        staged: ["newfile.txt"],
      });

      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();
      vi.clearAllMocks();

      // After commit (clean again)
      await runCommand(commitCommand, ["commit", "-m", "Add newfile"]);
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: [],
        added: [],
        deleted: [],
        untracked: [],
        staged: [],
      });

      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();
    });

    it("should show branch in status after checkout", async () => {
      // On main
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: [],
        added: [],
        deleted: [],
        untracked: [],
        staged: [],
      });

      await runCommand(statusCommand, ["status"]);
      vi.clearAllMocks();

      // Checkout feature branch
      await runCommand(checkoutCommand, ["checkout", "-b", "feature"]);
      mockRepo.getCurrentBranch.mockResolvedValue("feature");
      mockRepo.status.mockResolvedValue({
        branch: "feature",
        modified: [],
        added: [],
        deleted: [],
        untracked: [],
        staged: [],
      });

      await runCommand(statusCommand, ["status"]);
      expect(mockRepo.status).toHaveBeenCalled();
    });

    it("should support short status format", async () => {
      mockRepo.status.mockResolvedValue({
        branch: "main",
        modified: ["file.txt"],
        added: [],
        deleted: [],
        untracked: ["new.txt"],
        staged: [],
      });

      await runCommand(statusCommand, ["status", "-s"]);
      expect(mockRepo.status).toHaveBeenCalled();
    });
  });

  describe("Error Recovery Workflows", () => {
    it("should handle merge abort after conflict", async () => {
      // Abort the merge (requires a branch argument due to command design)
      await runCommand(mergeCommand, ["merge", "--abort", "dummy"]);
      expect(mockRepo.mergeAbort).toHaveBeenCalled();
      expect(success).toHaveBeenCalledWith("Merge aborted");
    });

    it("should allow force checkout when needed", async () => {
      // Checkout with force to discard local changes
      await runCommand(checkoutCommand, ["checkout", "-f", "main"]);

      expect(mockRepo.checkout).toHaveBeenCalledWith(
        "main",
        expect.objectContaining({ force: true })
      );
    });

    it("should force delete unmerged branch", async () => {
      await runCommand(branchCommand, ["branch", "-D", "feature"]);

      expect(mockRepo.deleteBranch).toHaveBeenCalledWith("feature", true);
      expect(success).toHaveBeenCalledWith("Deleted branch feature");
    });
  });
});
