import { Command } from "commander";
import { Repository } from "@browser-git/browser-git";
import { success, error } from "../utils/output.js";
import { parseGlobPatterns } from "../utils/parser.js";

export const addCommand = new Command("add")
  .description("Add file contents to the index")
  .argument("<paths...>", "files to add")
  .option("-A, --all", "add all changes")
  .option("-u, --update", "update tracked files")
  .option("-f, --force", "allow adding otherwise ignored files")
  .action(async (paths: string[], options) => {
    try {
      const repo = await Repository.open(process.cwd());

      if (options.all) {
        await repo.add(["."]);
        success("Added all changes to index");
      } else {
        const patterns = parseGlobPatterns(paths);
        await repo.add(patterns, {
          force: options.force,
          update: options.update,
        });
        success(`Added ${patterns.length} file(s) to index`);
      }
    } catch (err) {
      error(`Failed to add files: ${(err as Error).message}`);
      process.exit(1);
    }
  });
