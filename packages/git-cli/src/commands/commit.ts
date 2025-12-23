import { Command } from "commander";
import { Repository } from "@browser-git/browser-git";
import { success, error } from "../utils/output.js";
import { parseAuthor } from "../utils/parser.js";

export const commitCommand = new Command("commit")
  .description("Record changes to the repository")
  .option("-m, --message <message>", "commit message")
  .option("-a, --all", "commit all changed files")
  .option("--author <author>", 'override author (format: "Name <email>")')
  .option("--amend", "amend previous commit")
  .option("--allow-empty", "allow empty commit")
  .action(async (options) => {
    try {
      if (!options.message && !options.amend) {
        error("Commit message required. Use -m flag.");
        process.exit(1);
      }

      const repo = await Repository.open(process.cwd());

      if (options.all) {
        await repo.add(["."]);
      }

      const commitOptions: any = {
        message: options.message,
        amend: options.amend,
        allowEmpty: options.allowEmpty,
      };

      if (options.author) {
        const author = parseAuthor(options.author);
        if (!author) {
          error('Invalid author format. Use: "Name <email>"');
          process.exit(1);
        }
        commitOptions.author = author;
      }

      const commitHash = await repo.commit(options.message, commitOptions);

      if (options.amend) {
        success(`Amended commit ${commitHash.substring(0, 7)}`);
      } else {
        success(`Created commit ${commitHash.substring(0, 7)}`);
      }
    } catch (err) {
      error(`Failed to commit: ${(err as Error).message}`);
      process.exit(1);
    }
  });
