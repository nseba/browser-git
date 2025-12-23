import { Command } from "commander";
import { Repository, AuthMethod } from "@browser-git/browser-git";
import { success, error, progress } from "../utils/output.js";

export const fetchCommand = new Command("fetch")
  .description("Download objects and refs from another repository")
  .argument("[remote]", "remote name", "origin")
  .argument("[refspec]", "refspec to fetch")
  .option("--all", "fetch from all remotes")
  .option("--prune", "remove remote-tracking references that no longer exist")
  .option("--depth <depth>", "deepen shallow clone")
  .option("--username <username>", "username for authentication")
  .option("--token <token>", "token for authentication")
  .action(async (remote: string, refspec: string | undefined, options) => {
    try {
      const repo = await Repository.open(process.cwd());

      const fetchOptions: any = {
        prune: options.prune,
        depth: options.depth ? parseInt(options.depth) : undefined,
      };

      if (refspec) {
        fetchOptions.refspec = refspec;
      }

      if (options.username && options.token) {
        await repo.setAuth({
          method: AuthMethod.Basic,
          username: options.username,
          password: options.token,
        });
      }

      // Show progress
      let lastProgress = 0;
      fetchOptions.onProgress = (current: number, total: number) => {
        progress(current, total, "Fetching objects");
        lastProgress = current;
      };

      if (options.all) {
        const remotes = await repo.listRemotes();
        for (const remoteObj of remotes) {
          await repo.fetch({ ...fetchOptions, remote: remoteObj.name });
        }
        if (lastProgress > 0) console.log();
        success(`Fetched from all remotes`);
      } else {
        await repo.fetch({ ...fetchOptions, remote });
        if (lastProgress > 0) console.log();
        success(`Fetched from ${remote}`);
      }
    } catch (err) {
      error(`Failed to fetch: ${(err as Error).message}`);
      process.exit(1);
    }
  });
