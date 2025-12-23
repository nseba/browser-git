import { Command } from "commander";
import { Repository } from "@browser-git/browser-git";
import { error } from "../utils/output.js";
import {
  formatRelativeDate,
  shortHash,
  formatAuthor,
  truncate,
} from "../utils/parser.js";
import chalk from "chalk";

export const logCommand = new Command("log")
  .description("Show commit logs")
  .option("--oneline", "show condensed one-line format")
  .option("--graph", "show commit graph")
  .option("-n, --max-count <number>", "limit number of commits", "10")
  .option("--author <pattern>", "filter by author")
  .option("--since <date>", "show commits since date")
  .option("--until <date>", "show commits until date")
  .option("--grep <pattern>", "filter commits by message")
  .action(async (options) => {
    try {
      const repo = await Repository.open(process.cwd());
      const logOptions: any = {
        maxCount: parseInt(options.maxCount),
      };
      if (options.author) logOptions.author = options.author;
      if (options.since) logOptions.since = new Date(options.since);
      if (options.until) logOptions.until = new Date(options.until);
      if (options.grep) logOptions.grep = options.grep;

      const logs = await repo.log(logOptions);

      if (logs.length === 0) {
        console.log("No commits yet");
        return;
      }

      logs.forEach((commit: any, index: number) => {
        if (options.oneline) {
          console.log(
            chalk.yellow(shortHash(commit.hash)),
            truncate(commit.message, 60),
          );
        } else {
          if (index > 0) console.log();
          console.log(chalk.yellow(`commit ${commit.hash}`));
          console.log(`Author: ${formatAuthor(commit.author)}`);
          console.log(`Date:   ${formatRelativeDate(new Date(commit.date))}`);
          console.log();
          commit.message.split("\n").forEach((line: string) => {
            console.log(`    ${line}`);
          });
        }
      });
    } catch (err) {
      error(`Failed to show log: ${(err as Error).message}`);
      process.exit(1);
    }
  });
